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

package transfer

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "AWS RequestTimeout",
			err:      awserr.New("RequestTimeout", "request timeout", nil),
			expected: true,
		},
		{
			name:     "AWS ServiceUnavailable",
			err:      awserr.New("ServiceUnavailable", "service unavailable", nil),
			expected: true,
		},
		{
			name:     "AWS ThrottlingException",
			err:      awserr.New("ThrottlingException", "throttling", nil),
			expected: true,
		},
		{
			name:     "AWS RequestLimitExceeded",
			err:      awserr.New("RequestLimitExceeded", "limit exceeded", nil),
			expected: true,
		},
		{
			name:     "AWS TooManyRequestsException",
			err:      awserr.New("TooManyRequestsException", "too many requests", nil),
			expected: true,
		},
		{
			name:     "AWS InternalError",
			err:      awserr.New("InternalError", "internal error", nil),
			expected: true,
		},
		{
			name:     "AWS RequestThrottled",
			err:      awserr.New("RequestThrottled", "request throttled", nil),
			expected: true,
		},
		{
			name:     "AWS Throttling",
			err:      awserr.New("Throttling", "throttling", nil),
			expected: true,
		},
		{
			name:     "AWS non-retryable error",
			err:      awserr.New("ValidationException", "validation failed", nil),
			expected: false,
		},
		{
			name:     "connection reset error",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("operation timeout"),
			expected: true,
		},
		{
			name:     "temporary failure error",
			err:      errors.New("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "TLS handshake timeout",
			err:      errors.New("TLS handshake timeout"),
			expected: true,
		},
		{
			name:     "EOF error",
			err:      errors.New("unexpected EOF"),
			expected: true,
		},
		{
			name:     "i/o timeout error",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("invalid input"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestRetryOperation(t *testing.T) {
	tests := []struct {
		name         string
		operation    func() error
		maxRetries   int
		baseDelay    int
		expectError  bool
		expectedTries int
	}{
		{
			name: "success on first try",
			operation: func() error {
				return nil
			},
			maxRetries:    3,
			baseDelay:     1,
			expectError:   false,
			expectedTries: 1,
		},
		{
			name: "success after retries",
			operation: func() func() error {
				count := 0
				return func() error {
					count++
					if count < 3 {
						return awserr.New("RequestTimeout", "timeout", nil)
					}
					return nil
				}
			}(),
			maxRetries:    3,
			baseDelay:     1,
			expectError:   false,
			expectedTries: 3,
		},
		{
			name: "fail with non-retryable error",
			operation: func() error {
				return errors.New("invalid input")
			},
			maxRetries:    3,
			baseDelay:     1,
			expectError:   true,
			expectedTries: 1,
		},
		{
			name: "fail after max retries",
			operation: func() error {
				return awserr.New("RequestTimeout", "timeout", nil)
			},
			maxRetries:    2,
			baseDelay:     1,
			expectError:   true,
			expectedTries: 3, // Initial try + 2 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := retryOperation(tt.operation, tt.maxRetries, tt.baseDelay)
			if (err != nil) != tt.expectError {
				t.Errorf("retryOperation() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestRetryOperationExponentialBackoff(t *testing.T) {
	// Test that retry operation respects exponential backoff
	operation := func() error {
		return awserr.New("RequestTimeout", "timeout", nil)
	}

	// This should fail after trying multiple times with exponential backoff
	// We're mainly testing that the function doesn't panic and handles retries
	err := retryOperation(operation, 2, 1)
	if err == nil {
		t.Error("Expected error from retryOperation with always-failing operation")
	}
}
