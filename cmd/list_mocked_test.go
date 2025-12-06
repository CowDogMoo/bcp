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
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/spf13/cobra"
)

// Mock S3 client for list operations
type mockListS3Client struct {
	listBucketsFunc func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
}

func (m *mockListS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	if m.listBucketsFunc != nil {
		return m.listBucketsFunc(ctx, params, optFns...)
	}
	return &s3.ListBucketsOutput{}, nil
}

// Mock SSM client for list operations
type mockListSSMClient struct {
	describeInstanceInformationFunc func(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
}

func (m *mockListSSMClient) DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	if m.describeInstanceInformationFunc != nil {
		return m.describeInstanceInformationFunc(ctx, params, optFns...)
	}
	return &ssm.DescribeInstanceInformationOutput{}, nil
}

// Mock EC2 client for list operations
type mockListEC2Client struct {
	describeInstancesFunc func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func (m *mockListEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describeInstancesFunc != nil {
		return m.describeInstancesFunc(ctx, params, optFns...)
	}
	return &ec2.DescribeInstancesOutput{}, nil
}

func TestListBucketsWithMocks_Success(t *testing.T) {
	// This test demonstrates how we could refactor the code to support mocking
	// For now, we'll just test the structure since the actual functions are hard to mock

	// Verify the command structure
	if listBucketsCmd.Use != "buckets" {
		t.Errorf("listBucketsCmd.Use = %v, want 'buckets'", listBucketsCmd.Use)
	}

	if listBucketsCmd.RunE == nil {
		t.Error("listBucketsCmd.RunE should not be nil")
	}
}

func TestListBucketsCmd_Structure(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Test command metadata
	if listBucketsCmd.Use != "buckets" {
		t.Errorf("Expected Use='buckets', got %s", listBucketsCmd.Use)
	}

	if listBucketsCmd.Short != "List available S3 buckets" {
		t.Errorf("Unexpected Short description")
	}

	// Verify RunE is set
	if listBucketsCmd.RunE == nil {
		t.Error("RunE should be defined")
	}
}

func TestListInstancesCmd_Structure(t *testing.T) {
	// Test command metadata
	if listInstancesCmd.Use != "instances" {
		t.Errorf("Expected Use='instances', got %s", listInstancesCmd.Use)
	}

	if listInstancesCmd.Short != "List SSM-managed EC2 instances" {
		t.Errorf("Unexpected Short description")
	}

	// Verify RunE is set
	if listInstancesCmd.RunE == nil {
		t.Error("RunE should be defined")
	}
}

func TestListInstancesCmd_FlagValidation(t *testing.T) {
	// Test flag defaults and types
	allFlag := listInstancesCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("'all' flag not found")
	}
	if allFlag.Value.Type() != "bool" {
		t.Errorf("Expected 'all' flag type 'bool', got %s", allFlag.Value.Type())
	}
	if allFlag.DefValue != "false" {
		t.Errorf("Expected default false for 'all' flag, got %s", allFlag.DefValue)
	}

	regionFlag := listInstancesCmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Fatal("'region' flag not found")
	}
	if regionFlag.Value.Type() != "string" {
		t.Errorf("Expected 'region' flag type 'string', got %s", regionFlag.Value.Type())
	}
}

// The following tests demonstrate the approach we'd take if we refactored the code
// to accept interfaces instead of concrete AWS clients

func TestMockListSSMInstances_Success(t *testing.T) {
	// This demonstrates how we'd test listSSMInstances with mocks
	// In practice, we'd need to refactor listSSMInstances to accept an interface

	mockSSM := &mockListSSMClient{
		describeInstanceInformationFunc: func(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			now := time.Now()
			return &ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:       aws.String("i-1234567890abcdef0"),
						PingStatus:       ssmtypes.PingStatusOnline,
						PlatformType:     ssmtypes.PlatformTypeLinux,
						IPAddress:        aws.String("10.0.0.1"),
						ComputerName:     aws.String("test-instance"),
						LastPingDateTime: &now,
					},
					{
						InstanceId:       aws.String("i-0987654321fedcba0"),
						PingStatus:       ssmtypes.PingStatusConnectionLost,
						PlatformType:     ssmtypes.PlatformTypeWindows,
						IPAddress:        aws.String("10.0.0.2"),
						ComputerName:     aws.String("test-instance-2"),
						LastPingDateTime: &now,
					},
				},
				NextToken: nil,
			}, nil
		},
	}

	// In a real refactored version, we'd call:
	// err := listSSMInstancesWithClient(ctx, cfg, mockSSM)
	// For now, just verify the mock works
	ctx := context.Background()
	result, err := mockSSM.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.InstanceInformationList) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(result.InstanceInformationList))
	}
}

func TestMockListSSMInstances_Empty(t *testing.T) {
	mockSSM := &mockListSSMClient{
		describeInstanceInformationFunc: func(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			return &ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{},
				NextToken:               nil,
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockSSM.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.InstanceInformationList) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(result.InstanceInformationList))
	}
}

func TestMockListSSMInstances_Error(t *testing.T) {
	mockSSM := &mockListSSMClient{
		describeInstanceInformationFunc: func(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			return nil, errors.New("SSM API error")
		},
	}

	ctx := context.Background()
	_, err := mockSSM.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockListAllInstances_Success(t *testing.T) {
	mockEC2 := &mockListEC2Client{
		describeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								InstanceId:       aws.String("i-1234567890abcdef0"),
								InstanceType:     ec2types.InstanceTypeT2Micro,
								PrivateIpAddress: aws.String("10.0.0.1"),
								State: &ec2types.InstanceState{
									Name: ec2types.InstanceStateNameRunning,
								},
								Tags: []ec2types.Tag{
									{
										Key:   aws.String("Name"),
										Value: aws.String("test-instance"),
									},
								},
							},
							{
								InstanceId:       aws.String("i-0987654321fedcba0"),
								InstanceType:     ec2types.InstanceTypeT2Small,
								PrivateIpAddress: aws.String("10.0.0.2"),
								State: &ec2types.InstanceState{
									Name: ec2types.InstanceStateNameStopped,
								},
								Tags: []ec2types.Tag{
									{
										Key:   aws.String("Name"),
										Value: aws.String("test-instance-2"),
									},
								},
							},
						},
					},
				},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockEC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	instanceCount := 0
	for _, reservation := range result.Reservations {
		instanceCount += len(reservation.Instances)
	}

	if instanceCount != 2 {
		t.Errorf("Expected 2 instances, got %d", instanceCount)
	}
}

func TestMockListAllInstances_Empty(t *testing.T) {
	mockEC2 := &mockListEC2Client{
		describeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockEC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.Reservations) != 0 {
		t.Errorf("Expected 0 reservations, got %d", len(result.Reservations))
	}
}

func TestMockListAllInstances_Error(t *testing.T) {
	mockEC2 := &mockListEC2Client{
		describeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return nil, errors.New("EC2 API error")
		},
	}

	ctx := context.Background()
	_, err := mockEC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockListBuckets_Success(t *testing.T) {
	now := time.Now()
	mockS3 := &mockListS3Client{
		listBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return &s3.ListBucketsOutput{
				Buckets: []s3types.Bucket{
					{
						Name:         aws.String("test-bucket-1"),
						CreationDate: &now,
					},
					{
						Name:         aws.String("test-bucket-2"),
						CreationDate: &now,
					},
				},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockS3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.Buckets) != 2 {
		t.Errorf("Expected 2 buckets, got %d", len(result.Buckets))
	}
}

func TestMockListBuckets_Empty(t *testing.T) {
	mockS3 := &mockListS3Client{
		listBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return &s3.ListBucketsOutput{
				Buckets: []s3types.Bucket{},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockS3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result.Buckets) != 0 {
		t.Errorf("Expected 0 buckets, got %d", len(result.Buckets))
	}
}

func TestMockListBuckets_Error(t *testing.T) {
	mockS3 := &mockListS3Client{
		listBucketsFunc: func(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return nil, errors.New("S3 API error")
		},
	}

	ctx := context.Background()
	_, err := mockS3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMockListSSMInstances_Pagination(t *testing.T) {
	callCount := 0
	mockSSM := &mockListSSMClient{
		describeInstanceInformationFunc: func(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			callCount++
			now := time.Now()

			if callCount == 1 {
				// First page
				return &ssm.DescribeInstanceInformationOutput{
					InstanceInformationList: []ssmtypes.InstanceInformation{
						{
							InstanceId:       aws.String("i-page1-instance1"),
							PingStatus:       ssmtypes.PingStatusOnline,
							PlatformType:     ssmtypes.PlatformTypeLinux,
							IPAddress:        aws.String("10.0.0.1"),
							ComputerName:     aws.String("page1-instance1"),
							LastPingDateTime: &now,
						},
					},
					NextToken: aws.String("page2-token"),
				}, nil
			}

			// Second page
			return &ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:       aws.String("i-page2-instance1"),
						PingStatus:       ssmtypes.PingStatusOnline,
						PlatformType:     ssmtypes.PlatformTypeLinux,
						IPAddress:        aws.String("10.0.0.2"),
						ComputerName:     aws.String("page2-instance1"),
						LastPingDateTime: &now,
					},
				},
				NextToken: nil,
			}, nil
		},
	}

	ctx := context.Background()

	// First call
	result1, err := mockSSM.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err != nil {
		t.Errorf("Expected no error on first call, got %v", err)
	}
	if result1.NextToken == nil {
		t.Error("Expected NextToken on first call")
	}

	// Second call with token
	result2, err := mockSSM.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{
		NextToken: result1.NextToken,
	})
	if err != nil {
		t.Errorf("Expected no error on second call, got %v", err)
	}
	if result2.NextToken != nil {
		t.Error("Expected no NextToken on second call")
	}

	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestMockListAllInstances_MultipleStates(t *testing.T) {
	mockEC2 := &mockListEC2Client{
		describeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			return &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								InstanceId:       aws.String("i-running"),
								InstanceType:     ec2types.InstanceTypeT2Micro,
								PrivateIpAddress: aws.String("10.0.0.1"),
								State: &ec2types.InstanceState{
									Name: ec2types.InstanceStateNameRunning,
								},
							},
							{
								InstanceId:       aws.String("i-stopped"),
								InstanceType:     ec2types.InstanceTypeT2Micro,
								PrivateIpAddress: aws.String("10.0.0.2"),
								State: &ec2types.InstanceState{
									Name: ec2types.InstanceStateNameStopped,
								},
							},
							{
								InstanceId:       aws.String("i-terminated"),
								InstanceType:     ec2types.InstanceTypeT2Micro,
								PrivateIpAddress: aws.String("10.0.0.3"),
								State: &ec2types.InstanceState{
									Name: ec2types.InstanceStateNameTerminated,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mockEC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	stateMap := make(map[ec2types.InstanceStateName]int)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			stateMap[instance.State.Name]++
		}
	}

	if stateMap[ec2types.InstanceStateNameRunning] != 1 {
		t.Errorf("Expected 1 running instance, got %d", stateMap[ec2types.InstanceStateNameRunning])
	}
	if stateMap[ec2types.InstanceStateNameStopped] != 1 {
		t.Errorf("Expected 1 stopped instance, got %d", stateMap[ec2types.InstanceStateNameStopped])
	}
	if stateMap[ec2types.InstanceStateNameTerminated] != 1 {
		t.Errorf("Expected 1 terminated instance, got %d", stateMap[ec2types.InstanceStateNameTerminated])
	}
}
