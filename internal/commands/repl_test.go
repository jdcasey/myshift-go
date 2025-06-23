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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jdcasey/myshift-go/pkg/myshift"
)

// Since the REPL reads from stdin, we need to simulate user input
func TestReplCommand_Execute_BasicCommands(t *testing.T) {
	tests := []struct {
		testName       string
		input          string
		expectedOutput []string
		shouldExit     bool
	}{
		{
			testName:       "help command",
			input:          "help\nexit\n",
			expectedOutput: []string{"next [email]", "plan [days]", "upcoming [email]", "override"},
			shouldExit:     true,
		},
		{
			testName:       "question mark help",
			input:          "?\nexit\n",
			expectedOutput: []string{"next [email]", "plan [days]", "upcoming [email]", "override"},
			shouldExit:     true,
		},
		{
			testName:       "unknown command",
			input:          "invalid\nexit\n",
			expectedOutput: []string{"Unknown command: invalid"},
			shouldExit:     true,
		},
		{
			testName:       "empty lines ignored",
			input:          "\n\n  \nexit\n",
			expectedOutput: []string{"(myshift) "},
			shouldExit:     true,
		},
		{
			testName:       "exit command",
			input:          "exit\n",
			expectedOutput: []string{"Goodbye!"},
			shouldExit:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fixture := NewTestFixture()
			config := &myshift.Config{
				PagerDutyToken: "test-token",
				ScheduleID:     "SCHED123",
				MyUser:         "test@example.com",
			}

			// Create pipes for stdin and stdout
			stdinReader, stdinWriter, _ := os.Pipe()
			stdoutReader, stdoutWriter, _ := os.Pipe()

			// Save original stdin/stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout

			// Replace stdin/stdout
			os.Stdin = stdinReader
			os.Stdout = stdoutWriter

			replCmd := NewReplCommand(fixture.MockClient, config)

			// Channel to signal completion
			done := make(chan error, 1)

			// Start REPL in goroutine
			go func() {
				err := replCmd.Execute("SCHED123")
				done <- err
			}()

			// Write input and close stdin
			go func() {
				defer stdinWriter.Close()
				stdinWriter.WriteString(tt.input)
			}()

			// Wait for REPL to complete or timeout
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("REPL execution failed: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Error("REPL test timed out")
			}

			// Close stdout and restore
			stdoutWriter.Close()
			os.Stdin = oldStdin
			os.Stdout = oldStdout

			// Read captured output
			output, _ := io.ReadAll(stdoutReader)
			stdoutReader.Close()
			outputStr := string(output)

			// Verify expected outputs are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestReplCommand_Execute_NextCommand(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Setup mock for next command
	fixture.MockClient.AddUser("USER123", "Test User", "test@example.com")
	fixture.MockClient.AddOnCall("USER123", "Test User", "test@example.com",
		fixture.Now.Add(24*time.Hour),
		fixture.Now.Add(48*time.Hour))

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("next test@example.com\nexit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify next command output
	if !strings.Contains(outputStr, "Next shift:") {
		t.Errorf("Expected next command output, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_PlanCommand(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Setup mock for plan command
	fixture.MockClient.AddUser("USER123", "Test User", "test@example.com")
	fixture.MockClient.AddOnCall("USER123", "Test User", "test@example.com",
		fixture.Now.Add(2*time.Hour),
		fixture.Now.Add(26*time.Hour))

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("plan 1\nexit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify plan command output - should contain user info
	if !strings.Contains(outputStr, "Test User") {
		t.Errorf("Expected plan command output with user info, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_UpcomingCommand(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Setup mock for upcoming command
	fixture.MockClient.AddUser("USER123", "Test User", "test@example.com")
	fixture.MockClient.AddOnCall("USER123", "Test User", "test@example.com",
		fixture.Now.Add(24*time.Hour),
		fixture.Now.Add(48*time.Hour))

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("upcoming test@example.com 1\nexit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify upcoming command output - should contain shifts data and user name
	if !strings.Contains(outputStr, "Shifts for the next") || !strings.Contains(outputStr, "Test User") {
		t.Errorf("Expected upcoming command output with shifts and user info, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_OverrideCommand(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Setup mock for override command
	fixture.MockClient.AddUser("USER123", "Test User", "test@example.com")
	fixture.MockClient.AddUser("TARGET123", "Target User", "target@example.com")

	// Add a target shift to override
	fixture.MockClient.AddOnCall("TARGET123", "Target User", "target@example.com",
		fixture.Now.Add(2*time.Hour),
		fixture.Now.Add(10*time.Hour))

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	startTime := fixture.Now.Add(2 * time.Hour).Format("2006-01-02 15:04")
	endTime := fixture.Now.Add(6 * time.Hour).Format("2006-01-02 15:04")
	input := fmt.Sprintf("override test@example.com target@example.com \"%s\" \"%s\"\nexit\n", startTime, endTime)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString(input)
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Should not contain error
	if strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected no error in override command, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_CommandValidation(t *testing.T) {
	tests := []struct {
		testName       string
		input          string
		expectedOutput string
	}{
		{
			testName:       "next command missing email",
			input:          "next\nexit\n",
			expectedOutput: "email is required or set my_user in configuration",
		},
		{
			testName:       "upcoming command missing email",
			input:          "upcoming\nexit\n",
			expectedOutput: "email is required or set my_user in configuration",
		},
		{
			testName:       "override command insufficient args",
			input:          "override user@example.com\nexit\n",
			expectedOutput: "Usage: override <user-email> <target-email> <start> <end>",
		},
		{
			testName:       "plan command with invalid days",
			input:          "plan invalid\nexit\n",
			expectedOutput: "Invalid days value: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fixture := NewTestFixture()
			config := &myshift.Config{
				PagerDutyToken: "test-token",
				ScheduleID:     "SCHED123",
				// Note: no MyUser set to test validation
			}

			// Create pipes for stdin and stdout
			stdinReader, stdinWriter, _ := os.Pipe()
			stdoutReader, stdoutWriter, _ := os.Pipe()

			// Save original stdin/stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout

			// Replace stdin/stdout
			os.Stdin = stdinReader
			os.Stdout = stdoutWriter

			replCmd := NewReplCommand(fixture.MockClient, config)

			// Channel to signal completion
			done := make(chan error, 1)

			// Start REPL in goroutine
			go func() {
				err := replCmd.Execute("SCHED123")
				done <- err
			}()

			// Write input and close stdin
			go func() {
				defer stdinWriter.Close()
				stdinWriter.WriteString(tt.input)
			}()

			// Wait for REPL to complete or timeout
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("REPL execution failed: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Error("REPL test timed out")
			}

			// Close stdout and restore
			stdoutWriter.Close()
			os.Stdin = oldStdin
			os.Stdout = oldStdout

			// Read captured output
			output, _ := io.ReadAll(stdoutReader)
			stdoutReader.Close()
			outputStr := string(output)

			// Verify expected error output
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectedOutput, outputStr)
			}
		})
	}
}

func TestReplCommand_Execute_WelcomeMessage(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("exit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify welcome message
	if !strings.Contains(outputStr, "Welcome to MyShift REPL") {
		t.Errorf("Expected welcome message, got: %s", outputStr)
	}
}

// Test simulating EOF (Ctrl+D)
func TestReplCommand_Execute_EOF(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "test@example.com",
	}

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Close stdin immediately to simulate EOF
	stdinWriter.Close()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Unexpected error on EOF: %v", err)
		}
	case <-time.After(2 * time.Second):
		// EOF should cause immediate exit, so timeout is unexpected
		t.Error("REPL should exit immediately on EOF")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Should handle EOF gracefully
	if strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected no error on EOF, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_MyUserFallback(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "myuser@example.com",
	}

	// Setup mock for my_user fallback
	fixture.MockClient.AddUser("MYUSER123", "My User", "myuser@example.com")
	fixture.MockClient.AddOnCall("MYUSER123", "My User", "myuser@example.com",
		fixture.Now.Add(24*time.Hour),
		fixture.Now.Add(48*time.Hour))

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin - next command without email (should use my_user)
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("next\nexit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify next command used my_user
	if !strings.Contains(outputStr, "Next shift:") {
		t.Errorf("Expected next command to work with my_user fallback, got: %s", outputStr)
	}
}

func TestReplCommand_Execute_HelpWithMyUser(t *testing.T) {
	fixture := NewTestFixture()
	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "myuser@example.com",
	}

	// Create pipes for stdin and stdout
	stdinReader, stdinWriter, _ := os.Pipe()
	stdoutReader, stdoutWriter, _ := os.Pipe()

	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Replace stdin/stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	replCmd := NewReplCommand(fixture.MockClient, config)

	// Channel to signal completion
	done := make(chan error, 1)

	// Start REPL in goroutine
	go func() {
		err := replCmd.Execute("SCHED123")
		done <- err
	}()

	// Write input and close stdin
	go func() {
		defer stdinWriter.Close()
		stdinWriter.WriteString("help\nexit\n")
	}()

	// Wait for REPL to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("REPL execution failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("REPL test timed out")
	}

	// Close stdout and restore
	stdoutWriter.Close()
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	// Read captured output
	output, _ := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	outputStr := string(output)

	// Verify help shows my_user as default
	if !strings.Contains(outputStr, "(defaults to myuser@example.com)") {
		t.Errorf("Expected help to show my_user default, got: %s", outputStr)
	}
}

// Benchmark to ensure performance is reasonable
func BenchmarkReplCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	fixture.MockClient.AddUser("USER001", "John Doe", "john@example.com")

	config := &myshift.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "john@example.com",
	}

	cmd := NewReplCommand(fixture.MockClient, config)

	// Simple input that just quits
	input := "quit\n"
	inputBytes := []byte(input)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate stdin
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		// Silence output
		oldStdout := os.Stdout
		os.Stdout, _ = os.Open(os.DevNull)

		// Write input
		go func() {
			defer w.Close()
			w.Write(inputBytes)
		}()

		// Execute
		_ = cmd.Execute("SCHED123")

		// Restore
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}
}
