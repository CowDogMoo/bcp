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

package completion

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func GetBucketNames() ([]string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // S3 ListBuckets is global
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		buckets = append(buckets, aws.StringValue(bucket.Name))
	}

	return buckets, nil
}

func GetInstanceIDs() ([]string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := ssm.New(sess)

	input := &ssm.DescribeInstanceInformationInput{
		MaxResults: aws.Int64(50),
	}

	var instanceIDs []string

	err = svc.DescribeInstanceInformationPages(input,
		func(page *ssm.DescribeInstanceInformationOutput, lastPage bool) bool {
			for _, inst := range page.InstanceInformationList {
				if aws.StringValue(inst.PingStatus) == "Online" {
					instanceID := aws.StringValue(inst.InstanceId)
					name := aws.StringValue(inst.ComputerName)
					if name != "" {
						instanceIDs = append(instanceIDs, instanceID+"\t"+name)
					} else {
						instanceIDs = append(instanceIDs, instanceID)
					}
				}
			}
			return !lastPage
		})

	if err != nil {
		return nil, err
	}

	return instanceIDs, nil
}
