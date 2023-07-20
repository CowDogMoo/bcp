package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/l50/awsutils/s3"
	"github.com/l50/awsutils/ssm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bucket string

func init() {
	rootCmd.PersistentFlags().StringVarP(&bucket, "bucket", "b", "", "Bucket to use for the transfer")
	cobra.CheckErr(viper.BindPFlag("bucket", rootCmd.PersistentFlags().Lookup("bucket")))
}

// Execute adds child commands to the root
// command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// RootCmd represents the base command when called without any subcommands
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bcp [sourceDirectory] [ssmPath]",
		Short: "bcp copies a directory to an SSM instance via S3",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			sourceDirectory := args[0]
			// Check if source directory exists
			if _, err := os.Stat(sourceDirectory); os.IsNotExist(err) {
				cobra.CheckErr(err)
			}
			ssmPath := args[1]
			split := strings.Split(ssmPath, ":")
			ssmInstanceID := split[0]
			destinationDirectory := split[1]
			fmt.Println(destinationDirectory)

			// create S3 and SSM connections
			s3Connection := s3.CreateConnection()
			ssmConnection := ssm.CreateConnection()

			bucketName := bucket
			uploadFP := strings.TrimPrefix(sourceDirectory, "./")
			s3URL := fmt.Sprintf("s3://%s/%s", bucketName, uploadFP)

			if err := s3.UploadBucketDir(s3Connection.Session, bucketName, uploadFP); err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					fmt.Println("AWS Error Code:", aerr.Code())
					fmt.Println("Error Message:", aerr.Message())
				} else {
					fmt.Println(err.Error())
				}
				return
			}

			// Download the file from S3 via SSM to the remote instance
			awsCLICheck, err := ssm.CheckAWSCLIInstalled(ssmConnection.Client, ssmInstanceID)
			cobra.CheckErr(err)
			if !awsCLICheck {
				fmt.Println("AWS CLI is not installed on the instance")
				return
			}

			downloadCommand := fmt.Sprintf("aws s3 cp %s %s --recursive", s3URL, destinationDirectory)
			if _, err := ssm.RunCommand(ssmConnection.Client, ssmInstanceID, []string{downloadCommand}); err != nil {
				cobra.CheckErr(err)
			}

			confirmCommand := fmt.Sprintf("ls %s", destinationDirectory)
			if _, err := ssm.RunCommand(ssmConnection.Client, ssmInstanceID, []string{confirmCommand}); err != nil {
				cobra.CheckErr(err)
			}

			fmt.Println("File copied successfully!")
		},
	}

	return rootCmd
}

// Declare and initialize rootCmd outside the RootCmd function
var rootCmd = RootCmd()
