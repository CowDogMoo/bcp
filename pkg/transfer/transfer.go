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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go"
	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/cowdogmoo/bcp/pkg/model"
)

func Execute(transferConfig model.TransferConfig) error {
	log.Info("Starting transfer from %s to %s:%s", transferConfig.Source, transferConfig.SSMInstanceID, transferConfig.Destination)

	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	ssmClient := ssm.NewFromConfig(cfg)

	uploadPath := strings.TrimPrefix(transferConfig.Source, "./")
	s3URL := fmt.Sprintf("s3://%s/%s", transferConfig.BucketName, uploadPath)

	log.Info("Uploading %s to S3 bucket %s...", transferConfig.Source, transferConfig.BucketName)
	if err := retryOperation(func() error {
		return uploadToS3(ctx, s3Client, transferConfig.BucketName, uploadPath)
	}, transferConfig.MaxRetries, transferConfig.RetryDelay); err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	log.Info("Upload to S3 completed successfully")

	log.Info("Checking if AWS CLI is installed on instance %s...", transferConfig.SSMInstanceID)
	awsCLICheck, err := checkAWSCLIInstalled(ctx, ssmClient, transferConfig.SSMInstanceID)
	if err != nil {
		return fmt.Errorf("failed to check AWS CLI installation: %w", err)
	}
	if !awsCLICheck {
		return fmt.Errorf("AWS CLI is not installed on instance %s", transferConfig.SSMInstanceID)
	}
	log.Info("AWS CLI is installed on remote instance")

	log.Info("Downloading from S3 to remote instance...")
	var downloadCommand string
	if transferConfig.IsDirectory {
		downloadCommand = fmt.Sprintf("aws s3 cp %s %s --recursive", s3URL, transferConfig.Destination)
	} else {
		downloadCommand = fmt.Sprintf("aws s3 cp %s %s", s3URL, transferConfig.Destination)
	}
	if err := retryOperation(func() error {
		_, err := runSSMCommand(ctx, ssmClient, transferConfig.SSMInstanceID, []string{downloadCommand})
		return err
	}, transferConfig.MaxRetries, transferConfig.RetryDelay); err != nil {
		return fmt.Errorf("failed to download from S3 to remote instance: %w", err)
	}
	log.Info("Download to remote instance completed successfully")

	log.Info("Verifying files on remote instance...")
	confirmCommand := fmt.Sprintf("ls -la %s", transferConfig.Destination)
	output, err := runSSMCommand(ctx, ssmClient, transferConfig.SSMInstanceID, []string{confirmCommand})
	if err != nil {
		log.Warn("Failed to verify files on remote instance: %v", err)
	} else {
		log.Debug("Remote directory contents:\n%s", output)
	}

	log.Info("File transfer completed successfully!")
	return nil
}

// uploadToS3 uploads a file or directory to S3
func uploadToS3(ctx context.Context, client *s3.Client, bucketName, localPath string) error {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", localPath, err)
	}

	if fileInfo.IsDir() {
		return uploadDirectory(ctx, client, bucketName, localPath)
	}
	return uploadFile(ctx, client, bucketName, localPath, filepath.Base(localPath))
}

// uploadDirectory recursively uploads a directory to S3
func uploadDirectory(ctx context.Context, client *s3.Client, bucketName, localPath string) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate the relative path for the S3 key
		relPath, err := filepath.Rel(filepath.Dir(localPath), path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Convert Windows paths to Unix-style for S3
		s3Key := filepath.ToSlash(relPath)

		return uploadFile(ctx, client, bucketName, path, s3Key)
	})
}

// uploadFile uploads a single file to S3
func uploadFile(ctx context.Context, client *s3.Client, bucketName, filePath, s3Key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Warn("Failed to close file %s: %v", filePath, closeErr)
		}
	}()

	log.Debug("Uploading %s to s3://%s/%s", filePath, bucketName, s3Key)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s to S3: %w", filePath, err)
	}

	return nil
}

// runSSMCommand executes a command on an EC2 instance via SSM and waits for completion
func runSSMCommand(ctx context.Context, client *ssm.Client, instanceID string, commands []string) (string, error) {
	sendCommandInput := &ssm.SendCommandInput{
		InstanceIds:  []string{instanceID},
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]string{
			"commands": commands,
		},
	}

	result, err := client.SendCommand(ctx, sendCommandInput)
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	commandID := aws.ToString(result.Command.CommandId)

	// Wait for command to complete
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(2 * time.Second)

		invocationOutput, err := client.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		})
		if err != nil {
			continue
		}

		status := invocationOutput.Status
		if status == types.CommandInvocationStatusSuccess {
			return aws.ToString(invocationOutput.StandardOutputContent), nil
		} else if status == types.CommandInvocationStatusFailed ||
			status == types.CommandInvocationStatusCancelled ||
			status == types.CommandInvocationStatusTimedOut {
			stderr := aws.ToString(invocationOutput.StandardErrorContent)
			return "", fmt.Errorf("command failed with status %s: %s", status, stderr)
		}
	}

	return "", fmt.Errorf("command timed out waiting for completion")
}

// checkAWSCLIInstalled checks if AWS CLI is installed on an EC2 instance
func checkAWSCLIInstalled(ctx context.Context, client *ssm.Client, instanceID string) (bool, error) {
	output, err := runSSMCommand(ctx, client, instanceID, []string{"which aws"})
	if err != nil {
		// If the command fails, AWS CLI is not installed
		if strings.Contains(err.Error(), "command failed") {
			return false, nil
		}
		return false, err
	}

	// Check if output contains a valid path
	return strings.Contains(output, "/aws"), nil
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

	// Check for AWS SDK v2 smithy API errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// Check for non-retryable permission/authentication errors first
		switch apiErr.ErrorCode() {
		case "AccessDenied", "AccessDeniedException", "UnauthorizedAccess",
			"Forbidden", "InvalidAccessKeyId", "SignatureDoesNotMatch",
			"UnrecognizedClientException", "InvalidClientTokenId",
			"ExpiredToken", "ExpiredTokenException", "InvalidToken":
			return false
		}

		// Check for retryable errors
		switch apiErr.ErrorCode() {
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
