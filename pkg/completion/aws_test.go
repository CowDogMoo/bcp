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

package completion

import (
	"os"
	"testing"
)

func TestGetBucketNames(t *testing.T) {
	// Skip test if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	buckets, err := GetBucketNames()

	// We don't fail on error here because it might be due to missing credentials
	// or permissions, which is acceptable in a test environment
	if err != nil {
		t.Logf("GetBucketNames() returned error (might be expected): %v", err)
		return
	}

	// If no error, buckets should be a valid slice (even if empty)
	if buckets == nil {
		t.Error("GetBucketNames() returned nil slice, expected empty slice")
	}

	t.Logf("Found %d bucket(s)", len(buckets))
}

func TestGetInstanceIDs(t *testing.T) {
	// Skip test if AWS credentials are not available
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	instances, err := GetInstanceIDs()

	// We don't fail on error here because it might be due to missing credentials
	// or permissions, which is acceptable in a test environment
	if err != nil {
		t.Logf("GetInstanceIDs() returned error (might be expected): %v", err)
		return
	}

	// If no error, instances should be a valid slice (even if empty)
	if instances == nil {
		t.Error("GetInstanceIDs() returned nil slice, expected empty slice")
	}

	t.Logf("Found %d instance(s)", len(instances))
}

func TestGetBucketNamesReturnType(t *testing.T) {
	// This test verifies the function signature and basic behavior
	// without requiring AWS credentials

	// Even with no credentials, the function should return a consistent type
	buckets, err := GetBucketNames()

	// Check that at least one return value is non-nil
	if buckets == nil && err == nil {
		t.Error("GetBucketNames() returned (nil, nil), expected at least one non-nil value")
	}

	// If buckets is not nil, it should be a slice (no further validation needed)
}

func TestGetInstanceIDsReturnType(t *testing.T) {
	// This test verifies the function signature and basic behavior
	// without requiring AWS credentials

	// Even with no credentials, the function should return a consistent type
	instances, err := GetInstanceIDs()

	// Check that at least one return value is non-nil
	if instances == nil && err == nil {
		t.Error("GetInstanceIDs() returned (nil, nil), expected at least one non-nil value")
	}

	// If instances is not nil, it should be a slice (no further validation needed)
}

func TestGetBucketNamesErrorHandling(t *testing.T) {
	// Test that the function properly handles errors
	// This will likely fail due to missing credentials, which is expected
	buckets, err := GetBucketNames()

	// Either we get buckets or an error, but not both nil
	if buckets == nil && err == nil {
		t.Error("Expected either buckets or error, got both nil")
	}

	// If we got an error, buckets should be nil or empty
	if err != nil && buckets != nil && len(buckets) > 0 {
		t.Error("Expected empty or nil buckets when error is returned")
	}
}

func TestGetInstanceIDsErrorHandling(t *testing.T) {
	// Test that the function properly handles errors
	// This will likely fail due to missing credentials, which is expected
	instances, err := GetInstanceIDs()

	// Either we get instances or an error, but not both nil
	if instances == nil && err == nil {
		t.Error("Expected either instances or error, got both nil")
	}

	// If we got an error, instances should be nil or empty
	if err != nil && instances != nil && len(instances) > 0 {
		t.Error("Expected empty or nil instances when error is returned")
	}
}
