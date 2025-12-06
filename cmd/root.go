package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cowdogmoo/bcp/pkg/completion"
	"github.com/cowdogmoo/bcp/pkg/config"
	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/cowdogmoo/bcp/pkg/model"
	"github.com/cowdogmoo/bcp/pkg/transfer"
	"github.com/cowdogmoo/bcp/pkg/validation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucket  string
	cfgFile string
	verbose bool
	quiet   bool
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.bcp/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&bucket, "bucket", "b", "", "S3 bucket to use for the transfer (required)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output (debug level)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except errors")

	if err := rootCmd.RegisterFlagCompletionFunc("bucket", bucketCompletion); err != nil {
		log.Error("Failed to register bucket completion: %v", err)
	}

	if err := viper.BindPFlag("aws.bucket", rootCmd.PersistentFlags().Lookup("bucket")); err != nil {
		log.Error("Failed to bind bucket flag: %v", err)
	}
}

func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		log.Error("Failed to initialize config: %v", err)
		os.Exit(1)
	}

	if verbose {
		log.Init(config.GlobalConfig.Log.Format, "debug")
	} else if quiet {
		log.Init(config.GlobalConfig.Log.Format, "error")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bcp [sourceDirectory] [ssmPath]",
		Short: "bcp copies files/directories to an SSM instance via S3",
		Long: `bcp (Blob Copy) is a command-line tool that provides SCP-like functionality
for cloud systems using a blob store. It allows you to upload files to an
S3 bucket and download files from the bucket to a remote instance via
AWS Systems Manager (SSM).

Example:
  bcp ./my-files i-1234567890abcdef0:/home/ec2-user/files --bucket my-bucket`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: argsCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceDirectory := args[0]
			ssmPath := args[1]

			isDirectory, err := validation.ValidateSourcePath(sourceDirectory)
			if err != nil {
				return fmt.Errorf("invalid source path: %w", err)
			}

			ssmInstanceID, destinationDirectory, err := validation.ValidateSSMPath(ssmPath)
			if err != nil {
				return fmt.Errorf("invalid SSM path: %w", err)
			}

			bucketName := bucket
			if bucketName == "" {
				bucketName = config.GetBucket()
			}
			if bucketName == "" {
				return fmt.Errorf("bucket name is required (use --bucket flag or set in config)")
			}

			if err := validation.ValidateBucketName(bucketName); err != nil {
				return fmt.Errorf("invalid bucket name: %w", err)
			}

			transferConfig := model.TransferConfig{
				Source:        sourceDirectory,
				SSMInstanceID: ssmInstanceID,
				Destination:   destinationDirectory,
				BucketName:    bucketName,
				MaxRetries:    config.MaxRetries,
				RetryDelay:    config.RetryDelay,
				IsDirectory:   isDirectory,
			}

			if err := transfer.Execute(transferConfig); err != nil {
				return fmt.Errorf("transfer failed: %w", err)
			}

			return nil
		},
	}

	return rootCmd
}

var rootCmd = RootCmd()

func bucketCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	buckets, err := completion.GetBucketNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var matches []string
	for _, bucket := range buckets {
		if strings.HasPrefix(bucket, toComplete) {
			matches = append(matches, bucket)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

func instanceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	instances, err := completion.GetInstanceIDs()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var matches []string
	for _, instance := range instances {
		instanceID := strings.Split(instance, "\t")[0]
		if strings.HasPrefix(instanceID, toComplete) {
			matches = append(matches, instance)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

func argsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveDefault
	}

	if len(args) == 1 {
		if strings.Contains(toComplete, ":") {
			parts := strings.Split(toComplete, ":")
			if len(parts) == 2 {
				instanceID := parts[0]
				commonPaths := []string{
					instanceID + ":/tmp/",
					instanceID + ":/home/ec2-user/",
					instanceID + ":/opt/",
					instanceID + ":/usr/local/bin/",
					instanceID + ":/var/tmp/",
				}
				return commonPaths, cobra.ShellCompDirectiveNoSpace
			}
		}

		instances, err := completion.GetInstanceIDs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var matches []string
		for _, instance := range instances {
			instanceID := strings.Split(instance, "\t")[0]
			if strings.HasPrefix(instanceID, toComplete) {
				matches = append(matches, instanceID+":")
			}
		}

		return matches, cobra.ShellCompDirectiveNoSpace
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
