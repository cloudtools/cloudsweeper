package cloud

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type baseSnapshot struct {
	baseResource
	encrypted bool
	inUse     bool
	sizeGB    int64
}

func (s *baseSnapshot) Encrypted() bool {
	return s.encrypted
}

func (s *baseSnapshot) InUse() bool {
	return s.inUse
}

func (s *baseSnapshot) SizeGB() int64 {
	return s.sizeGB
}

type awsSnapshot struct {
	baseSnapshot
}

func (s *awsSnapshot) Cleanup() error {
	log.Println("Cleaning up snapshot", s.ID())
	client := clientForAWSResource(s)
	input := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(s.ID()),
	}
	_, err := client.DeleteSnapshot(input)
	return err
}

func (s *awsSnapshot) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(s, key, value, overwrite)
}
