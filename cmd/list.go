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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	listAll    bool
	listRegion string
)

func init() {
	listCmd.AddCommand(listBucketsCmd)
	listCmd.AddCommand(listInstancesCmd)

	listInstancesCmd.Flags().BoolVarP(&listAll, "all", "a", false, "list all instances (not just SSM-managed)")
	listInstancesCmd.Flags().StringVarP(&listRegion, "region", "r", "", "AWS region (defaults to config or AWS_REGION)")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS resources (buckets, instances)",
	Long: `List various AWS resources to help discover available targets for bcp.

Available subcommands:
  buckets    - List S3 buckets
  instances  - List SSM-managed EC2 instances`,
}

var listBucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "List available S3 buckets",
	Long: `List all S3 buckets in your AWS account.

Example:
  bcp list buckets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Fetching S3 buckets...")

		ctx := context.TODO()
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		svc := s3.NewFromConfig(cfg)
		result, err := svc.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err != nil {
			return fmt.Errorf("failed to list buckets: %w", err)
		}

		if len(result.Buckets) == 0 {
			log.Info("No S3 buckets found")
			return nil
		}

		log.Info("Found %d S3 bucket(s):", len(result.Buckets))
		fmt.Println("\nBucket Name                                      Created")
		fmt.Println("================================================ =========================")

		for _, bucket := range result.Buckets {
			bucketName := aws.ToString(bucket.Name)
			created := ""
			if bucket.CreationDate != nil {
				created = bucket.CreationDate.Format("2006-01-02 15:04:05 MST")
			}
			fmt.Printf("%-48s %s\n", bucketName, created)
		}

		return nil
	},
}

var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List SSM-managed EC2 instances",
	Long: `List EC2 instances that are managed by AWS Systems Manager.
By default, only shows instances with SSM agent running and ready.

Use --all to see all EC2 instances (including those without SSM).

Example:
  bcp list instances
  bcp list instances --all
  bcp list instances --region us-west-2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.TODO()

		// Determine region
		region := listRegion
		if region == "" {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to load AWS config: %w", err)
			}
			region = cfg.Region
			if region == "" {
				region = "us-east-1" // fallback default
			}
		}

		log.Info("Fetching instances in region: %s", region)

		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}

		if listAll {
			return listAllInstances(ctx, cfg)
		}
		return listSSMInstances(ctx, cfg)
	},
}

func listSSMInstances(ctx context.Context, cfg aws.Config) error {
	svc := ssm.NewFromConfig(cfg)

	input := &ssm.DescribeInstanceInformationInput{
		MaxResults: aws.Int32(50),
	}

	var instances []ssmtypes.InstanceInformation

	paginator := ssm.NewDescribeInstanceInformationPaginator(svc, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list SSM instances: %w", err)
		}
		instances = append(instances, page.InstanceInformationList...)
	}

	if len(instances) == 0 {
		log.Info("No SSM-managed instances found")
		fmt.Println("\nTip: Use --all to see all EC2 instances")
		return nil
	}

	log.Info("Found %d SSM-managed instance(s):", len(instances))
	fmt.Println("\nInstance ID          Status   Platform        IP Address      Name")
	fmt.Println("==================== ======== =============== =============== ================================")

	for _, inst := range instances {
		instanceID := aws.ToString(inst.InstanceId)
		status := string(inst.PingStatus)
		platform := string(inst.PlatformType)
		ipAddress := aws.ToString(inst.IPAddress)
		name := aws.ToString(inst.ComputerName)

		var statusStr string
		if status == "Online" {
			statusStr = "\033[32m" + status + "\033[0m" // green
		} else {
			statusStr = "\033[31m" + status + "\033[0m" // red
		}

		fmt.Printf("%-20s %-8s %-15s %-15s %s\n",
			instanceID, statusStr, platform, ipAddress, name)
	}

	return nil
}

func listAllInstances(ctx context.Context, cfg aws.Config) error {
	svc := ec2.NewFromConfig(cfg)

	result, err := svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return fmt.Errorf("failed to list EC2 instances: %w", err)
	}

	var instanceCount int
	var runningCount int

	log.Info("Fetching all EC2 instances...")

	fmt.Println("\nInstance ID          State    Type          IP Address      Name")
	fmt.Println("==================== ======== ============= =============== ================================")

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceCount++

			instanceID := aws.ToString(instance.InstanceId)
			state := string(instance.State.Name)
			instanceType := string(instance.InstanceType)
			privateIP := aws.ToString(instance.PrivateIpAddress)

			if state == "running" {
				runningCount++
			}

			var name string
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}

			var stateStr string
			switch state {
			case "running":
				stateStr = "\033[32m" + state + "\033[0m" // green
			case "stopped":
				stateStr = "\033[33m" + state + "\033[0m" // yellow
			default:
				stateStr = "\033[31m" + state + "\033[0m" // red
			}

			fmt.Printf("%-20s %-8s %-13s %-15s %s\n",
				instanceID, stateStr, instanceType, privateIP, name)
		}
	}

	log.Info("Total: %d instances (%d running)", instanceCount, runningCount)
	fmt.Println("\nTip: Use 'bcp list instances' (without --all) to see only SSM-managed instances")

	return nil
}
