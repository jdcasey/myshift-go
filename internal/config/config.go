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

	"github.com/jdcasey/myshift-go/pkg/myshift"
	"gopkg.in/yaml.v3"
)

// getConfigPaths returns the list of possible configuration file paths in order of precedence.
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

// Load loads configuration from the first available config file.
//
// The configuration file should be a YAML file containing:
// - pagerduty_token: Required API token for PagerDuty
// - my_user: Optional user ID or email for the current user
// - schedule_id: Optional default schedule ID
func Load() (*myshift.Config, error) {
	for _, path := range configPathsFunc() {
		if _, err := os.Stat(path); err == nil {
			return loadFromFile(path)
		}
	}

	return nil, fmt.Errorf("no configuration file found. Please create one using 'myshift config --print'")
}

// loadFromFile loads configuration from a specific file path.
func loadFromFile(path string) (*myshift.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	var config myshift.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %w", path, err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", path, err)
	}

	return &config, nil
}

// validate validates the configuration parameters.
func validate(config *myshift.Config) error {
	if config.PagerDutyToken == "" {
		return fmt.Errorf("'pagerduty_token' is required in configuration")
	}

	return nil
}

// PrintSample prints a sample configuration file to stdout.
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

// GetConfigPaths returns the list of configuration file paths (exported for CLI usage).
func GetConfigPaths() []string {
	return getConfigPaths()
}

// ValidationResult represents the result of configuration validation.
type ValidationResult struct {
	ConfigPath      string          // Path to the config file that was found
	Valid           bool            // Whether the configuration is valid
	Errors          []string        // Validation errors
	Warnings        []string        // Validation warnings
	RequiredFields  map[string]bool // Required fields and whether they're present
	OptionalFields  map[string]bool // Optional fields and whether they're present
	ConfigLocations []string        // All locations that were checked
}

// ValidateConfig performs detailed validation of the configuration and returns comprehensive results.
func ValidateConfig() (*ValidationResult, error) {
	result := &ValidationResult{
		RequiredFields:  make(map[string]bool),
		OptionalFields:  make(map[string]bool),
		ConfigLocations: configPathsFunc(),
	}

	// Try to find and load config from standard locations
	var config *myshift.Config
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
