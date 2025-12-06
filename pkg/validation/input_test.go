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

package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSourcePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid directory", tmpDir, false},
		{"valid file", testFile, false},
		{"empty path", "", true},
		{"non-existent path", "/nonexistent/path", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourcePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourcePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSSMPath(t *testing.T) {
	tests := []struct {
		name             string
		ssmPath          string
		wantInstanceID   string
		wantDestination  string
		wantErr          bool
	}{
		{
			name:             "valid SSM path",
			ssmPath:          "i-1234567890abcdef0:/home/ec2-user",
			wantInstanceID:   "i-1234567890abcdef0",
			wantDestination:  "/home/ec2-user",
			wantErr:          false,
		},
		{
			name:    "empty path",
			ssmPath: "",
			wantErr: true,
		},
		{
			name:    "missing colon",
			ssmPath: "i-1234567890abcdef0/home/ec2-user",
			wantErr: true,
		},
		{
			name:    "invalid instance ID",
			ssmPath: "invalid:/home/ec2-user",
			wantErr: true,
		},
		{
			name:    "relative destination path",
			ssmPath: "i-1234567890abcdef0:home/ec2-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instanceID, destination, err := ValidateSSMPath(tt.ssmPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSMPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if instanceID != tt.wantInstanceID {
					t.Errorf("ValidateSSMPath() instanceID = %v, want %v", instanceID, tt.wantInstanceID)
				}
				if destination != tt.wantDestination {
					t.Errorf("ValidateSSMPath() destination = %v, want %v", destination, tt.wantDestination)
				}
			}
		})
	}
}

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name    string
		bucket  string
		wantErr bool
	}{
		{"valid bucket", "my-test-bucket", false},
		{"valid with dots", "my.test.bucket", false},
		{"valid with numbers", "mybucket123", false},
		{"empty bucket", "", true},
		{"too short", "ab", true},
		{"too long", "this-is-a-very-long-bucket-name-that-exceeds-the-maximum-allowed-length-for-s3-buckets", true},
		{"uppercase letters", "MyBucket", true},
		{"consecutive dots", "my..bucket", true},
		{"IP address format", "192.168.1.1", true},
		{"starts with hyphen", "-mybucket", true},
		{"ends with hyphen", "mybucket-", true},
		{"adjacent period and hyphen", "my.-bucket", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBucketName(tt.bucket)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBucketName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
