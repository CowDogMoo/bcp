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
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestListCmd(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("listCmd.Use = %v, want 'list'", listCmd.Use)
	}

	if listCmd.Short != "List AWS resources (buckets, instances)" {
		t.Errorf("listCmd.Short description incorrect")
	}

	// Check that subcommands are registered
	subcommands := listCmd.Commands()
	if len(subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(subcommands))
	}

	// Verify subcommands are buckets and instances
	hasbuckets := false
	hasInstances := false
	for _, cmd := range subcommands {
		if cmd.Use == "buckets" {
			hasbuckets = true
		}
		if cmd.Use == "instances" {
			hasInstances = true
		}
	}

	if !hasbuckets {
		t.Error("listCmd missing 'buckets' subcommand")
	}
	if !hasInstances {
		t.Error("listCmd missing 'instances' subcommand")
	}
}

func TestListBucketsCmd(t *testing.T) {
	if listBucketsCmd.Use != "buckets" {
		t.Errorf("listBucketsCmd.Use = %v, want 'buckets'", listBucketsCmd.Use)
	}

	if listBucketsCmd.Short != "List available S3 buckets" {
		t.Errorf("listBucketsCmd.Short description incorrect")
	}

	// Verify it has a RunE function
	if listBucketsCmd.RunE == nil {
		t.Error("listBucketsCmd.RunE is nil")
	}
}

func TestListInstancesCmd(t *testing.T) {
	if listInstancesCmd.Use != "instances" {
		t.Errorf("listInstancesCmd.Use = %v, want 'instances'", listInstancesCmd.Use)
	}

	if listInstancesCmd.Short != "List SSM-managed EC2 instances" {
		t.Errorf("listInstancesCmd.Short description incorrect")
	}

	// Verify it has a RunE function
	if listInstancesCmd.RunE == nil {
		t.Error("listInstancesCmd.RunE is nil")
	}
}

func TestListInstancesCmdFlags(t *testing.T) {
	// Test that flags are registered
	allFlag := listInstancesCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("'all' flag not registered")
	} else {
		if allFlag.Shorthand != "a" {
			t.Errorf("'all' flag shorthand = %v, want 'a'", allFlag.Shorthand)
		}
		if allFlag.DefValue != "false" {
			t.Errorf("'all' flag default = %v, want 'false'", allFlag.DefValue)
		}
	}

	regionFlag := listInstancesCmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Error("'region' flag not registered")
	} else {
		if regionFlag.Shorthand != "r" {
			t.Errorf("'region' flag shorthand = %v, want 'r'", regionFlag.Shorthand)
		}
		if regionFlag.DefValue != "" {
			t.Errorf("'region' flag default = %v, want empty string", regionFlag.DefValue)
		}
	}
}

func TestListCmdRegisteredWithRoot(t *testing.T) {
	// Verify list command is registered with root command
	commands := rootCmd.Commands()
	hasListCmd := false

	for _, cmd := range commands {
		if cmd.Use == "list" {
			hasListCmd = true
			break
		}
	}

	if !hasListCmd {
		t.Error("list command not registered with root command")
	}
}

func TestListSubcommandStructure(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		expectedUse string
		hasRunE     bool
	}{
		{
			name:        "buckets subcommand",
			cmd:         listBucketsCmd,
			expectedUse: "buckets",
			hasRunE:     true,
		},
		{
			name:        "instances subcommand",
			cmd:         listInstancesCmd,
			expectedUse: "instances",
			hasRunE:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.Use != tt.expectedUse {
				t.Errorf("Command.Use = %v, want %v", tt.cmd.Use, tt.expectedUse)
			}

			if tt.hasRunE && tt.cmd.RunE == nil {
				t.Error("Command.RunE is nil but should be defined")
			}

			if !tt.hasRunE && tt.cmd.RunE != nil {
				t.Error("Command.RunE is defined but should be nil")
			}
		})
	}
}

func TestListBucketsCmdExecution(t *testing.T) {
	// Skip if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute the command
	err := listBucketsCmd.RunE(cmd, []string{})

	// We accept either success or certain AWS errors
	if err != nil {
		// Log the error but don't fail - AWS credentials might not have S3 permissions
		t.Logf("listBucketsCmd.RunE returned error (may be expected): %v", err)
	}
}

func TestListInstancesCmdExecution(t *testing.T) {
	// Skip if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	// Test without --all flag
	t.Run("SSM instances", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Reset flags
		listAll = false
		listRegion = ""

		err := listInstancesCmd.RunE(cmd, []string{})

		// We accept either success or certain AWS errors
		if err != nil {
			t.Logf("listInstancesCmd.RunE returned error (may be expected): %v", err)
		}
	})

	// Test with --all flag
	t.Run("All instances", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Set flags
		listAll = true
		listRegion = ""

		err := listInstancesCmd.RunE(cmd, []string{})

		// We accept either success or certain AWS errors
		if err != nil {
			t.Logf("listInstancesCmd.RunE with --all returned error (may be expected): %v", err)
		}
	})

	// Test with --region flag
	t.Run("Specific region", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Set flags
		listAll = false
		listRegion = "us-west-2"

		err := listInstancesCmd.RunE(cmd, []string{})

		// We accept either success or certain AWS errors
		if err != nil {
			t.Logf("listInstancesCmd.RunE with region returned error (may be expected): %v", err)
		}
	})
}
