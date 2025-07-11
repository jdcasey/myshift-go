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

// Command myshift is a CLI tool for managing PagerDuty on-call schedules.
//
// myshift provides commands to view upcoming shifts, plan schedules, create
// overrides, and interact with PagerDuty schedules through an easy-to-use
// command-line interface or interactive REPL.
//
// Usage:
//
//	myshift <command> [options]
//
// Available commands:
//   - next: Show the next upcoming on-call shift for a user
//   - plan: Display planned shifts for a schedule over a date range
//   - override: Create schedule overrides for specific time periods
//   - upcoming: Show all upcoming shifts for a user
//   - repl: Start an interactive shell for running multiple commands
//   - config: Manage application configuration
//
// Configuration is loaded from YAML files in standard locations:
//   - Linux: ~/.config/myshift.yaml
//   - macOS: ~/Library/Application Support/myshift.yaml
//
// See 'myshift config --print' for sample configuration.
package main

import (
	"fmt"
	"os"

	"github.com/jdcasey/myshift-go/internal/commands"
	"github.com/jdcasey/myshift-go/internal/config"
	"github.com/jdcasey/myshift-go/internal/pagerduty"
)

// version holds the application version and can be set at build time using ldflags:
//
//	go build -ldflags="-X main.version=v1.0.0"
var version = "dev"

// main is the entry point for the myshift CLI application.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run contains the main application logic and returns errors instead of calling os.Exit
func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("no command provided")
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Handle special commands that don't require config
	switch command {
	case "--version":
		fmt.Printf("myshift-go %s\n", version)
		return nil
	case "config":
		return handleConfigCommand(args)
	case "--help", "-h", "help":
		printUsage()
		return nil
	}

	// Load configuration for all other commands
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Create command context
	ctx := commands.NewCommandContext(
		pagerduty.NewClient(cfg.PagerDutyToken),
		cfg,
		os.Stdout,
	)

	// Handle REPL separately to avoid circular dependency
	if command == "repl" {
		replCmd := commands.NewReplCommand(ctx)
		return replCmd.Execute(args)
	}

	// Create command registry
	registry := commands.NewCommandRegistry(ctx)

	// Execute the command
	return registry.Execute(command, args)
}

// printUsage displays the application usage information
func printUsage() {
	fmt.Print(`myshift-go - PagerDuty on-call schedule management tool

Usage:
  myshift-go <command> [options]

Commands:
  next      Show next shift for a user
  plan      Show planned shifts for a schedule
  override  Create schedule overrides
  upcoming  Show upcoming shifts for a user
  repl      Start interactive REPL
  config    Manage configuration
  --version Show version

Use 'myshift-go <command> --help' for more information about a command.
`)
}

// handleConfigCommand processes the 'config' command and its subcommands
func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		// Default behavior: load and show basic config info
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		fmt.Printf("Configuration loaded successfully from one of:\n")
		for _, path := range config.GetConfigPaths() {
			fmt.Printf("  %s\n", path)
		}
		fmt.Printf("PagerDuty token: %s\n", maskToken(cfg.PagerDutyToken))
		return nil
	}

	switch args[0] {
	case "--print":
		config.PrintSample()
		return nil
	case "--validate":
		return handleConfigValidation()
	default:
		return fmt.Errorf("unknown config option: %s", args[0])
	}
}

// handleConfigValidation performs detailed configuration validation
func handleConfigValidation() error {
	result, err := config.ValidateConfig()
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	fmt.Println("Configuration Validation Report")
	fmt.Println("==============================")
	fmt.Println()

	// Show searched locations
	fmt.Println("Searched locations:")
	for _, path := range result.ConfigLocations {
		if path == result.ConfigPath {
			fmt.Printf("  ✓ %s (FOUND)\n", path)
		} else {
			fmt.Printf("  ✗ %s\n", path)
		}
	}
	fmt.Println()

	// Show overall status
	if result.ConfigPath == "" {
		fmt.Println("Status: ❌ NO CONFIGURATION FOUND")
		fmt.Println()
		fmt.Println("Please create a configuration file using:")
		fmt.Println("  myshift config --print > ~/.config/myshift.yaml")
		return fmt.Errorf("no configuration found")
	}

	fmt.Printf("Configuration file: %s\n", result.ConfigPath)
	if result.Valid {
		fmt.Println("Status: ✅ VALID")
	} else {
		fmt.Println("Status: ❌ INVALID")
	}
	fmt.Println()

	// Show required fields
	fmt.Println("Required fields:")
	for field, present := range result.RequiredFields {
		if present {
			fmt.Printf("  ✓ %s: present\n", field)
		} else {
			fmt.Printf("  ❌ %s: MISSING\n", field)
		}
	}
	fmt.Println()

	// Show optional fields
	fmt.Println("Optional fields:")
	for field, present := range result.OptionalFields {
		if present {
			fmt.Printf("  ✓ %s: present\n", field)
		} else {
			fmt.Printf("  - %s: not set\n", field)
		}
	}
	fmt.Println()

	// Show errors
	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  ❌ %s\n", err)
		}
		fmt.Println()
	}

	// Show warnings
	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  ⚠️  %s\n", warning)
		}
		fmt.Println()
	}

	// Show next steps if needed
	if !result.Valid {
		fmt.Println("Next steps:")
		fmt.Println("  1. Edit your configuration file:")
		fmt.Printf("     %s\n", result.ConfigPath)
		fmt.Println("  2. Add the missing required fields")
		fmt.Println("  3. Run 'myshift config --validate' again to verify")
		fmt.Println()
		fmt.Println("For a sample configuration:")
		fmt.Println("  myshift config --print")
		return fmt.Errorf("configuration is invalid")
	} else if len(result.Warnings) > 0 {
		fmt.Println("Configuration is valid but consider setting the optional fields above for better user experience.")
	} else {
		fmt.Println("🎉 Configuration is perfect!")
	}

	return nil
}

// maskToken masks a PagerDuty token for display
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
