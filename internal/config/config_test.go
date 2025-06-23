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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdcasey/myshift-go/pkg/myshift"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `
pagerduty_token: "test-token-123"
schedule_id: "SCHED123"
my_user: "test@example.com"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths to use our temp file
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test loading
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify config values
	if config.PagerDutyToken != "test-token-123" {
		t.Errorf("Expected PagerDutyToken 'test-token-123', got '%s'", config.PagerDutyToken)
	}
	if config.ScheduleID != "SCHED123" {
		t.Errorf("Expected ScheduleID 'SCHED123', got '%s'", config.ScheduleID)
	}
	if config.MyUser != "test@example.com" {
		t.Errorf("Expected MyUser 'test@example.com', got '%s'", config.MyUser)
	}
}

func TestLoad_MinimalConfig(t *testing.T) {
	// Test with minimal valid config (only required fields)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `pagerduty_token: "test-token-123"`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.PagerDutyToken != "test-token-123" {
		t.Errorf("Expected PagerDutyToken 'test-token-123', got '%s'", config.PagerDutyToken)
	}

	// Optional fields should be empty
	if config.ScheduleID != "" {
		t.Errorf("Expected empty ScheduleID, got '%s'", config.ScheduleID)
	}
	if config.MyUser != "" {
		t.Errorf("Expected empty MyUser, got '%s'", config.MyUser)
	}
}

func TestLoad_ConfigNotFound(t *testing.T) {
	// Mock config paths to non-existent files
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{"/non/existent/path/myshift.yaml"}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when config file not found")
	}

	expectedMsg := "no configuration file found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	// Invalid YAML content
	configContent := `
pagerduty_token: "test-token"
invalid_yaml: [unclosed bracket
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	_, err = Load()
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}

	expectedMsg := "error parsing config file"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		testName string
		config   *myshift.Config
		wantErr  bool
		errMsg   string
	}{
		{
			testName: "valid config",
			config: &myshift.Config{
				PagerDutyToken: "valid-token",
				ScheduleID:     "SCHED123",
				MyUser:         "user@example.com",
			},
			wantErr: false,
		},
		{
			testName: "missing token",
			config: &myshift.Config{
				ScheduleID: "SCHED123",
				MyUser:     "user@example.com",
			},
			wantErr: true,
			errMsg:  "pagerduty_token' is required",
		},
		{
			testName: "empty token",
			config: &myshift.Config{
				PagerDutyToken: "",
				ScheduleID:     "SCHED123",
			},
			wantErr: true,
			errMsg:  "pagerduty_token' is required",
		},
		{
			testName: "minimal valid config",
			config: &myshift.Config{
				PagerDutyToken: "valid-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := validate(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		testName    string
		content     string
		expectError bool
		expected    *myshift.Config
	}{
		{
			testName: "valid config file",
			content: `
pagerduty_token: "token123"
schedule_id: "SCHED456"
my_user: "user@example.com"
`,
			expectError: false,
			expected: &myshift.Config{
				PagerDutyToken: "token123",
				ScheduleID:     "SCHED456",
				MyUser:         "user@example.com",
			},
		},
		{
			testName: "config with comments",
			content: `
# This is a comment
pagerduty_token: "token123"  # inline comment
# Another comment
schedule_id: "SCHED456"
`,
			expectError: false,
			expected: &myshift.Config{
				PagerDutyToken: "token123",
				ScheduleID:     "SCHED456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test-config.yaml")

			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			config, err := loadFromFile(configPath)

			if (err != nil) != tt.expectError {
				t.Errorf("loadFromFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if config.PagerDutyToken != tt.expected.PagerDutyToken {
					t.Errorf("Expected PagerDutyToken '%s', got '%s'",
						tt.expected.PagerDutyToken, config.PagerDutyToken)
				}
				if config.ScheduleID != tt.expected.ScheduleID {
					t.Errorf("Expected ScheduleID '%s', got '%s'",
						tt.expected.ScheduleID, config.ScheduleID)
				}
				if config.MyUser != tt.expected.MyUser {
					t.Errorf("Expected MyUser '%s', got '%s'",
						tt.expected.MyUser, config.MyUser)
				}
			}
		})
	}
}

func TestGetConfigPaths(t *testing.T) {
	// Save original environment
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Test with XDG_CONFIG_HOME set
	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	paths := GetConfigPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one config path")
	}

	// Should include XDG path
	found := false
	for _, path := range paths {
		if strings.Contains(path, "/custom/config/myshift.yaml") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected XDG_CONFIG_HOME path to be included")
	}

	// Test without XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")
	paths = GetConfigPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one config path when XDG_CONFIG_HOME is not set")
	}

	// Should include default .config path
	found = false
	for _, path := range paths {
		if strings.Contains(path, ".config/myshift.yaml") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected default .config path to be included")
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	// Create a valid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `
pagerduty_token: "test-token-123"
schedule_id: "SCHED123"
my_user: "test@example.com"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results
	if result.ConfigPath != configPath {
		t.Errorf("Expected ConfigPath '%s', got '%s'", configPath, result.ConfigPath)
	}
	if !result.Valid {
		t.Error("Expected Valid to be true")
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %v", result.Errors)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Expected no warnings, got %v", result.Warnings)
	}

	// Check required fields
	if !result.RequiredFields["pagerduty_token"] {
		t.Error("Expected pagerduty_token to be present")
	}

	// Check optional fields
	if !result.OptionalFields["schedule_id"] {
		t.Error("Expected schedule_id to be present")
	}
	if !result.OptionalFields["my_user"] {
		t.Error("Expected my_user to be present")
	}
}

func TestValidateConfig_MinimalValidConfig(t *testing.T) {
	// Create a minimal valid config file (only required fields)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `pagerduty_token: "test-token-123"`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results
	if !result.Valid {
		t.Error("Expected Valid to be true")
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %v", result.Errors)
	}
	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Check required fields
	if !result.RequiredFields["pagerduty_token"] {
		t.Error("Expected pagerduty_token to be present")
	}

	// Check optional fields
	if result.OptionalFields["schedule_id"] {
		t.Error("Expected schedule_id to be absent")
	}
	if result.OptionalFields["my_user"] {
		t.Error("Expected my_user to be absent")
	}

	// Check warnings
	expectedWarnings := []string{
		"Optional field 'schedule_id' is not set",
		"Optional field 'my_user' is not set",
	}
	for _, expectedWarning := range expectedWarnings {
		found := false
		for _, warning := range result.Warnings {
			if strings.Contains(warning, expectedWarning) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected warning containing '%s', not found in %v", expectedWarning, result.Warnings)
		}
	}
}

func TestValidateConfig_InvalidConfig(t *testing.T) {
	// Create an invalid config file (missing required field)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `
schedule_id: "SCHED123"
my_user: "test@example.com"
# Missing pagerduty_token
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results
	if result.Valid {
		t.Error("Expected Valid to be false")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be present")
	}

	// Check required fields
	if result.RequiredFields["pagerduty_token"] {
		t.Error("Expected pagerduty_token to be absent")
	}

	// Check that error mentions the missing field
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "pagerduty_token") && strings.Contains(err, "required") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about missing pagerduty_token, got %v", result.Errors)
	}
}

func TestValidateConfig_MalformedConfig(t *testing.T) {
	// Create a malformed config file (invalid YAML)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "myshift.yaml")

	configContent := `
pagerduty_token: "test-token"
invalid_yaml: [unclosed bracket
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Mock the config paths
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results
	if result.Valid {
		t.Error("Expected Valid to be false")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be present")
	}

	// Check that error mentions parsing error
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "Error loading config") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about loading config, got %v", result.Errors)
	}
}

func TestValidateConfig_NoConfigFound(t *testing.T) {
	// Mock config paths to non-existent files
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{"/non/existent/path1/myshift.yaml", "/non/existent/path2/myshift.yaml"}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results
	if result.ConfigPath != "" {
		t.Errorf("Expected empty ConfigPath, got '%s'", result.ConfigPath)
	}
	if result.Valid {
		t.Error("Expected Valid to be false")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be present")
	}

	// Check that error mentions no config found
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "No configuration file found") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about no config found, got %v", result.Errors)
	}

	// Check that all searched locations are listed
	if len(result.ConfigLocations) != 2 {
		t.Errorf("Expected 2 config locations, got %d", len(result.ConfigLocations))
	}
}

func TestValidateConfig_MultipleLocations(t *testing.T) {
	// Create multiple config files, only the first should be used
	tmpDir := t.TempDir()
	configPath1 := filepath.Join(tmpDir, "config1.yaml")
	configPath2 := filepath.Join(tmpDir, "config2.yaml")

	configContent1 := `pagerduty_token: "token1"`
	configContent2 := `pagerduty_token: "token2"`

	err := os.WriteFile(configPath1, []byte(configContent1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file 1: %v", err)
	}

	err = os.WriteFile(configPath2, []byte(configContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file 2: %v", err)
	}

	// Mock the config paths - first one should be used
	originalConfigPathsFunc := configPathsFunc
	configPathsFunc = func() []string {
		return []string{configPath1, configPath2}
	}
	defer func() { configPathsFunc = originalConfigPathsFunc }()

	// Test validation
	result, err := ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig() failed: %v", err)
	}

	// Verify results - should use first config
	if result.ConfigPath != configPath1 {
		t.Errorf("Expected ConfigPath '%s', got '%s'", configPath1, result.ConfigPath)
	}
	if !result.Valid {
		t.Error("Expected Valid to be true")
	}

	// Check that all locations are listed
	if len(result.ConfigLocations) != 2 {
		t.Errorf("Expected 2 config locations, got %d", len(result.ConfigLocations))
	}
}
