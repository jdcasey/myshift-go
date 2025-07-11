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
	"strings"
)

// ReplCommand handles the "repl" command functionality.
type ReplCommand struct {
	*BaseCommand
}

// NewReplCommand creates a new ReplCommand instance.
func NewReplCommand(ctx *CommandContext) *ReplCommand {
	return &ReplCommand{
		BaseCommand: NewBaseCommand(ctx.Client, ctx.Config, ctx.Writer),
	}
}

// Execute runs the REPL (Read-Eval-Print Loop) for interactive commands.
func (r *ReplCommand) Execute(args []string) error {
	fmt.Fprintln(r.writer, "Welcome to MyShift REPL. Type 'help' or '?' to list commands.")
	fmt.Fprintln(r.writer, "Type 'quit' or 'exit' to quit, or use Ctrl+C.")

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
			fmt.Fprintln(r.writer, "Goodbye!")
			return nil
		case "next", "plan", "upcoming", "override":
			r.handleCommand(command, commandArgs)
		default:
			fmt.Fprintf(r.writer, "Unknown command: %s. Type 'help' for available commands.\n", command)
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

	fmt.Fprintf(r.writer, `Available commands:

  next [--user email] [--days N]  Show the next on-call shift for a user%s
  plan [--days N]                 Show planned shifts (default: 28 days)
  upcoming [--user email] [--days N]  Show upcoming shifts for a user%s (default: 28 days)
  override --user U --target T --start S --end E  Create an override
  help, ?                         Show this help message
  quit, exit                      Exit the REPL

Examples:
  next
  next --user user@example.com
  plan --days 14
  upcoming --days 7
  upcoming --user user@example.com --days 7
  override --user user@example.com --target target@example.com --start "2024-03-20 09:00" --end "2024-03-20 17:00"

`, myUserInfo, myUserInfo)
}

func (r *ReplCommand) handleCommand(command string, args []string) {
	// Create a registry for executing commands
	ctx := NewCommandContext(r.client, r.config, r.writer)
	registry := NewCommandRegistry(ctx)

	if err := registry.Execute(command, args); err != nil {
		fmt.Fprintf(r.writer, "Error: %v\n", err)
	}
}

// Usage returns the usage information for the repl command
func (r *ReplCommand) Usage() string {
	return `Usage: myshift repl

Start an interactive REPL (Read-Eval-Print Loop) for running multiple commands.
`
}
