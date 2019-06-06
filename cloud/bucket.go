// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	storage "google.golang.org/api/storage/v1"
)

type baseBucket struct {
	baseResource
	lastModified       time.Time
	objectCount        int64
	totalSizeGB        float64
	storageTypeSizesGB map[string]float64
}

func (b *baseBucket) LastModified() time.Time {
	return b.lastModified
}

func (b *baseBucket) ObjectCount() int64 {
	return b.objectCount
}

func (b *baseBucket) TotalSizeGB() float64 {
	return b.totalSizeGB
}

func (b *baseBucket) StorageTypeSizesGB() map[string]float64 {
	return b.storageTypeSizesGB
}

func cleanupBuckets(buckets []Bucket) error {
	resList := []Resource{}
	for i := range buckets {
		v, ok := buckets[i].(Resource)
		if !ok {
			return errors.New("Could not convert Bucket to Resource")
		}
		resList = append(resList, v)
	}
	return cleanupResources(resList)
}

// AWS

type awsBucket struct {
	baseBucket
}

func (b *awsBucket) Cleanup() error {
	log.Printf("Cleaning up bucket %s in %s", b.ID(), b.Owner())
	sess := session.Must(session.NewSession())
	creds := stscreds.NewCredentials(sess, fmt.Sprintf(assumeRoleARNTemplate, b.Owner()))
	s3Client := s3.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String(b.Location()),
	})

	var internalErr error
	err := s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(b.ID()),
	}, func(output *s3.ListObjectsV2Output, lastPage bool) bool {
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(b.ID()),
		}
		delete := &s3.Delete{
			Objects: []*s3.ObjectIdentifier{},
		}
		for i := range output.Contents {
			delete.Objects = append(delete.Objects, &s3.ObjectIdentifier{Key: output.Contents[i].Key})
		}
		input.Delete = delete
		if len(delete.Objects) == 0 {
			// A request with an empty list of objects is not allowed
			return true
		}
		out, e := s3Client.DeleteObjects(input)
		if e != nil {
			internalErr = e
			return false
		}
		if len(out.Errors) > 0 {
			for i := range out.Errors {
				log.Printf("ERROR: Could not delete '%s': %s\n", *out.Errors[i].Key, *out.Errors[i].Message)
			}
			internalErr = errors.New("Failed to delete one or more objects")
			return false
		}
		return !lastPage
	})
	if err != nil {
		return err
	}
	if internalErr != nil {
		return internalErr
	}

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(b.ID()),
	}
	_, err = s3Client.DeleteBucket(input)
	return err
}

func (b *awsBucket) SetTag(key, value string, overwrite bool) error {
	_, exist := b.Tags()[key]
	if exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, b.ID())
	}
	sess := session.Must(session.NewSession())
	creds := stscreds.NewCredentials(sess, fmt.Sprintf(assumeRoleARNTemplate, b.Owner()))
	s3Client := s3.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String(b.Location()),
	})
	tagging := &s3.Tagging{
		TagSet: []*s3.Tag{&s3.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}},
	}
	input := &s3.PutBucketTaggingInput{
		Bucket:  aws.String(b.ID()),
		Tagging: tagging,
	}
	_, err := s3Client.PutBucketTagging(input)
	return err
}

func (b *awsBucket) RemoveTag(key string) error {
	// TODO: Implement
	log.Fatalln("Not implemented for buckets")
	return errors.New("Not implemented for buckets")
}

// GCP

type gcpBucket struct {
	baseBucket
	storage *storage.Service
}

func (b *gcpBucket) Cleanup() error {
	log.Printf("Cleaning up bucket %s in %s", b.ID(), b.Owner())
	// TODO: Currently only works if bucket is empty, cleanup
	// the objects in the bucket too
	return b.storage.Buckets.Delete(b.ID()).Do()
}

func (b *gcpBucket) SetTag(key, value string, overwrite bool) error {
	log.Println("Bucket tagging not supported on GCP")
	return nil
}

func (b *gcpBucket) RemoveTag(key string) error {
	log.Println("Bucket tagging not supported on GCP")
	return nil
}
