package cloud

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	compute "google.golang.org/api/compute/v1"
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

func cleanupVolumes(volumes []Volume) error {
	resList := []Resource{}
	for i := range volumes {
		v, ok := volumes[i].(Resource)
		if !ok {
			return errors.New("Could not convert Volume to Resource")
		}
		resList = append(resList, v)
	}
	return cleanupResources(resList)
}

// AWS

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

// GCP

type gcpVolume struct {
	baseVolume
	compute *compute.Service
}

func (v *gcpVolume) Cleanup() error {
	log.Printf("Cleaning up volume %s in %s", v.ID(), v.Owner())
	_, err := v.compute.Disks.Delete(v.Owner(), v.Location(), v.ID()).Do()
	return err
}

func (v *gcpVolume) SetTag(key, value string, overwrite bool) error {
	disk, err := v.compute.Disks.Get(v.Owner(), v.Location(), v.ID()).Do()
	if err != nil {
		return err
	}
	newLabels := disk.Labels
	if newLabels == nil {
		newLabels = make(map[string]string)
	}
	if _, exist := newLabels[key]; exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, v.ID())
	}
	newLabels[key] = value
	req := &compute.ZoneSetLabelsRequest{
		LabelFingerprint: disk.LabelFingerprint,
		Labels:           newLabels,
	}
	_, err = v.compute.Disks.SetLabels(v.Owner(), v.Location(), v.ID(), req).Do()
	if err != nil {
		return err
	}
	v.tags = newLabels
	return nil
}

func (v *gcpVolume) RemoveTag(key string) error {
	newLabels := make(map[string]string)
	for k, val := range v.tags {
		if k != key {
			newLabels[k] = val
		}
	}
	disk, err := v.compute.Disks.Get(v.Owner(), v.Location(), v.ID()).Do()
	if err != nil {
		return err
	}
	req := &compute.ZoneSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: disk.LabelFingerprint,
	}
	_, err = v.compute.Disks.SetLabels(v.Owner(), v.Location(), v.ID(), req).Do()
	if err != nil {
		return err
	}
	v.tags = newLabels
	return nil
}
