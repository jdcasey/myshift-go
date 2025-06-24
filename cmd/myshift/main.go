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
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/commands"
	"github.com/jdcasey/myshift-go/internal/config"
	"github.com/jdcasey/myshift-go/internal/pagerduty"
)

// version holds the application version and can be set at build time using ldflags:
//
//	go build -ldflags="-X main.version=v1.0.0"
var version = "dev"

// main is the entry point for the myshift CLI application.
// It parses command-line arguments and dispatches to the appropriate command handler.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--version":
		fmt.Printf("myshift-go %s\n", version)
	case "config":
		handleConfigCommand(os.Args[2:])
	case "next":
		handleNextCommand(os.Args[2:])
	case "plan":
		handlePlanCommand(os.Args[2:])
	case "override":
		handleOverrideCommand(os.Args[2:])
	case "upcoming":
		handleUpcomingCommand(os.Args[2:])
	case "repl":
		handleReplCommand(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays the application usage information to stdout.
// It shows available commands and their brief descriptions.
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

// handleConfigCommand processes the 'config' command and its subcommands.
// It handles configuration printing, validation, and basic config information display.
// The args parameter contains the command-line arguments after 'config'.
func handleConfigCommand(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	printSample := fs.Bool("print", false, "Print sample configuration")
	validate := fs.Bool("validate", false, "Validate existing configuration and show details")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift-go config [options]

Options:
  --print      Print sample configuration
  --validate   Validate existing configuration and show details

`)
	}

	_ = fs.Parse(args) // ExitOnError flag set handles errors by calling os.Exit

	if *printSample {
		config.PrintSample()
		return
	}

	if *validate {
		handleConfigValidation()
		return
	}

	// Default behavior: load and show basic config info
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration loaded successfully from one of:\n")
	for _, path := range config.GetConfigPaths() {
		fmt.Printf("  %s\n", path)
	}
	fmt.Printf("PagerDuty token: %s\n", maskToken(cfg.PagerDutyToken))
}

// handleConfigValidation performs detailed configuration validation and displays
// a comprehensive report including file locations, required/optional fields,
// errors, warnings, and next steps for fixing configuration issues.
func handleConfigValidation() {
	result, err := config.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during validation: %v\n", err)
		os.Exit(1)
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
		os.Exit(1)
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
		os.Exit(1)
	} else if len(result.Warnings) > 0 {
		fmt.Println("Configuration is valid but consider setting the optional fields above for better user experience.")
	} else {
		fmt.Println("🎉 Configuration is perfect!")
	}
}

// handleNextCommand processes the 'next' command to find and display the next
// upcoming on-call shift for a specified user. It accepts user email and days
// to look ahead as parameters, with fallback to configuration values.
func handleNextCommand(args []string) {
	fs := flag.NewFlagSet("next", flag.ExitOnError)
	userEmail := fs.String("user", "", "User email address (uses my_user from config if not provided)")
	days := fs.Int("days", 90, "Number of days to look ahead")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift-go next [options]

Options:
  --user string   User email address (uses my_user from config if not provided)
  --days int      Number of days to look ahead (default: 90)

`)
	}

	_ = fs.Parse(args) // ExitOnError flag set handles errors by calling os.Exit

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Use --user flag if provided, otherwise fall back to my_user config
	finalUserEmail := *userEmail
	if finalUserEmail == "" {
		finalUserEmail = cfg.MyUser
	}

	if finalUserEmail == "" {
		fmt.Fprintf(os.Stderr, "Error: --user is required (or set my_user in configuration)\n")
		fs.Usage()
		os.Exit(1)
	}

	client := pagerduty.NewClient(cfg.PagerDutyToken)
	nextCmd := commands.NewNextCommand(client)

	scheduleID := cfg.ScheduleID
	if scheduleID == "" {
		fmt.Fprintf(os.Stderr, "Error: schedule_id must be configured\n")
		os.Exit(1)
	}

	if err := nextCmd.Execute(scheduleID, finalUserEmail, *days); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handlePlanCommand processes the 'plan' command to display planned shifts
// for a schedule over a specified date range. It supports multiple output
// formats including text and iCalendar, with configurable time ranges.
func handlePlanCommand(args []string) {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	days := fs.Int("days", 28, "Number of days to show (default: 28)")
	startDate := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDate := fs.String("end", "", "End date (YYYY-MM-DD)")
	format := fs.String("format", "text", "Output format (text, ical)")
	fs.StringVar(format, "o", "text", "Output format (short)")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift-go plan [options]

Options:
  --days int         Number of days to show (default: 28)
  --start string     Start date (YYYY-MM-DD) 
  --end string       End date (YYYY-MM-DD)
  --format, -o string  Output format: text, ical (default: text)

`)
	}

	_ = fs.Parse(args) // ExitOnError flag set handles errors by calling os.Exit

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	client := pagerduty.NewClient(cfg.PagerDutyToken)
	planCmd := commands.NewPlanCommand(client)

	scheduleID := cfg.ScheduleID
	if scheduleID == "" {
		fmt.Fprintf(os.Stderr, "Error: schedule_id must be configured\n")
		os.Exit(1)
	}

	var start, end time.Time
	if *startDate != "" {
		if start, err = time.Parse("2006-01-02", *startDate); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing start date: %v\n", err)
			os.Exit(1)
		}
	} else {
		start = time.Now()
	}

	if *endDate != "" {
		if end, err = time.Parse("2006-01-02", *endDate); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing end date: %v\n", err)
			os.Exit(1)
		}
	} else {
		end = start.AddDate(0, 0, *days)
	}

	// Get the appropriate formatter
	formatter, err := commands.GetFormatter(*format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := planCmd.Execute(scheduleID, start, end, formatter, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleOverrideCommand processes the 'override' command to create schedule
// overrides. It requires user email, target user email, start time, and end
// time parameters to create a temporary schedule change in PagerDuty.
func handleOverrideCommand(args []string) {
	fs := flag.NewFlagSet("override", flag.ExitOnError)
	userEmail := fs.String("user", "", "User email to override with (required)")
	targetEmail := fs.String("target", "", "Target user email to override (required)")
	startTime := fs.String("start", "", "Start time (YYYY-MM-DD HH:MM) (required)")
	endTime := fs.String("end", "", "End time (YYYY-MM-DD HH:MM) (required)")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift-go override [options]

Options:
  --user string     User email to override with (required)
  --target string   Target user email to override (required)  
  --start string    Start time (YYYY-MM-DD HH:MM) (required)
  --end string      End time (YYYY-MM-DD HH:MM) (required)

`)
	}

	_ = fs.Parse(args) // ExitOnError flag set handles errors by calling os.Exit

	if *userEmail == "" || *targetEmail == "" || *startTime == "" || *endTime == "" {
		fmt.Fprintf(os.Stderr, "Error: --user, --target, --start, and --end are all required\n")
		fs.Usage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	client := pagerduty.NewClient(cfg.PagerDutyToken)
	overrideCmd := commands.NewOverrideCommand(client)

	scheduleID := cfg.ScheduleID
	if scheduleID == "" {
		fmt.Fprintf(os.Stderr, "Error: schedule_id must be configured\n")
		os.Exit(1)
	}

	start, err := time.Parse("2006-01-02 15:04", *startTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing start time: %v\n", err)
		os.Exit(1)
	}

	end, err := time.Parse("2006-01-02 15:04", *endTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing end time: %v\n", err)
		os.Exit(1)
	}

	if err := overrideCmd.Execute(scheduleID, *userEmail, *targetEmail, start, end); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleUpcomingCommand processes the 'upcoming' command to display all
// upcoming shifts for a specified user over a configurable time period.
// It supports multiple output formats and falls back to configuration values.
func handleUpcomingCommand(args []string) {
	fs := flag.NewFlagSet("upcoming", flag.ExitOnError)
	userEmail := fs.String("user", "", "User email address (uses my_user from config if not provided)")
	days := fs.Int("days", 28, "Number of days to look ahead")
	format := fs.String("format", "text", "Output format (text, ical)")
	fs.StringVar(format, "o", "text", "Output format (short)")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift-go upcoming [options]

Options:
  --user string        User email address (uses my_user from config if not provided)
  --days int           Number of days to look ahead (default: 28)
  --format, -o string  Output format: text, ical (default: text)

`)
	}

	_ = fs.Parse(args) // ExitOnError flag set handles errors by calling os.Exit

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Use --user flag if provided, otherwise fall back to my_user config
	finalUserEmail := *userEmail
	if finalUserEmail == "" {
		finalUserEmail = cfg.MyUser
	}

	if finalUserEmail == "" {
		fmt.Fprintf(os.Stderr, "Error: --user is required (or set my_user in configuration)\n")
		fs.Usage()
		os.Exit(1)
	}

	client := pagerduty.NewClient(cfg.PagerDutyToken)
	upcomingCmd := commands.NewUpcomingCommand(client)

	scheduleID := cfg.ScheduleID
	if scheduleID == "" {
		fmt.Fprintf(os.Stderr, "Error: schedule_id must be configured\n")
		os.Exit(1)
	}

	// Get the appropriate formatter
	formatter, err := commands.GetFormatter(*format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := upcomingCmd.Execute(scheduleID, finalUserEmail, *days, formatter, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleReplCommand processes the 'repl' command to start an interactive
// shell session. This allows users to run multiple commands without
// restarting the application and reloading configuration each time.
func handleReplCommand(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	client := pagerduty.NewClient(cfg.PagerDutyToken)
	replCmd := commands.NewReplCommand(client, cfg)

	scheduleID := cfg.ScheduleID
	if scheduleID == "" {
		fmt.Fprintf(os.Stderr, "Error: schedule_id must be configured\n")
		os.Exit(1)
	}

	if err := replCmd.Execute(scheduleID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// maskToken returns a masked version of the API token for safe display.
// It shows the first 4 and last 4 characters with asterisks in between,
// or all asterisks if the token is 8 characters or shorter.
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
