// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	compute "google.golang.org/api/compute/v1"
)

type baseInstance struct {
	baseResource
	instanceType string
}

func (i *baseInstance) InstanceType() string {
	return i.instanceType
}

func cleanupInstances(instances []Instance) error {
	resList := []Resource{}
	for i := range instances {
		v, ok := instances[i].(Resource)
		if !ok {
			return errors.New("Could not convert Instance to Resource")
		}
		resList = append(resList, v)
	}
	return cleanupResources(resList)
}

// AWS

type awsInstance struct {
	baseInstance
}

// Cleanup will termiante this instance
func (i *awsInstance) Cleanup() error {
	log.Printf("Cleaning up instance %s in %s", i.ID(), i.Owner())
	return awsTryWithBackoff(i.cleanup)
}

func (i *awsInstance) cleanup() error {
	client := clientForAWSResource(i)
	input := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{i.id}),
	}
	_, err := client.TerminateInstances(input)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == requestLimitErrorCode {
			return errAWSRequestLimit
		}
	}
	return err
}

func (i *awsInstance) SetTag(key, value string, overwrite bool) error {
	return addAWSTag(i, key, value, overwrite)
}

func (i *awsInstance) RemoveTag(key string) error {
	return removeAWSTag(i, key)
}

// GCP

type gcpInstance struct {
	baseInstance
	compute *compute.Service
}

func (i *gcpInstance) Cleanup() error {
	log.Printf("Cleaning up instance %s in %s", i.ID(), i.Owner())
	_, err := i.compute.Instances.Delete(i.Owner(), i.Location(), i.ID()).Do()
	return err
}

func (i *gcpInstance) SetTag(key, value string, overwrite bool) error {
	inst, err := i.compute.Instances.Get(i.Owner(), i.Location(), i.ID()).Do()
	if err != nil {
		return err
	}
	newLabels := inst.Labels
	if newLabels == nil {
		newLabels = make(map[string]string)
	}
	if _, exist := newLabels[key]; exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, i.ID())
	}
	newLabels[key] = value
	req := &compute.InstancesSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: inst.LabelFingerprint,
	}
	_, err = i.compute.Instances.SetLabels(i.Owner(), i.Location(), i.ID(), req).Do()
	if err != nil {
		return err
	}
	i.tags = newLabels
	return nil
}

func (i *gcpInstance) RemoveTag(key string) error {
	newLabels := make(map[string]string)
	for k, val := range i.tags {
		if k != key {
			newLabels[k] = val
		}
	}
	inst, err := i.compute.Instances.Get(i.Owner(), i.Location(), i.ID()).Do()
	if err != nil {
		return err
	}
	req := &compute.InstancesSetLabelsRequest{
		Labels:           newLabels,
		LabelFingerprint: inst.LabelFingerprint,
	}
	_, err = i.compute.Instances.SetLabels(i.Owner(), i.Location(), i.ID(), req).Do()
	if err != nil {
		return err
	}
	i.tags = newLabels
	return nil
}
