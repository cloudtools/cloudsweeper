package cloud

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type baseImage struct {
	baseResource
	name   string
	sizeGB int64
}

func (i *baseImage) Name() string {
	return i.name
}

func (i *baseImage) SizeGB() int64 {
	return i.sizeGB
}

type awsImage struct {
	baseImage
}

func (i *awsImage) Cleanup() error {
	log.Println("Cleaning up image", i.ID())
	client := clientForAWSResource(i)
	input := &ec2.DeregisterImageInput{
		ImageId: aws.String(i.ID()),
	}
	_, err := client.DeregisterImage(input)
	return err
}

func (i *awsImage) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(i, key, value, overwrite)
}

func (i *awsImage) MakePrivate() error {
	log.Println("Making image private:", i.ID())
	if !i.Public() {
		// Image is already private
		return nil
	}
	client := clientForAWSResource(i)
	input := &ec2.ModifyImageAttributeInput{
		ImageId: aws.String(i.ID()),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: []*ec2.LaunchPermission{&ec2.LaunchPermission{
				Group: aws.String("all"),
			}},
		},
	}
	_, err := client.ModifyImageAttribute(input)
	if err != nil {
		return err
	}
	i.public = false
	return nil
}
