package cmd

import (
	"fmt"
	"os"

	"github.com/cowdogmoo/bcp/pkg/config"
	log "github.com/cowdogmoo/bcp/pkg/logging"
	"github.com/cowdogmoo/bcp/pkg/model"
	"github.com/cowdogmoo/bcp/pkg/transfer"
	"github.com/cowdogmoo/bcp/pkg/validation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucket    string
	cfgFile   string
	verbose   bool
	quiet     bool
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.bcp/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&bucket, "bucket", "b", "", "S3 bucket to use for the transfer (required)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output (debug level)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except errors")

	if err := viper.BindPFlag("aws.bucket", rootCmd.PersistentFlags().Lookup("bucket")); err != nil {
		log.Error("Failed to bind bucket flag: %v", err)
	}
}

// initConfig initializes the configuration
func initConfig() {
	if err := config.Init(cfgFile); err != nil {
		log.Error("Failed to initialize config: %v", err)
		os.Exit(1)
	}

	// Override log level based on flags
	if verbose {
		log.Init(config.GlobalConfig.Log.Format, "debug")
	} else if quiet {
		log.Init(config.GlobalConfig.Log.Format, "error")
	}
}

// Execute adds child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error("Command execution failed: %v", err)
		os.Exit(1)
	}
}

// RootCmd represents the base command when called without any subcommands
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
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceDirectory := args[0]
			ssmPath := args[1]

			// Validate source path
			if err := validation.ValidateSourcePath(sourceDirectory); err != nil {
				return fmt.Errorf("invalid source path: %w", err)
			}

			// Validate and parse SSM path
			ssmInstanceID, destinationDirectory, err := validation.ValidateSSMPath(ssmPath)
			if err != nil {
				return fmt.Errorf("invalid SSM path: %w", err)
			}

			// Get bucket name (from flag, config, or error)
			bucketName := bucket
			if bucketName == "" {
				bucketName = config.GetBucket()
			}
			if bucketName == "" {
				return fmt.Errorf("bucket name is required (use --bucket flag or set in config)")
			}

			// Validate bucket name
			if err := validation.ValidateBucketName(bucketName); err != nil {
				return fmt.Errorf("invalid bucket name: %w", err)
			}

			// Create transfer configuration
			transferConfig := model.TransferConfig{
				Source:        sourceDirectory,
				SSMInstanceID: ssmInstanceID,
				Destination:   destinationDirectory,
				BucketName:    bucketName,
				MaxRetries:    config.MaxRetries,
				RetryDelay:    config.RetryDelay,
			}

			// Execute the transfer
			if err := transfer.Execute(transferConfig); err != nil {
				return fmt.Errorf("transfer failed: %w", err)
			}

			return nil
		},
	}

	return rootCmd
}

// Declare and initialize rootCmd outside the RootCmd function
var rootCmd = RootCmd()
