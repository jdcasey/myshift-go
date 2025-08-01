#!/bin/bash

# Copyright 2025 John Casey
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# pr-preflight - Run PR checks locally before pushing
# This script runs the same checks as .github/workflows/pr-checks.yml

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly TEMP_DIR=$(mktemp -d)
readonly LOG_FILE="${TEMP_DIR}/pr-preflight.log"

# Cleanup function
cleanup() {
    rm -rf "${TEMP_DIR}"
}
trap cleanup EXIT

# Logging functions
log() {
    echo -e "${1}" | tee -a "${LOG_FILE}"
}

log_info() {
    log "${BLUE}ℹ${NC} ${1}"
}

log_success() {
    log "${GREEN}✓${NC} ${1}"
}

log_warning() {
    log "${YELLOW}⚠${NC} ${1}"
}

log_error() {
    log "${RED}✗${NC} ${1}"
}

log_step() {
    log "\n${BLUE}=== ${1} ===${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking Prerequisites"
    
    local missing_tools=()
    
    if ! command_exists go; then
        missing_tools+=("go")
    fi
    
    if ! command_exists podman; then
        missing_tools+=("podman")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and try again."
        exit 1
    fi
    
    log_success "All prerequisites available"
    
    # Show Go version
    local go_version=$(go version)
    log_info "Using: ${go_version}"
}

# Install tools if not present
install_tools() {
    log_step "Installing/Updating Tools"
    
    # Install gosec if not present
    if ! command_exists gosec && [ ! -f "$(go env GOPATH)/bin/gosec" ]; then
        log_info "Installing gosec..."
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    else
        log_info "gosec already installed"
    fi
    
    # Install nancy if not present
    if ! command_exists nancy && [ ! -f "$(go env GOPATH)/bin/nancy" ]; then
        log_info "Installing nancy..."
        go install github.com/sonatype-nexus-community/nancy@latest
    else
        log_info "nancy already installed"
    fi
    
    log_success "All tools ready"
}

# Initialize Go modules
init_go_modules() {
    log_step "Initializing Go Modules"
    
    log_info "Downloading Go modules..."
    if go mod download; then
        log_success "Go modules downloaded"
    else
        log_error "Failed to download Go modules"
        return 1
    fi
    
    log_info "Verifying Go modules..."
    if go mod verify; then
        log_success "Go modules verified"
    else
        log_error "Go module verification failed"
        return 1
    fi
}

# Run tests
run_tests() {
    log_step "Running Tests"
    
    log_info "Running tests with race detection and coverage..."
    
    if go test -v -race -coverprofile=coverage.out ./...; then
        log_success "All tests passed"
        
        # Show coverage summary if coverage file exists
        if [ -f "coverage.out" ]; then
            local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            log_info "Test coverage: ${coverage}"
        fi
    else
        log_error "Tests failed - see output above for details"
        return 1
    fi
}

# Build and test binary
run_build() {
    log_step "Building and Testing Binary"
    
    log_info "Building binary..."
    if go build -v ./cmd/myshift; then
        log_success "Binary built successfully"
    else
        log_error "Build failed"
        return 1
    fi
    
    log_info "Testing binary functionality..."
    
    # Test version command
    if ./myshift --version >/dev/null 2>&1; then
        log_success "Version command works"
    else
        log_warning "Version command failed"
    fi
    
    # Test config command
    if ./myshift config --print >/dev/null 2>&1; then
        log_success "Config command works"
    else
        log_warning "Config print command failed (may be expected if no config)"
    fi
    
    # Clean up binary
    rm -f ./myshift
}

# Run security analysis
run_security() {
    log_step "Running Security Analysis (gosec)"
    
    local gosec_cmd="$(go env GOPATH)/bin/gosec"
    if command_exists gosec; then
        gosec_cmd="gosec"
    fi
    
    log_info "Scanning for security vulnerabilities..."
    
    # Run gosec with text output for human readability
    if "${gosec_cmd}" -fmt text -stdout ./... 2>&1; then
        log_success "No security issues found"
    else
        local exit_code=$?
        if [ $exit_code -eq 1 ]; then
            log_error "Security vulnerabilities found - see output above"
            log_info "Review and fix the issues, or use #nosec comments for false positives"
            return 1
        else
            log_error "Security scan failed with exit code ${exit_code}"
            return 1
        fi
    fi
}

# Run dependency vulnerability scan
run_dependency_scan() {
    log_step "Running Dependency Vulnerability Scan (nancy)"
    
    local nancy_cmd="$(go env GOPATH)/bin/nancy"
    if command_exists nancy; then
        nancy_cmd="nancy"
    fi
    
    log_info "Scanning dependencies for known vulnerabilities..."
    
    if go list -json -deps ./... | "${nancy_cmd}" sleuth --loud; then
        log_success "No dependency vulnerabilities found"
    else
        log_error "Dependency vulnerabilities found - see output above"
        log_info "Update vulnerable dependencies or assess if they affect your usage"
        return 1
    fi
}

# Build container
run_container_build() {
    log_step "Building Container"
    
    if [ ! -f "Containerfile" ]; then
        log_warning "No Containerfile found, skipping container build"
        return 0
    fi
    
    local containerfile="Containerfile"
    
    log_info "Building container image..."
    
    if podman build -f "${containerfile}" -t myshift-go:pr-check .; then
        log_success "Container built successfully"
        
        # Clean up the test image
        podman rmi myshift-go:pr-check >/dev/null 2>&1 || true
    else
        log_error "Container build failed"
        return 1
    fi
}

# Show summary
show_summary() {
    log_step "Summary"
    
    if [ -f "${LOG_FILE}" ]; then
        local passed=$(grep -c "✓" "${LOG_FILE}" || echo "0")
        local failed=$(grep -c "✗" "${LOG_FILE}" || echo "0")
        local warnings=$(grep -c "⚠" "${LOG_FILE}" || echo "0")
        
        log_info "Checks completed:"
        log_info "  • Passed: ${passed}"
        if [ "${warnings}" -gt 0 ]; then
            log_info "  • Warnings: ${warnings}"
        fi
        if [ "${failed}" -gt 0 ]; then
            log_info "  • Failed: ${failed}"
        fi
    fi
    
    if [ "${FAILED_STEPS:-0}" -eq 0 ]; then
        log_success "All checks passed! Your PR should pass CI/CD pipeline."
    else
        log_error "Some checks failed. Please fix the issues above before creating your PR."
        log_info "Log file available at: ${LOG_FILE}"
    fi
}

# Main execution
main() {
    local start_time=$(date +%s)
    
    log_step "PR Preflight Checks Starting"
    log_info "Running pre-PR validation checks..."
    log_info "Log file: ${LOG_FILE}"
    
    # Change to project root
    cd "${PROJECT_ROOT}"
    
    local failed_steps=0
    
    # Run all checks
    check_prerequisites || ((failed_steps++))
    install_tools || ((failed_steps++))
    init_go_modules || ((failed_steps++))
    run_tests || ((failed_steps++))
    run_build || ((failed_steps++))
    run_security || ((failed_steps++))
    run_dependency_scan || ((failed_steps++))
    run_container_build || ((failed_steps++))
    
    # Export for summary
    export FAILED_STEPS=${failed_steps}
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_step "Completed in ${duration}s"
    show_summary
    
    # Exit with error if any checks failed
    if [ ${failed_steps} -gt 0 ]; then
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "pr-preflight - Run PR checks locally"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "This script runs the same checks as .github/workflows/pr-checks.yml"
        echo "to help you catch issues before creating a pull request."
        echo ""
        echo "Checks performed:"
        echo "  • Go module initialization and verification"
        echo "  • Tests with race detection"
        echo "  • Binary build and basic functionality"
        echo "  • Security analysis (gosec)"
        echo "  • Dependency vulnerability scan (nancy)"
        echo "  • Container build"
        echo ""
        echo "Options:"
        echo "  -h, --help    Show this help message"
        echo ""
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac 
