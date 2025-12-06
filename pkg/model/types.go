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

package model

// TransferConfig contains configuration for file transfer operations
type TransferConfig struct {
	// Source is the local source path
	Source string

	// SSMInstanceID is the target EC2 instance ID
	SSMInstanceID string

	// Destination is the remote destination path
	Destination string

	// BucketName is the S3 bucket used for transfer
	BucketName string

	// MaxRetries is the maximum number of retry attempts for AWS operations
	MaxRetries int

	// RetryDelay is the base delay between retries (exponential backoff)
	RetryDelay int
}

// AWSConfig contains AWS-specific configuration
type AWSConfig struct {
	// Region is the AWS region
	Region string

	// Profile is the AWS CLI profile to use
	Profile string

	// Bucket is the default S3 bucket for transfers
	Bucket string
}

// LogConfig contains logging configuration
type LogConfig struct {
	// Format is the log format (text, json, color)
	Format string

	// Level is the log level (debug, info, warn, error)
	Level string
}

// Config is the root configuration structure
type Config struct {
	// AWS configuration
	AWS AWSConfig `yaml:"aws"`

	// Log configuration
	Log LogConfig `yaml:"log"`

	// Transfer configuration defaults
	Transfer TransferDefaults `yaml:"transfer"`
}

// TransferDefaults contains default values for transfer operations
type TransferDefaults struct {
	// MaxRetries is the default maximum retry attempts
	MaxRetries int `yaml:"max_retries"`

	// RetryDelay is the default base retry delay in seconds
	RetryDelay int `yaml:"retry_delay"`
}
