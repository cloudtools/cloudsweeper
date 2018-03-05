package cloud

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type baseVolume struct {
	baseResource
	sizeGB     int64
	attached   bool
	encrypted  bool
	volumeType string
}

func (v *baseVolume) SizeGB() int64 {
	return v.sizeGB
}

func (v *baseVolume) Attached() bool {
	return v.attached
}

func (v *baseVolume) Encrypted() bool {
	return v.encrypted
}

func (v *baseVolume) VolumeType() string {
	return v.volumeType
}

type awsVolume struct {
	baseVolume
}

func (v *awsVolume) Cleanup() error {
	log.Printf("Cleaning up volume %s in %s", v.ID(), v.Owner())
	client := clientForAWSResource(v)
	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(v.ID()),
	}
	_, err := client.DeleteVolume(input)
	return err
}

func (v *awsVolume) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(v, key, value, overwrite)
}

func (v *awsVolume) RemoveTag(key string) error {
	return removeAWSTag(v, key)
}
