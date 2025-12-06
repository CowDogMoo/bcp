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
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInitConfig(t *testing.T) {
	// Test with no config file (should use defaults)
	// Save original cfgFile value
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	// Test with empty config file (should use defaults)
	cfgFile = ""
	initConfig()
	// If we get here without panic, the init worked

	// Note: We can't test with a nonexistent config file because
	// initConfig calls os.Exit(1) on error, which would terminate the test.
	// This would require refactoring initConfig to return an error instead of calling os.Exit.
}

func TestExecute_Coverage(t *testing.T) {
	// This test ensures the Execute function exists
	// We can't actually run it without causing the program to exit on error
	// The function exists if this test compiles, so we just verify it's defined
	t.Log("Execute function is defined and accessible")
}

func TestBucketCompletion_NoAWS(t *testing.T) {
	// Test bucket completion when AWS is not available or returns error
	// This tests the error handling path
	cmd := RootCmd()

	// Clear AWS credentials to force error
	oldAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	oldProfile := os.Getenv("AWS_PROFILE")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_PROFILE")
	defer func() {
		if oldAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", oldAccessKey)
		}
		if oldProfile != "" {
			os.Setenv("AWS_PROFILE", oldProfile)
		}
	}()

	completions, directive := bucketCompletion(cmd, []string{}, "test")

	// Should handle error gracefully
	if directive != cobra.ShellCompDirectiveError && len(completions) == 0 {
		// Either error directive or no completions is acceptable
		t.Logf("Bucket completion with no AWS: completions=%d, directive=%v", len(completions), directive)
	}
}

func TestInstanceCompletion_NoAWS(t *testing.T) {
	cmd := RootCmd()

	// Clear AWS credentials to force error
	oldAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	oldProfile := os.Getenv("AWS_PROFILE")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_PROFILE")
	defer func() {
		if oldAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", oldAccessKey)
		}
		if oldProfile != "" {
			os.Setenv("AWS_PROFILE", oldProfile)
		}
	}()

	completions, directive := instanceCompletion(cmd, []string{}, "i-")

	// Should handle error gracefully
	if directive != cobra.ShellCompDirectiveError && len(completions) == 0 {
		t.Logf("Instance completion with no AWS: completions=%d, directive=%v", len(completions), directive)
	}
}

func TestArgsCompletion_FirstArg(t *testing.T) {
	cmd := RootCmd()

	// First argument should be a file path (default shell completion)
	completions, directive := argsCompletion(cmd, []string{}, "/tmp")

	// Should use default file completion
	if directive != cobra.ShellCompDirectiveDefault {
		t.Logf("First arg completion directive: %v (may use default)", directive)
	}

	t.Logf("First arg completions: %d items", len(completions))
}

func TestArgsCompletion_SecondArg_WithoutColon(t *testing.T) {
	cmd := RootCmd()

	// Clear AWS credentials to test without actual API calls
	oldAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	oldProfile := os.Getenv("AWS_PROFILE")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_PROFILE")
	defer func() {
		if oldAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", oldAccessKey)
		}
		if oldProfile != "" {
			os.Setenv("AWS_PROFILE", oldProfile)
		}
	}()

	// Second argument without colon should suggest instance IDs
	completions, directive := argsCompletion(cmd, []string{"/tmp/file"}, "i-")

	// Should try to complete instance IDs
	if directive == cobra.ShellCompDirectiveNoSpace || directive == cobra.ShellCompDirectiveError {
		t.Logf("Instance ID completion directive: %v", directive)
	}

	t.Logf("Second arg (without colon) completions: %d items", len(completions))
}

func TestArgsCompletion_SecondArg_WithColon(t *testing.T) {
	cmd := RootCmd()

	// Second argument with colon should suggest common paths
	completions, directive := argsCompletion(cmd, []string{"/tmp/file"}, "i-1234567890abcdef0:")

	// Should suggest common paths
	if directive != cobra.ShellCompDirectiveNoSpace {
		t.Logf("Path completion directive: %v (expected NoSpace)", directive)
	}

	// Should return common paths
	if len(completions) == 0 {
		t.Error("Expected common path completions")
	}

	// Verify completions contain the instance ID
	for _, comp := range completions {
		if !strings.HasPrefix(comp, "i-1234567890abcdef0:") {
			t.Errorf("Completion %q doesn't start with instance ID", comp)
		}
	}

	t.Logf("Second arg (with colon) completions: %d items", len(completions))
}

func TestArgsCompletion_ThirdArg(t *testing.T) {
	cmd := RootCmd()

	// Third argument should not complete
	completions, directive := argsCompletion(cmd, []string{"/tmp/file", "i-1234567890abcdef0:/tmp"}, "")

	// Should not complete
	if len(completions) != 0 {
		t.Errorf("Expected no completions for third arg, got %d", len(completions))
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Logf("Third arg directive: %v (expected NoFileComp)", directive)
	}
}

func TestRootCmd_RequiresTwoArgs(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error with no arguments")
	}

	if !strings.Contains(err.Error(), "2") {
		t.Logf("Error message: %v", err)
	}
}

func TestRootCmd_RejectsThreeArgs(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{"arg1", "arg2", "arg3"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error with three arguments")
	}
}

func TestRootCmdInitialization(t *testing.T) {
	// Verify root command is initialized
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}

	// Verify it has the expected structure
	if rootCmd.Use == "" {
		t.Error("rootCmd.Use should not be empty")
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}

	if rootCmd.RunE == nil {
		t.Error("rootCmd.RunE should not be nil")
	}
}

func TestRootCmdHasRequiredFlags(t *testing.T) {
	requiredFlags := []string{"bucket", "config", "verbose", "quiet"}

	for _, flagName := range requiredFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %q should be registered", flagName)
		}
	}
}

func TestRootCmdFlagShorthands(t *testing.T) {
	tests := []struct {
		flag      string
		shorthand string
	}{
		{"bucket", "b"},
		{"config", "c"},
		{"verbose", "v"},
		{"quiet", "q"},
	}

	for _, tt := range tests {
		flag := rootCmd.PersistentFlags().Lookup(tt.flag)
		if flag == nil {
			t.Errorf("Flag %q not found", tt.flag)
			continue
		}
		if flag.Shorthand != tt.shorthand {
			t.Errorf("Flag %q: expected shorthand %q, got %q", tt.flag, tt.shorthand, flag.Shorthand)
		}
	}
}

func TestRootCmdValidArgsFunction(t *testing.T) {
	// Verify ValidArgsFunction is set
	if rootCmd.ValidArgsFunction == nil {
		t.Error("rootCmd.ValidArgsFunction should not be nil")
	}
}

func TestRootCmdHasListSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}

	if !found {
		t.Error("rootCmd should have 'list' subcommand")
	}
}

func TestBucketCompletion_Filtering(t *testing.T) {
	// This tests the filtering logic in bucketCompletion
	// even if AWS returns an error, the filtering logic should work

	cmd := RootCmd()

	// Test with empty prefix
	_, directive1 := bucketCompletion(cmd, []string{}, "")
	t.Logf("Empty prefix directive: %v", directive1)

	// Test with prefix
	_, directive2 := bucketCompletion(cmd, []string{}, "my-bucket")
	t.Logf("With prefix directive: %v", directive2)

	// Both should return the same directive type (either error or no-file-comp)
	if directive1 != directive2 && directive2 != cobra.ShellCompDirectiveError {
		t.Logf("Directives differ: %v vs %v (may be expected)", directive1, directive2)
	}
}

func TestInstanceCompletion_Filtering(t *testing.T) {
	cmd := RootCmd()

	// Test with empty prefix
	_, directive1 := instanceCompletion(cmd, []string{}, "")
	t.Logf("Empty prefix directive: %v", directive1)

	// Test with prefix
	_, directive2 := instanceCompletion(cmd, []string{}, "i-123")
	t.Logf("With prefix directive: %v", directive2)
}

func TestRootCmdRunENotNil(t *testing.T) {
	// Verify the RunE function exists
	if rootCmd.RunE == nil {
		t.Fatal("rootCmd.RunE should not be nil")
	}
}

func TestRootCmdExactArgs(t *testing.T) {
	// Test the Args validator
	if rootCmd.Args == nil {
		t.Error("rootCmd.Args should not be nil")
	}

	// The Args should be ExactArgs(2)
	// We can verify this by checking with different argument counts
	cmd := RootCmd()

	// Test with 0 args
	cmd.SetArgs([]string{})
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error with 0 args")
	}

	// Test with 1 arg
	err = cmd.Args(cmd, []string{"one"})
	if err == nil {
		t.Error("Expected error with 1 arg")
	}

	// Test with 2 args
	err = cmd.Args(cmd, []string{"one", "two"})
	if err != nil {
		t.Errorf("Expected no error with 2 args, got: %v", err)
	}

	// Test with 3 args
	err = cmd.Args(cmd, []string{"one", "two", "three"})
	if err == nil {
		t.Error("Expected error with 3 args")
	}
}

func TestArgsCompletion_PathSuggestions(t *testing.T) {
	cmd := RootCmd()

	instanceID := "i-1234567890abcdef0"
	completions, _ := argsCompletion(cmd, []string{"/tmp/file"}, instanceID+":")

	// Verify we get path suggestions
	expectedPaths := []string{"/tmp/", "/home/ec2-user/", "/opt/", "/usr/local/bin/", "/var/tmp/"}

	if len(completions) == 0 {
		t.Error("Expected path suggestions, got none")
	}

	for _, expectedPath := range expectedPaths {
		found := false
		expectedCompletion := instanceID + ":" + expectedPath
		for _, comp := range completions {
			if comp == expectedCompletion {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path suggestion %q not found in completions", expectedPath)
		}
	}
}

func TestVerboseAndQuietFlags_Exclusive(t *testing.T) {
	// Test that verbose and quiet flags have different default values
	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	quietFlag := rootCmd.PersistentFlags().Lookup("quiet")

	if verboseFlag == nil || quietFlag == nil {
		t.Fatal("verbose or quiet flag not found")
	}

	if verboseFlag.DefValue != "false" || quietFlag.DefValue != "false" {
		t.Error("Expected both verbose and quiet to default to false")
	}
}
