/*
Copyright © 2025 Jayson Grace <jayson.e.grace@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"testing"
)

func TestCompletionCmd(t *testing.T) {
	if completionCmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("completionCmd.Use = %v, want 'completion [bash|zsh|fish|powershell]'", completionCmd.Use)
	}

	if completionCmd.Short != "Generate shell completion scripts" {
		t.Errorf("completionCmd.Short description incorrect")
	}

	// Verify DisableFlagsInUseLine is set
	if !completionCmd.DisableFlagsInUseLine {
		t.Error("completionCmd.DisableFlagsInUseLine should be true")
	}
}

func TestCompletionCmdValidArgs(t *testing.T) {
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}

	if len(completionCmd.ValidArgs) != len(expectedArgs) {
		t.Errorf("completionCmd.ValidArgs length = %d, want %d", len(completionCmd.ValidArgs), len(expectedArgs))
	}

	// Check that all expected args are present
	for _, expected := range expectedArgs {
		found := false
		for _, actual := range completionCmd.ValidArgs {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected valid arg %q not found in ValidArgs", expected)
		}
	}
}

func TestCompletionCmdArgs(t *testing.T) {
	// Test that Args is set to ExactValidArgs(1)
	// This is hard to test directly, so we verify the function exists
	if completionCmd.Args == nil {
		t.Error("completionCmd.Args is nil, expected ExactValidArgs(1)")
	}
}

func TestCompletionCmdRun(t *testing.T) {
	// Verify Run function is defined
	if completionCmd.Run == nil {
		t.Error("completionCmd.Run is nil, expected a Run function")
	}
}

func TestCompletionCmdRegisteredWithRoot(t *testing.T) {
	// Verify completion command is registered with root command
	commands := rootCmd.Commands()
	hasCompletionCmd := false

	for _, cmd := range commands {
		if cmd.Use == "completion [bash|zsh|fish|powershell]" {
			hasCompletionCmd = true
			break
		}
	}

	if !hasCompletionCmd {
		t.Error("completion command not registered with root command")
	}
}

func TestCompletionCmdAllShellsSupported(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		found := false
		for _, validArg := range completionCmd.ValidArgs {
			if validArg == shell {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Shell %q not found in ValidArgs", shell)
		}
	}
}

func TestCompletionCmdStructure(t *testing.T) {
	tests := []struct {
		name     string
		check    func() bool
		errorMsg string
	}{
		{
			name:     "has Use field",
			check:    func() bool { return completionCmd.Use != "" },
			errorMsg: "completionCmd.Use is empty",
		},
		{
			name:     "has Short description",
			check:    func() bool { return completionCmd.Short != "" },
			errorMsg: "completionCmd.Short is empty",
		},
		{
			name:     "has Long description",
			check:    func() bool { return completionCmd.Long != "" },
			errorMsg: "completionCmd.Long is empty",
		},
		{
			name:     "has Run function",
			check:    func() bool { return completionCmd.Run != nil },
			errorMsg: "completionCmd.Run is nil",
		},
		{
			name:     "has ValidArgs",
			check:    func() bool { return len(completionCmd.ValidArgs) > 0 },
			errorMsg: "completionCmd.ValidArgs is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Error(tt.errorMsg)
			}
		})
	}
}
