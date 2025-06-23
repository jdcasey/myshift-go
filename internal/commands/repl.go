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

// Package commands provides CLI command implementations for myshift-go.
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/pagerduty"
	"github.com/jdcasey/myshift-go/pkg/myshift"
)

// ReplCommand handles the "repl" command functionality.
type ReplCommand struct {
	client pagerduty.PagerDutyClient
	config *myshift.Config
}

// NewReplCommand creates a new ReplCommand instance.
func NewReplCommand(client pagerduty.PagerDutyClient, config *myshift.Config) *ReplCommand {
	return &ReplCommand{
		client: client,
		config: config,
	}
}

// Execute runs the REPL (Read-Eval-Print Loop) for interactive commands.
func (r *ReplCommand) Execute(scheduleID string) error {
	fmt.Println("Welcome to MyShift REPL. Type 'help' or '?' to list commands.")
	fmt.Println("Type 'quit' or 'exit' to quit, or use Ctrl+C.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("(myshift) ")

		if !scanner.Scan() {
			// EOF or error
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		command := args[0]
		commandArgs := args[1:]

		switch command {
		case "help", "?":
			r.printHelp()
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return nil
		case "next":
			r.handleNext(scheduleID, commandArgs)
		case "plan":
			r.handlePlan(scheduleID, commandArgs)
		case "upcoming":
			r.handleUpcoming(scheduleID, commandArgs)
		case "override":
			r.handleOverride(scheduleID, commandArgs)
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

func (r *ReplCommand) printHelp() {
	myUserInfo := ""
	if r.config != nil && r.config.MyUser != "" {
		myUserInfo = fmt.Sprintf(" (defaults to %s)", r.config.MyUser)
	}

	fmt.Printf(`Available commands:

  next [email]                    Show the next on-call shift for a user%s
  plan [days]                     Show planned shifts (default: 7 days)
  upcoming [email] [days]         Show upcoming shifts for a user%s (default: 28 days)
  override <user> <target> <start> <end>  Create an override
  help, ?                         Show this help message
  quit, exit                      Exit the REPL

Examples:
  next
  next user@example.com
  plan 14
  upcoming 7
  upcoming user@example.com 7
  override user@example.com target@example.com "2024-03-20 09:00" "2024-03-20 17:00"

`, myUserInfo, myUserInfo)
}

func (r *ReplCommand) handleNext(scheduleID string, args []string) {
	var email string
	days := 90

	if len(args) > 0 {
		email = args[0]
		if len(args) > 1 {
			if d, err := strconv.Atoi(args[1]); err == nil {
				days = d
			}
		}
	}

	// Use my_user from config if no email provided
	if email == "" && r.config != nil {
		email = r.config.MyUser
	}

	if email == "" {
		fmt.Println("Usage: next [email] - email is required or set my_user in configuration")
		return
	}

	nextCmd := NewNextCommand(r.client)
	if err := nextCmd.Execute(scheduleID, email, days); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (r *ReplCommand) handlePlan(scheduleID string, args []string) {
	days := 7
	if len(args) > 0 {
		if d, err := strconv.Atoi(args[0]); err == nil {
			days = d
		} else {
			fmt.Printf("Invalid days value: %s\n", args[0])
			return
		}
	}

	start := time.Now()
	end := start.AddDate(0, 0, days)

	planCmd := NewPlanCommand(r.client)
	formatter := NewTextFormatter()
	if err := planCmd.Execute(scheduleID, start, end, formatter, os.Stdout); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (r *ReplCommand) handleUpcoming(scheduleID string, args []string) {
	var email string
	days := 28

	if len(args) > 0 {
		email = args[0]
		if len(args) > 1 {
			if d, err := strconv.Atoi(args[1]); err == nil {
				days = d
			}
		}
	}

	// Use my_user from config if no email provided
	if email == "" && r.config != nil {
		email = r.config.MyUser
	}

	if email == "" {
		fmt.Println("Usage: upcoming [email] [days] - email is required or set my_user in configuration")
		return
	}

	upcomingCmd := NewUpcomingCommand(r.client)
	formatter := NewTextFormatter()
	if err := upcomingCmd.Execute(scheduleID, email, days, formatter, os.Stdout); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (r *ReplCommand) handleOverride(scheduleID string, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: override <user-email> <target-email> <start> <end>")
		fmt.Println("Example: override user@example.com target@example.com \"2024-03-20 09:00\" \"2024-03-20 17:00\"")
		return
	}

	userEmail := args[0]
	targetEmail := args[1]
	startStr := args[2]
	endStr := args[3]

	start, err := time.Parse("2006-01-02 15:04", startStr)
	if err != nil {
		fmt.Printf("Error parsing start time '%s': %v\n", startStr, err)
		fmt.Println("Use format: YYYY-MM-DD HH:MM")
		return
	}

	end, err := time.Parse("2006-01-02 15:04", endStr)
	if err != nil {
		fmt.Printf("Error parsing end time '%s': %v\n", endStr, err)
		fmt.Println("Use format: YYYY-MM-DD HH:MM")
		return
	}

	overrideCmd := NewOverrideCommand(r.client)
	if err := overrideCmd.Execute(scheduleID, userEmail, targetEmail, start, end); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
