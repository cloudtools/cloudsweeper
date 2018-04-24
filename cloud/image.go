package cloud

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	compute "google.golang.org/api/compute/v1"
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

func cleanupImages(images []Image) error {
	resList := []Resource{}
	for i := range images {
		v, ok := images[i].(Resource)
		if !ok {
			return errors.New("Could not convert Image to Resource")
		}
		resList = append(resList, v)
	}
	return cleanupResources(resList)
}

// AWS

type awsImage struct {
	baseImage
}

func (i *awsImage) Cleanup() error {
	log.Printf("Cleaning up image %s in %s", i.ID(), i.Owner())
	return awsTryWithBackoff(i.cleanup)
}

func (i *awsImage) cleanup() error {
	client := clientForAWSResource(i)
	input := &ec2.DeregisterImageInput{
		ImageId: aws.String(i.ID()),
	}
	_, err := client.DeregisterImage(input)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == requestLimitErrorCode {
			return errAWSRequestLimit
		}
	}
	return err
}

func (i *awsImage) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(i, key, value, overwrite)
}

func (i *awsImage) RemoveTag(key string) error {
	return removeAWSTag(i, key)
}

func (i *awsImage) MakePrivate() error {
	log.Printf("Making image %s private in %s", i.ID(), i.Owner())
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

// GCP

type gcpImage struct {
	baseImage
	compute *compute.Service
}

func (i *gcpImage) Cleanup() error {
	log.Printf("Cleaning up image %s in %s", i.ID(), i.Owner())
	_, err := i.compute.Images.Delete(i.Owner(), i.ID()).Do()
	return err
}

func (i *gcpImage) SetTag(key, value string, overwrite bool) error {
	img, err := i.compute.Images.Get(i.Owner(), i.ID()).Do()
	if err != nil {
		return nil
	}
	newLabels := img.Labels
	if newLabels == nil {
		newLabels = make(map[string]string)
	}
	if _, exist := newLabels[key]; exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, i.ID())
	}
	newLabels[key] = value
	req := &compute.GlobalSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: img.LabelFingerprint,
	}
	_, err = i.compute.Images.SetLabels(i.Owner(), i.ID(), req).Do()
	if err != nil {
		return err
	}
	i.tags = newLabels
	return nil
}

func (i *gcpImage) RemoveTag(key string) error {
	newLabels := make(map[string]string)
	for k, val := range i.tags {
		if k != key {
			newLabels[k] = val
		}
	}
	img, err := i.compute.Images.Get(i.Owner(), i.ID()).Do()
	if err != nil {
		return err
	}
	req := &compute.GlobalSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: img.LabelFingerprint,
	}
	_, err = i.compute.Images.SetLabels(i.Owner(), i.ID(), req).Do()
	if err != nil {
		return err
	}
	i.tags = newLabels
	return nil
}

func (i *gcpImage) MakePrivate() error {
	log.Println("Attempted to make GCP image private, NO-OP")
	return nil
}
