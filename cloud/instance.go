package cloud

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type baseInstance struct {
	baseResource
	instanceType string
}

func (i *baseInstance) InstanceType() string {
	return i.instanceType
}

type awsInstance struct {
	baseInstance
}

// Cleanup will termiante this instance
func (i *awsInstance) Cleanup() error {
	log.Println("Cleaning up instance", i.ID())
	client := clientForAWSResource(i)
	input := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{i.id}),
	}
	_, err := client.TerminateInstances(input)
	return err
}

func (i *awsInstance) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(i, key, value, overwrite)
}

func (i *awsInstance) RemoveTag(key string) error {
	return removeAWSTag(i, key)
}
