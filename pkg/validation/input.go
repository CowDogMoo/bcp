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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ValidateSourcePath(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("source path cannot be empty")
	}

	cleanPath := filepath.Clean(path)

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("source path does not exist: %s", cleanPath)
		}
		return false, fmt.Errorf("failed to access source path: %w", err)
	}

	if !info.Mode().IsDir() && !info.Mode().IsRegular() {
		return false, fmt.Errorf("source path is not a regular file or directory: %s", cleanPath)
	}

	return info.Mode().IsDir(), nil
}

func ValidateSSMPath(ssmPath string) (instanceID string, destination string, err error) {
	if ssmPath == "" {
		return "", "", fmt.Errorf("SSM path cannot be empty")
	}

	parts := strings.Split(ssmPath, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid SSM path format, expected 'instance-id:destination', got: %s", ssmPath)
	}

	instanceID = strings.TrimSpace(parts[0])
	destination = strings.TrimSpace(parts[1])

	if instanceID == "" {
		return "", "", fmt.Errorf("SSM instance ID cannot be empty")
	}

	if destination == "" {
		return "", "", fmt.Errorf("destination path cannot be empty")
	}

	instanceIDPattern := regexp.MustCompile(`^i-[0-9a-f]{8,17}$`)
	if !instanceIDPattern.MatchString(instanceID) {
		return "", "", fmt.Errorf("invalid SSM instance ID format: %s (expected format: i-xxxxxxxxx)", instanceID)
	}

	if !filepath.IsAbs(destination) {
		return "", "", fmt.Errorf("destination must be an absolute path, got: %s", destination)
	}

	return instanceID, destination, nil
}

func ValidateBucketName(bucket string) error {
	if bucket == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}

	// S3 bucket naming rules:
	// - Must be between 3 and 63 characters long
	// - Can only contain lowercase letters, numbers, dots, and hyphens
	// - Must start and end with a letter or number
	// - Cannot contain uppercase characters or underscores
	// - Cannot be formatted as an IP address

	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters, got: %d", len(bucket))
	}

	bucketPattern := regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]*[a-z0-9]$`)
	if !bucketPattern.MatchString(bucket) {
		return fmt.Errorf("invalid bucket name format: %s (must contain only lowercase letters, numbers, dots, and hyphens)", bucket)
	}

	ipPattern := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if ipPattern.MatchString(bucket) {
		return fmt.Errorf("bucket name cannot be formatted as an IP address: %s", bucket)
	}

	if strings.Contains(bucket, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive periods: %s", bucket)
	}

	if strings.Contains(bucket, ".-") || strings.Contains(bucket, "-.") {
		return fmt.Errorf("bucket name cannot have adjacent periods and hyphens: %s", bucket)
	}

	return nil
}
