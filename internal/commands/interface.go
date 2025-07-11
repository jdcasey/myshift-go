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
	"io"

	"github.com/jdcasey/myshift-go/internal/pagerduty"
	"github.com/jdcasey/myshift-go/internal/types"
)

// Command represents a CLI command that can be executed
type Command interface {
	Execute(args []string) error
	Usage() string
}

// CommandContext holds the shared context for all commands
type CommandContext struct {
	Client pagerduty.PagerDutyClient
	Config *types.Config
	Writer io.Writer
}

// NewCommandContext creates a new command context
func NewCommandContext(client pagerduty.PagerDutyClient, config *types.Config, writer io.Writer) *CommandContext {
	return &CommandContext{
		Client: client,
		Config: config,
		Writer: writer,
	}
}
