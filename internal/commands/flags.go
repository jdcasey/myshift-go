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
	"flag"
	"fmt"
)

// CommonFlags holds common command-line flags used across commands
type CommonFlags struct {
	User   string
	Days   int
	Format string
	Start  string
	End    string
}

// FlagParser provides a unified way to parse command flags
type FlagParser struct {
	fs    *flag.FlagSet
	flags *CommonFlags
}

// NewFlagParser creates a new flag parser for the given command
func NewFlagParser(cmdName string) *FlagParser {
	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	flags := &CommonFlags{}

	return &FlagParser{
		fs:    fs,
		flags: flags,
	}
}

// AddUserFlag adds the --user flag
func (p *FlagParser) AddUserFlag(defaultValue, usage string) *FlagParser {
	p.fs.StringVar(&p.flags.User, "user", defaultValue, usage)
	return p
}

// AddDaysFlag adds the --days flag
func (p *FlagParser) AddDaysFlag(defaultValue int, usage string) *FlagParser {
	p.fs.IntVar(&p.flags.Days, "days", defaultValue, usage)
	return p
}

// AddFormatFlag adds the --format and -o flags
func (p *FlagParser) AddFormatFlag(defaultValue, usage string) *FlagParser {
	p.fs.StringVar(&p.flags.Format, "format", defaultValue, usage)
	p.fs.StringVar(&p.flags.Format, "o", defaultValue, usage+" (short)")
	return p
}

// AddStartFlag adds the --start flag
func (p *FlagParser) AddStartFlag(defaultValue, usage string) *FlagParser {
	p.fs.StringVar(&p.flags.Start, "start", defaultValue, usage)
	return p
}

// AddEndFlag adds the --end flag
func (p *FlagParser) AddEndFlag(defaultValue, usage string) *FlagParser {
	p.fs.StringVar(&p.flags.End, "end", defaultValue, usage)
	return p
}

// SetUsage sets the usage function for the flag set
func (p *FlagParser) SetUsage(usage func()) *FlagParser {
	p.fs.Usage = usage
	return p
}

// Parse parses the command line arguments
func (p *FlagParser) Parse(args []string) (*CommonFlags, error) {
	if err := p.fs.Parse(args); err != nil {
		// Handle help request gracefully - don't treat it as an error
		if err == flag.ErrHelp {
			return nil, nil // Return nil flags and nil error for help
		}
		return nil, err
	}
	return p.flags, nil
}

// RequiredFlags holds information about required flags
type RequiredFlags struct {
	User   bool
	Target bool
	Start  bool
	End    bool
}

// ValidateRequired validates that required flags are present
func (p *FlagParser) ValidateRequired(required RequiredFlags) error {
	if required.User && p.flags.User == "" {
		return fmt.Errorf("--user is required")
	}
	if required.Start && p.flags.Start == "" {
		return fmt.Errorf("--start is required")
	}
	if required.End && p.flags.End == "" {
		return fmt.Errorf("--end is required")
	}
	return nil
}

// OverrideFlags holds flags specific to the override command
type OverrideFlags struct {
	User   string
	Target string
	Start  string
	End    string
}

// ParseOverrideFlags parses flags for the override command
func ParseOverrideFlags(args []string) (*OverrideFlags, error) {
	fs := flag.NewFlagSet("override", flag.ContinueOnError)
	flags := &OverrideFlags{}

	fs.StringVar(&flags.User, "user", "", "User email to override with (required)")
	fs.StringVar(&flags.Target, "target", "", "Target user email to override (required)")
	fs.StringVar(&flags.Start, "start", "", "Start time (YYYY-MM-DD HH:MM) (required)")
	fs.StringVar(&flags.End, "end", "", "End time (YYYY-MM-DD HH:MM) (required)")

	fs.Usage = func() {
		fmt.Print(`Usage: myshift override [options]

Options:
  --user string     User email to override with (required)
  --target string   Target user email to override (required)  
  --start string    Start time (YYYY-MM-DD HH:MM) (required)
  --end string      End time (YYYY-MM-DD HH:MM) (required)

`)
	}

	if err := fs.Parse(args); err != nil {
		// Handle help request gracefully - don't treat it as an error
		if err == flag.ErrHelp {
			return nil, nil // Return nil flags and nil error for help
		}
		return nil, err
	}

	// Validate required flags
	if flags.User == "" || flags.Target == "" || flags.Start == "" || flags.End == "" {
		return nil, fmt.Errorf("--user, --target, --start, and --end are all required")
	}

	return flags, nil
}
