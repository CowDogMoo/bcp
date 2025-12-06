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
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/cowdogmoo/bcp/pkg/model"
	"github.com/l50/awsutils/s3"
	"github.com/l50/awsutils/ssm"
)

func Execute(config model.TransferConfig) error {
	log.Info("Starting transfer from %s to %s:%s", config.Source, config.SSMInstanceID, config.Destination)

	s3Connection := s3.CreateConnection()
	ssmConnection := ssm.CreateConnection()

	uploadPath := strings.TrimPrefix(config.Source, "./")
	s3URL := fmt.Sprintf("s3://%s/%s", config.BucketName, uploadPath)

	log.Info("Uploading %s to S3 bucket %s...", config.Source, config.BucketName)
	if err := retryOperation(func() error {
		return s3.UploadBucketDir(s3Connection.Session, config.BucketName, uploadPath)
	}, config.MaxRetries, config.RetryDelay); err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	log.Info("Upload to S3 completed successfully")

	log.Info("Checking if AWS CLI is installed on instance %s...", config.SSMInstanceID)
	awsCLICheck, err := ssm.CheckAWSCLIInstalled(ssmConnection.Client, config.SSMInstanceID)
	if err != nil {
		return fmt.Errorf("failed to check AWS CLI installation: %w", err)
	}
	if !awsCLICheck {
		return fmt.Errorf("AWS CLI is not installed on instance %s", config.SSMInstanceID)
	}
	log.Info("AWS CLI is installed on remote instance")

	log.Info("Downloading from S3 to remote instance...")
	var downloadCommand string
	if config.IsDirectory {
		downloadCommand = fmt.Sprintf("aws s3 cp %s %s --recursive", s3URL, config.Destination)
	} else {
		downloadCommand = fmt.Sprintf("aws s3 cp %s %s", s3URL, config.Destination)
	}
	if err := retryOperation(func() error {
		_, err := ssm.RunCommand(ssmConnection.Client, config.SSMInstanceID, []string{downloadCommand})
		return err
	}, config.MaxRetries, config.RetryDelay); err != nil {
		return fmt.Errorf("failed to download from S3 to remote instance: %w", err)
	}
	log.Info("Download to remote instance completed successfully")

	log.Info("Verifying files on remote instance...")
	confirmCommand := fmt.Sprintf("ls -la %s", config.Destination)
	output, err := ssm.RunCommand(ssmConnection.Client, config.SSMInstanceID, []string{confirmCommand})
	if err != nil {
		log.Warn("Failed to verify files on remote instance: %v", err)
	} else {
		log.Debug("Remote directory contents:\n%s", output)
	}

	log.Info("File transfer completed successfully!")
	return nil
}

func retryOperation(operation func() error, maxRetries int, baseDelay int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(baseDelay*(1<<uint(attempt-1))) * time.Second
			log.Warn("Retry attempt %d/%d after %v...", attempt, maxRetries, delay)
			time.Sleep(delay)
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				log.Info("Operation succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err

		if !isRetryableError(err) {
			log.Error("Non-retryable error encountered: %v", err)
			return err
		}

		log.Warn("Operation failed (attempt %d/%d): %v", attempt+1, maxRetries+1, err)
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if awsErr, ok := err.(awserr.Error); ok {
		// Check for non-retryable permission/authentication errors first
		switch awsErr.Code() {
		case "AccessDenied", "AccessDeniedException", "UnauthorizedAccess",
			"Forbidden", "InvalidAccessKeyId", "SignatureDoesNotMatch",
			"UnrecognizedClientException", "InvalidClientTokenId",
			"ExpiredToken", "ExpiredTokenException", "InvalidToken":
			return false
		}

		// Check for retryable errors
		switch awsErr.Code() {
		case "RequestTimeout", "ServiceUnavailable", "ThrottlingException",
			"RequestLimitExceeded", "TooManyRequestsException", "InternalError",
			"RequestThrottled", "Throttling":
			return true
		}
	}

	errStr := err.Error()

	// Check for non-retryable permission strings
	nonRetryableStrings := []string{
		"access denied",
		"unauthorized",
		"forbidden",
		"invalid credentials",
		"permission denied",
		"not authorized",
	}

	for _, nonRetryable := range nonRetryableStrings {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(nonRetryable)) {
			return false
		}
	}

	// Check for retryable errors
	retryableStrings := []string{
		"connection reset",
		"connection refused",
		"timeout",
		"temporary failure",
		"TLS handshake timeout",
		"EOF",
		"i/o timeout",
	}

	for _, retryable := range retryableStrings {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(retryable)) {
			return true
		}
	}

	return false
}
