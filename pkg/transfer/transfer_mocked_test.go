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
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/cowdogmoo/bcp/pkg/model"
)

// Mock S3 client
type mockS3Client struct {
	putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

// Mock SSM client
type mockSSMClient struct {
	sendCommandFunc          func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error)
	getCommandInvocationFunc func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error)
}

func (m *mockSSMClient) SendCommand(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
	if m.sendCommandFunc != nil {
		return m.sendCommandFunc(ctx, params, optFns...)
	}
	return &ssm.SendCommandOutput{
		Command: &types.Command{
			CommandId: aws.String("test-command-id"),
		},
	}, nil
}

func (m *mockSSMClient) GetCommandInvocation(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
	if m.getCommandInvocationFunc != nil {
		return m.getCommandInvocationFunc(ctx, params, optFns...)
	}
	return &ssm.GetCommandInvocationOutput{
		Status:                types.CommandInvocationStatusSuccess,
		StandardOutputContent: aws.String("success"),
	}, nil
}

func TestExecuteWithClients_Success(t *testing.T) {
	// Create temp directory and file
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("/usr/bin/aws"),
			}, nil
		},
	}

	config := model.TransferConfig{
		Source:        testFile,
		SSMInstanceID: "i-1234567890abcdef0",
		Destination:   "/tmp/test.txt",
		BucketName:    "test-bucket",
		MaxRetries:    3,
		RetryDelay:    1,
		IsDirectory:   false,
	}

	ctx := context.Background()
	err = ExecuteWithClients(ctx, config, mockS3, mockSSM)
	if err != nil {
		t.Errorf("ExecuteWithClients() error = %v", err)
	}
}

func TestExecuteWithClients_S3UploadFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, errors.New("S3 upload failed")
		},
	}

	mockSSM := &mockSSMClient{}

	config := model.TransferConfig{
		Source:        testFile,
		SSMInstanceID: "i-1234567890abcdef0",
		Destination:   "/tmp/test.txt",
		BucketName:    "test-bucket",
		MaxRetries:    1,
		RetryDelay:    1,
		IsDirectory:   false,
	}

	ctx := context.Background()
	err = ExecuteWithClients(ctx, config, mockS3, mockSSM)
	if err == nil {
		t.Error("Expected error from S3 upload failure")
	}
}

func TestExecuteWithClients_AWSCLINotInstalled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{}

	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusFailed,
				StandardErrorContent:  aws.String("command not found"),
				StandardOutputContent: aws.String(""),
			}, nil
		},
	}

	config := model.TransferConfig{
		Source:        testFile,
		SSMInstanceID: "i-1234567890abcdef0",
		Destination:   "/tmp/test.txt",
		BucketName:    "test-bucket",
		MaxRetries:    1,
		RetryDelay:    1,
		IsDirectory:   false,
	}

	ctx := context.Background()
	err = ExecuteWithClients(ctx, config, mockS3, mockSSM)
	if err == nil {
		t.Error("Expected error when AWS CLI not installed")
	}
}

func TestExecuteWithClients_SSMDownloadFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{}

	callCount := 0
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			callCount++
			// First call (AWS CLI check) succeeds
			if callCount == 1 {
				return &ssm.GetCommandInvocationOutput{
					Status:                types.CommandInvocationStatusSuccess,
					StandardOutputContent: aws.String("/usr/bin/aws"),
				}, nil
			}
			// Second call (download) fails
			return &ssm.GetCommandInvocationOutput{
				Status:               types.CommandInvocationStatusFailed,
				StandardErrorContent: aws.String("download failed"),
			}, nil
		},
	}

	config := model.TransferConfig{
		Source:        testFile,
		SSMInstanceID: "i-1234567890abcdef0",
		Destination:   "/tmp/test.txt",
		BucketName:    "test-bucket",
		MaxRetries:    1,
		RetryDelay:    1,
		IsDirectory:   false,
	}

	ctx := context.Background()
	err = ExecuteWithClients(ctx, config, mockS3, mockSSM)
	if err == nil {
		t.Error("Expected error from SSM download failure")
	}
}

func TestUploadFile_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			if aws.ToString(params.Bucket) != "test-bucket" {
				t.Errorf("Expected bucket 'test-bucket', got %s", aws.ToString(params.Bucket))
			}
			if aws.ToString(params.Key) != "test.txt" {
				t.Errorf("Expected key 'test.txt', got %s", aws.ToString(params.Key))
			}
			return &s3.PutObjectOutput{}, nil
		},
	}

	ctx := context.Background()
	err = uploadFile(ctx, mockS3, "test-bucket", testFile, "test.txt")
	if err != nil {
		t.Errorf("uploadFile() error = %v", err)
	}
}

func TestUploadFile_OpenError(t *testing.T) {
	mockS3 := &mockS3Client{}

	ctx := context.Background()
	err := uploadFile(ctx, mockS3, "test-bucket", "/nonexistent/file.txt", "file.txt")
	if err == nil {
		t.Error("Expected error when file doesn't exist")
	}
}

func TestUploadFile_S3Error(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, errors.New("S3 error")
		},
	}

	ctx := context.Background()
	err = uploadFile(ctx, mockS3, "test-bucket", testFile, "test.txt")
	if err == nil {
		t.Error("Expected error from S3 upload")
	}
}

func TestUploadDirectory_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectory and files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	uploadCount := 0
	mockS3 := &mockS3Client{
		putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			uploadCount++
			return &s3.PutObjectOutput{}, nil
		},
	}

	ctx := context.Background()
	err = uploadDirectory(ctx, mockS3, "test-bucket", tmpDir)
	if err != nil {
		t.Errorf("uploadDirectory() error = %v", err)
	}

	if uploadCount != 2 {
		t.Errorf("Expected 2 files uploaded, got %d", uploadCount)
	}
}

func TestUploadToS3_File(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{}

	ctx := context.Background()
	err = uploadToS3(ctx, mockS3, "test-bucket", testFile)
	if err != nil {
		t.Errorf("uploadToS3() error = %v", err)
	}
}

func TestUploadToS3_Directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockS3 := &mockS3Client{}

	ctx := context.Background()
	err = uploadToS3(ctx, mockS3, "test-bucket", tmpDir)
	if err != nil {
		t.Errorf("uploadToS3() error = %v", err)
	}
}

func TestUploadToS3_NonExistent(t *testing.T) {
	mockS3 := &mockS3Client{}

	ctx := context.Background()
	err := uploadToS3(ctx, mockS3, "test-bucket", "/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestRunSSMCommand_Success(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			if params.InstanceIds[0] != "i-1234567890abcdef0" {
				t.Errorf("Expected instance ID 'i-1234567890abcdef0', got %s", params.InstanceIds[0])
			}
			if aws.ToString(params.DocumentName) != "AWS-RunShellScript" {
				t.Errorf("Expected document 'AWS-RunShellScript', got %s", aws.ToString(params.DocumentName))
			}
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("command output"),
			}, nil
		},
	}

	ctx := context.Background()
	output, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err != nil {
		t.Errorf("runSSMCommand() error = %v", err)
	}
	if output != "command output" {
		t.Errorf("Expected output 'command output', got %s", output)
	}
}

func TestRunSSMCommand_SendCommandError(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return nil, errors.New("send command failed")
		},
	}

	ctx := context.Background()
	_, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err == nil {
		t.Error("Expected error from SendCommand")
	}
}

func TestRunSSMCommand_Failed(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:               types.CommandInvocationStatusFailed,
				StandardErrorContent: aws.String("command failed"),
			}, nil
		},
	}

	ctx := context.Background()
	_, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err == nil {
		t.Error("Expected error from failed command")
	}
}

func TestRunSSMCommand_Cancelled(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:               types.CommandInvocationStatusCancelled,
				StandardErrorContent: aws.String("command cancelled"),
			}, nil
		},
	}

	ctx := context.Background()
	_, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err == nil {
		t.Error("Expected error from cancelled command")
	}
}

func TestRunSSMCommand_TimedOut(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:               types.CommandInvocationStatusTimedOut,
				StandardErrorContent: aws.String("command timed out"),
			}, nil
		},
	}

	ctx := context.Background()
	_, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err == nil {
		t.Error("Expected error from timed out command")
	}
}

func TestRunSSMCommand_GetCommandInvocationError(t *testing.T) {
	callCount := 0
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			callCount++
			// Return error for first few calls, then success
			if callCount < 3 {
				return nil, errors.New("GetCommandInvocation error")
			}
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("output"),
			}, nil
		},
	}

	ctx := context.Background()
	output, err := runSSMCommand(ctx, mockSSM, "i-1234567890abcdef0", []string{"echo test"})
	if err != nil {
		t.Errorf("runSSMCommand() error = %v", err)
	}
	if output != "output" {
		t.Errorf("Expected output 'output', got %s", output)
	}
	if callCount < 3 {
		t.Errorf("Expected at least 3 calls to GetCommandInvocation, got %d", callCount)
	}
}

func TestCheckAWSCLIInstalled_Installed(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("/usr/bin/aws"),
			}, nil
		},
	}

	ctx := context.Background()
	installed, err := checkAWSCLIInstalled(ctx, mockSSM, "i-1234567890abcdef0")
	if err != nil {
		t.Errorf("checkAWSCLIInstalled() error = %v", err)
	}
	if !installed {
		t.Error("Expected AWS CLI to be installed")
	}
}

func TestCheckAWSCLIInstalled_NotInstalled(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:               types.CommandInvocationStatusFailed,
				StandardErrorContent: aws.String("command failed"),
			}, nil
		},
	}

	ctx := context.Background()
	installed, err := checkAWSCLIInstalled(ctx, mockSSM, "i-1234567890abcdef0")
	if err != nil {
		t.Errorf("checkAWSCLIInstalled() error = %v", err)
	}
	if installed {
		t.Error("Expected AWS CLI to not be installed")
	}
}

func TestCheckAWSCLIInstalled_Error(t *testing.T) {
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return nil, errors.New("SSM error")
		},
	}

	ctx := context.Background()
	_, err := checkAWSCLIInstalled(ctx, mockSSM, "i-1234567890abcdef0")
	if err == nil {
		t.Error("Expected error from SSM")
	}
}

func TestExecuteWithClients_DirectoryUpload(t *testing.T) {
	// Create temp directory with files
	tmpDir, err := os.MkdirTemp("", "bcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	mockS3 := &mockS3Client{}
	mockSSM := &mockSSMClient{
		sendCommandFunc: func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
			return &ssm.SendCommandOutput{
				Command: &types.Command{
					CommandId: aws.String("test-command-id"),
				},
			}, nil
		},
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				Status:                types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("/usr/bin/aws"),
			}, nil
		},
	}

	config := model.TransferConfig{
		Source:        tmpDir,
		SSMInstanceID: "i-1234567890abcdef0",
		Destination:   "/tmp/testdir",
		BucketName:    "test-bucket",
		MaxRetries:    3,
		RetryDelay:    1,
		IsDirectory:   true,
	}

	ctx := context.Background()
	err = ExecuteWithClients(ctx, config, mockS3, mockSSM)
	if err != nil {
		t.Errorf("ExecuteWithClients() with directory error = %v", err)
	}
}

func TestIsRetryableError_NonRetryablePermissions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AccessDenied",
			err:      newMockAPIError("AccessDenied", "access denied"),
			expected: false,
		},
		{
			name:     "AccessDeniedException",
			err:      newMockAPIError("AccessDeniedException", "access denied exception"),
			expected: false,
		},
		{
			name:     "UnauthorizedAccess",
			err:      newMockAPIError("UnauthorizedAccess", "unauthorized access"),
			expected: false,
		},
		{
			name:     "Forbidden",
			err:      newMockAPIError("Forbidden", "forbidden"),
			expected: false,
		},
		{
			name:     "InvalidAccessKeyId",
			err:      newMockAPIError("InvalidAccessKeyId", "invalid access key"),
			expected: false,
		},
		{
			name:     "SignatureDoesNotMatch",
			err:      newMockAPIError("SignatureDoesNotMatch", "signature does not match"),
			expected: false,
		},
		{
			name:     "UnrecognizedClientException",
			err:      newMockAPIError("UnrecognizedClientException", "unrecognized client"),
			expected: false,
		},
		{
			name:     "InvalidClientTokenId",
			err:      newMockAPIError("InvalidClientTokenId", "invalid client token"),
			expected: false,
		},
		{
			name:     "ExpiredToken",
			err:      newMockAPIError("ExpiredToken", "expired token"),
			expected: false,
		},
		{
			name:     "ExpiredTokenException",
			err:      newMockAPIError("ExpiredTokenException", "expired token exception"),
			expected: false,
		},
		{
			name:     "InvalidToken",
			err:      newMockAPIError("InvalidToken", "invalid token"),
			expected: false,
		},
		{
			name:     "access denied string",
			err:      errors.New("access denied"),
			expected: false,
		},
		{
			name:     "unauthorized string",
			err:      errors.New("unauthorized"),
			expected: false,
		},
		{
			name:     "forbidden string",
			err:      errors.New("forbidden"),
			expected: false,
		},
		{
			name:     "invalid credentials string",
			err:      errors.New("invalid credentials"),
			expected: false,
		},
		{
			name:     "permission denied string",
			err:      errors.New("permission denied"),
			expected: false,
		},
		{
			name:     "not authorized string",
			err:      errors.New("not authorized"),
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
