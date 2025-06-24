// Copyright 2025 John Casey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config handles configuration file discovery, loading, and validation for myshift-go.
//
// The configuration can be stored in multiple locations:
// - Linux: $XDG_CONFIG_HOME/myshift.yaml or ~/.config/myshift.yaml
// - macOS: ~/Library/Application Support/myshift.yaml
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jdcasey/myshift-go/internal/types"
	"gopkg.in/yaml.v3"
)

// getConfigPaths returns the list of possible configuration file paths in order of precedence.
// This function handles platform-specific configuration directory detection and supports
// XDG Base Directory Specification on Linux systems.
//
// The search order prioritizes user-specific locations and follows platform conventions:
//   - Linux: Respects XDG_CONFIG_HOME environment variable, falls back to ~/.config
//   - macOS: Uses the standard Application Support directory
//   - Other platforms: Uses the Linux behavior as a reasonable default
//
// Returns a slice of absolute file paths to check for configuration files.
func getConfigPaths() []string {
	var paths []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return paths
	}

	// Add XDG config path for Linux
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		paths = append(paths, filepath.Join(xdgConfig, "myshift.yaml"))
	} else {
		paths = append(paths, filepath.Join(homeDir, ".config", "myshift.yaml"))
	}

	// Add macOS path
	if runtime.GOOS == "darwin" {
		paths = append(paths, filepath.Join(homeDir, "Library", "Application Support", "myshift.yaml"))
	}

	return paths
}

// configPathsFunc is a variable that can be overridden for testing
var configPathsFunc = getConfigPaths

// Load attempts to load configuration from the first available config file
// in the standard platform-specific locations. It searches for YAML configuration
// files in order of precedence and loads the first one found.
//
// The configuration file should be a YAML file containing:
//   - pagerduty_token: Required API token for PagerDuty (string)
//   - my_user: Optional user ID or email for the current user (string)
//   - schedule_id: Optional default schedule ID (string)
//
// Search locations (in order):
//   - Linux: $XDG_CONFIG_HOME/myshift.yaml or ~/.config/myshift.yaml
//   - macOS: ~/Library/Application Support/myshift.yaml
//
// Returns a validated Config object or an error if no config file is found
// or if the configuration is invalid.
func Load() (*types.Config, error) {
	for _, path := range configPathsFunc() {
		if _, err := os.Stat(path); err == nil {
			return loadFromFile(path)
		}
	}

	return nil, fmt.Errorf("no configuration file found. Please create one using 'myshift config --print'")
}

// loadFromFile loads and validates configuration from a specific file path.
// It reads the YAML file, unmarshals it into a Config struct, and validates
// that all required fields are present and valid.
// The function validates the file path to prevent directory traversal attacks.
//
// Parameters:
//   - path: The file system path to the YAML configuration file
//
// Returns a validated Config object or an error if the file cannot be read,
// parsed, or if validation fails.
func loadFromFile(path string) (*types.Config, error) {
	// Validate and clean the file path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return nil, fmt.Errorf("invalid file path: %s", path)
	}

	// Additional validation: ensure the path doesn't contain directory traversal patterns
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("directory traversal not allowed in path: %s", path)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", cleanPath, err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %w", cleanPath, err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", cleanPath, err)
	}

	return &config, nil
}

// validate performs validation checks on a loaded configuration object.
// It ensures that all required fields are present and non-empty, and that
// optional fields conform to expected formats when provided.
//
// Currently validates:
//   - pagerduty_token: Must be present and non-empty
//
// Parameters:
//   - config: The Config object to validate
//
// Returns nil if validation passes, or an error describing the validation failure.
func validate(config *types.Config) error {
	if config.PagerDutyToken == "" {
		return fmt.Errorf("'pagerduty_token' is required in configuration")
	}

	return nil
}

// PrintSample outputs a well-documented sample configuration file to stdout.
// This function is used by the 'myshift config --print' command to help users
// create their initial configuration file with proper formatting and comments
// explaining each field.
//
// The sample includes:
//   - Comments explaining file placement locations
//   - All available configuration options with examples
//   - Helpful usage notes for each field
func PrintSample() {
	sample := `# MyShift Configuration
# This file should be placed in one of the following locations:
# - Linux: ~/.config/myshift.yaml
# - macOS: ~/Library/Application Support/myshift.yaml

# PagerDuty API token (required)
pagerduty_token: "your-pagerduty-token"

# Default schedule ID (optional)
# schedule_id: "your-default-schedule-id"

# Your PagerDuty user ID or email (optional)
# This will be used when no --user or --user-email is provided
# my_user: "your-email@example.com"  # or "your-user-id"
`
	fmt.Print(sample)
}

// GetConfigPaths returns the list of platform-specific configuration file paths
// in order of precedence. This function is exported for use by CLI commands
// that need to display configuration file locations to users.
//
// Returns a slice of file paths where configuration files are searched for,
// ordered from highest to lowest precedence.
func GetConfigPaths() []string {
	return getConfigPaths()
}

// ValidationResult represents the comprehensive result of configuration validation.
// It provides detailed information about configuration file discovery, field validation,
// errors, warnings, and guidance for fixing configuration issues.
type ValidationResult struct {
	ConfigPath      string          // Path to the config file that was found (empty if none found)
	Valid           bool            // Whether the configuration is valid and usable
	Errors          []string        // Critical validation errors that prevent usage
	Warnings        []string        // Non-critical warnings about missing optional fields
	RequiredFields  map[string]bool // Required fields and whether they're present
	OptionalFields  map[string]bool // Optional fields and whether they're present
	ConfigLocations []string        // All locations that were searched for config files
}

// ValidateConfig performs comprehensive validation of the configuration and returns
// detailed results including file discovery, field validation, and helpful guidance.
// This function is used by the 'myshift config --validate' command to provide
// users with actionable feedback about their configuration.
//
// The validation process:
//  1. Searches for configuration files in standard locations
//  2. Attempts to load and parse the found configuration
//  3. Validates required and optional fields
//  4. Generates helpful error messages and warnings
//  5. Provides next steps for fixing issues
//
// Returns a ValidationResult with comprehensive information, or an error if
// the validation process itself fails (distinct from configuration errors).
func ValidateConfig() (*ValidationResult, error) {
	result := &ValidationResult{
		RequiredFields:  make(map[string]bool),
		OptionalFields:  make(map[string]bool),
		ConfigLocations: configPathsFunc(),
	}

	// Try to find and load config from standard locations
	var config *types.Config
	var loadErr error

	for _, path := range result.ConfigLocations {
		if _, err := os.Stat(path); err == nil {
			result.ConfigPath = path
			config, loadErr = loadFromFile(path)
			break
		}
	}

	if result.ConfigPath == "" {
		result.Errors = append(result.Errors, "No configuration file found in any of the standard locations")
		return result, nil
	}

	if loadErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error loading config from %s: %v", result.ConfigPath, loadErr))
		return result, nil
	}

	// Perform detailed validation
	result.Valid = true

	// Check required fields
	result.RequiredFields["pagerduty_token"] = config.PagerDutyToken != ""
	if !result.RequiredFields["pagerduty_token"] {
		result.Errors = append(result.Errors, "Required field 'pagerduty_token' is missing or empty")
		result.Valid = false
	}

	// Check optional fields
	result.OptionalFields["schedule_id"] = config.ScheduleID != ""
	result.OptionalFields["my_user"] = config.MyUser != ""

	// Add warnings for missing optional fields
	if !result.OptionalFields["schedule_id"] {
		result.Warnings = append(result.Warnings, "Optional field 'schedule_id' is not set - you'll need to specify schedule for each command")
	}
	if !result.OptionalFields["my_user"] {
		result.Warnings = append(result.Warnings, "Optional field 'my_user' is not set - you'll need to specify --user for next/upcoming commands")
	}

	return result, nil
}
