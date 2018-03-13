package cloud

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	compute "google.golang.org/api/compute/v1"
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

func cleanupSnapshots(snapshots []Snapshot) error {
	resList := []Resource{}
	for i := range snapshots {
		v, ok := snapshots[i].(Resource)
		if !ok {
			return errors.New("Could not convert Snapshot to Resource")
		}
		resList = append(resList, v)
	}
	return cleanupResources(resList)
}

// AWS

type awsSnapshot struct {
	baseSnapshot
}

func (s *awsSnapshot) Cleanup() error {
	log.Printf("Cleaning up snapshot %s in %s", s.ID(), s.Owner())
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

func (s *awsSnapshot) RemoveTag(key string) error {
	return removeAWSTag(s, key)
}

// GCP

type gcpSnapshot struct {
	baseSnapshot
	compute *compute.Service
}

func (s *gcpSnapshot) Cleanup() error {
	log.Printf("Cleaning up snapshot %s in %s", s.ID(), s.Owner())
	_, err := s.compute.Snapshots.Delete(s.Owner(), s.ID()).Do()
	return err
}

func (s *gcpSnapshot) SetTag(key, value string, overwrite bool) error {
	snap, err := s.compute.Snapshots.Get(s.Owner(), s.ID()).Do()
	if err != nil {
		return err
	}
	newLabels := snap.Labels
	if newLabels == nil {
		newLabels = make(map[string]string)
	}
	if _, exist := newLabels[key]; exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, s.ID())
	}
	newLabels[key] = value
	req := &compute.GlobalSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: snap.LabelFingerprint,
	}
	_, err = s.compute.Snapshots.SetLabels(s.Owner(), s.ID(), req).Do()
	if err != nil {
		return err
	}
	s.tags = newLabels
	return nil
}

func (s *gcpSnapshot) RemoveTag(key string) error {
	newLabels := make(map[string]string)
	for k, val := range s.tags {
		if k != key {
			newLabels[k] = val
		}
	}
	snap, err := s.compute.Snapshots.Get(s.Owner(), s.ID()).Do()
	if err != nil {
		return err
	}
	req := &compute.GlobalSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: snap.LabelFingerprint,
	}
	_, err = s.compute.Snapshots.SetLabels(s.Owner(), s.ID(), req).Do()
	if err != nil {
		return err
	}
	s.tags = newLabels
	return nil
}
