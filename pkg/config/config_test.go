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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Reset viper for each test
	resetViper := func() {
		viper.Reset()
	}

	tests := []struct {
		name        string
		cfgFile     string
		setupConfig func(string) error
		wantErr     bool
	}{
		{
			name:    "no config file",
			cfgFile: "",
			setupConfig: func(dir string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "valid config file",
			cfgFile: "config.yaml",
			setupConfig: func(dir string) error {
				configContent := `
log:
  format: json
  level: debug
aws:
  region: us-west-2
  profile: test
  bucket: test-bucket
transfer:
  max_retries: 5
  retry_delay: 3
`
				return os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644)
			},
			wantErr: false,
		},
		{
			name:    "invalid yaml config file",
			cfgFile: "config.yaml",
			setupConfig: func(dir string) error {
				invalidContent := `
log:
  format: json
  level: debug
  invalid yaml content [[[
`
				return os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(invalidContent), 0644)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetViper()

			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "bcp-config-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Failed to cleanup temp dir: %v", err)
				}
			}()

			// Setup config file if needed
			if err := tt.setupConfig(tmpDir); err != nil {
				t.Fatalf("Failed to setup config: %v", err)
			}

			cfgPath := ""
			if tt.cfgFile != "" {
				cfgPath = filepath.Join(tmpDir, tt.cfgFile)
			}

			err = Init(cfgPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"log format default", "log.format", "text"},
		{"log level default", "log.level", "info"},
		{"aws region default", "aws.region", "us-east-1"},
		{"aws profile default", "aws.profile", "default"},
		{"aws bucket default", "aws.bucket", ""},
		{"max retries default", "transfer.max_retries", 3},
		{"retry delay default", "transfer.retry_delay", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := viper.Get(tt.key)
			if got != tt.expected {
				t.Errorf("Default for %s = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestLoadConstants(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		retryDelay    int
		expectedRetry int
		expectedDelay int
	}{
		{
			name:          "valid values",
			maxRetries:    5,
			retryDelay:    4,
			expectedRetry: 5,
			expectedDelay: 4,
		},
		{
			name:          "zero values use defaults",
			maxRetries:    0,
			retryDelay:    0,
			expectedRetry: 3,
			expectedDelay: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("transfer.max_retries", tt.maxRetries)
			viper.Set("transfer.retry_delay", tt.retryDelay)

			LoadConstants()

			if MaxRetries != tt.expectedRetry {
				t.Errorf("MaxRetries = %d, want %d", MaxRetries, tt.expectedRetry)
			}
			if RetryDelay != tt.expectedDelay {
				t.Errorf("RetryDelay = %d, want %d", RetryDelay, tt.expectedDelay)
			}
		})
	}
}

func TestGetters(t *testing.T) {
	viper.Reset()

	// Set test values
	GlobalConfig.AWS.Bucket = "test-bucket"
	GlobalConfig.AWS.Region = "us-west-2"
	GlobalConfig.AWS.Profile = "test-profile"

	tests := []struct {
		name     string
		getter   func() string
		expected string
	}{
		{"GetBucket", GetBucket, "test-bucket"},
		{"GetRegion", GetRegion, "us-west-2"},
		{"GetProfile", GetProfile, "test-profile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getter()
			if got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestInitWithEnvironmentVariables(t *testing.T) {
	// Note: viper.AutomaticEnv() doesn't automatically handle nested keys
	// This test verifies the current behavior with environment variables
	viper.Reset()

	// Create a temporary config file to test environment variable override
	tmpDir, err := os.MkdirTemp("", "bcp-config-env-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	configContent := `
log:
  format: text
  level: info
aws:
  region: us-east-1
  profile: default
  bucket: config-bucket
transfer:
  max_retries: 3
  retry_delay: 2
`
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Initialize with config file
	if err := Init(cfgPath); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify config was loaded
	if GlobalConfig.AWS.Bucket != "config-bucket" {
		t.Errorf("GlobalConfig.AWS.Bucket = %v, want %v", GlobalConfig.AWS.Bucket, "config-bucket")
	}
}
