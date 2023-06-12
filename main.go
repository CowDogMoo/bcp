package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/l50/awsutils/s3"
	"github.com/l50/awsutils/ssm"

	ssmutils "github.com/l50/awsutils/ssm"
)

func main() {
	// get the input arguments
	args := os.Args
	if len(args) != 3 {
		log.Fatalf("Incorrect usage. Use it like this: `bcp mydirectory/ $SSM_INSTANCE_ID:~/mydirectory`")
	}

	sourceDirectory := args[1]
	ssmPath := args[2]
	split := strings.Split(ssmPath, ":")
	ssmInstanceID := split[0]
	destinationDirectory := split[1]
	fmt.Println(destinationDirectory)

	// create S3 and SSM connections
	s3Connection := s3.CreateConnection()
	ssmConnection := ssm.CreateConnection()

	bucketName := "0ec895b739" // assuming bucket already exists, if not, create one
	uploadFP := sourceDirectory

	// Upload the file to S3
	if err := s3.UploadBucketDir(s3Connection.Session, bucketName, uploadFP); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println("AWS Error Code: ", aerr.Code())
			fmt.Println("Error Message: ", aerr.Message())
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	// Download the file from S3 via SSM to the remote instance
	s3URL := fmt.Sprintf("s3://%s/%s", bucketName, uploadFP)
	// check if AWS CLI is installed on the instance
	awsCLICheck, err := ssmutils.CheckAWSCLIInstalled(ssmConnection.Client, ssmInstanceID)
	if err != nil || !awsCLICheck {
		if err != nil {
			log.Fatalf("Unable to check if AWS CLI is installed: %v", err)
		} else {
			log.Fatalf("AWS CLI is not installed on the instance")
		}
	}
	downloadCommand := fmt.Sprintf("aws s3 cp %s %s --recursive", s3URL, destinationDirectory)
	if _, err := ssm.RunCommand(ssmConnection.Client, ssmInstanceID, []string{downloadCommand}); err != nil {
		log.Fatalf("Unable to run command via SSM: %v", err)
	}

	// Confirm that the upload has been copied successfully to the instance
	confirmCommand := fmt.Sprintf("ls %s", destinationDirectory)
	if _, err := ssm.RunCommand(ssmConnection.Client, ssmInstanceID, []string{confirmCommand}); err != nil {
		log.Fatalf("Unable to run command via SSM: %v", err)
	}

	log.Println("File copied successfully!")
}
