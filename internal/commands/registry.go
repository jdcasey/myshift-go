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

package commands

import (
	"fmt"
)

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry creates a new command registry with all commands
func NewCommandRegistry(ctx *CommandContext) *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}

	// Register all commands
	registry.commands["next"] = NewNextCommand(ctx)
	registry.commands["plan"] = NewPlanCommand(ctx)
	registry.commands["upcoming"] = NewUpcomingCommand(ctx)
	registry.commands["override"] = NewOverrideCommand(ctx)
	// Note: REPL is not included in the registry to avoid circular dependency

	return registry
}

// Execute executes a command by name
func (r *CommandRegistry) Execute(cmdName string, args []string) error {
	cmd, exists := r.commands[cmdName]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmdName)
	}
	return cmd.Execute(args)
}

// GetCommand returns a command by name
func (r *CommandRegistry) GetCommand(cmdName string) (Command, bool) {
	cmd, exists := r.commands[cmdName]
	return cmd, exists
}

// ListCommands returns all available command names
func (r *CommandRegistry) ListCommands() []string {
	var commands []string
	for name := range r.commands {
		commands = append(commands, name)
	}
	return commands
}

// GetUsage returns usage information for a command
func (r *CommandRegistry) GetUsage(cmdName string) string {
	cmd, exists := r.commands[cmdName]
	if !exists {
		return fmt.Sprintf("Unknown command: %s", cmdName)
	}
	return cmd.Usage()
}
