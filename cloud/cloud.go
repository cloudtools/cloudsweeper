// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	oauth2 "golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	storage "google.golang.org/api/storage/v1"
)

const (
	// GcpCredentialsFileKey is the Env variable to store path
	// to service accounts credentials JSON file
	GcpCredentialsFileKey = "GOOGLE_APPLICATION_CREDENTIALS"

	scopeGCPCompute = "https://www.googleapis.com/auth/compute"
	scopeGCPStorage = "https://www.googleapis.com/auth/devstorage.read_write"
)

// ResourceManager is used to manage the different resources on
// a CSP. It can be used to get e.g. all instances for all accounts
// in AWS.
type ResourceManager interface {
	// Owners return a list of all owners the manager handle
	Owners() []string
	// BucketsPerAccount returns a mapping from account/project to
	// its associated buckets
	BucketsPerAccount() map[string][]Bucket
	// InstancesPerAccount returns a mapping from account/project
	// to its associated instances
	InstancesPerAccount() map[string][]Instance
	// ImagesPerAccount returns a mapping from account/project
	// to its associated images
	ImagesPerAccount() map[string][]Image
	// VolumesPerAccount returns a mapping from account/project
	// to its associated volumes
	VolumesPerAccount() map[string][]Volume
	// SnapshotsPerAccount returns a mapping from account/project
	// to its associated snaphots
	SnapshotsPerAccount() map[string][]Snapshot
	// AllResourcesPerAccount will return a mapping from account/project
	// to all of the resources associated with that account/project
	AllResourcesPerAccount() map[string]*ResourceCollection
	// CleanupInstances termiantes a list of instances, which is faster
	// than calling Cleanup() on every individual instance
	CleanupInstances([]Instance) error
	// CleanupImages de-registers a list of images
	CleanupImages([]Image) error
	// CleanupVolumes deletes a list of volumes
	CleanupVolumes([]Volume) error
	// CleanupSnapshots delete a list of snapshots
	CleanupSnapshots([]Snapshot) error
	// CleanupBuckets deletes the specified buckets
	CleanupBuckets([]Bucket) error
}

// Resource represents a generic resource in any CSP. It should be
// concretizised further.
type Resource interface {
	CSP() CSP
	Owner() string
	ID() string
	Tags() map[string]string
	Location() string
	Public() bool
	CreationTime() time.Time

	SetTag(key, value string, overwrite bool) error
	RemoveTag(key string) error
	Cleanup() error
}

// Instance composes the Resource interface, and descibes an instance
// in any CSP.
type Instance interface {
	Resource
	InstanceType() string
}

// Image composes the Resource interface, and descibe an image in
// any CSP. Such as an AMI in AWS.
type Image interface {
	Resource
	Name() string
	SizeGB() int64

	MakePrivate() error
}

// Volume composes the Resource interface, and describe a volume in
// any CSP.
type Volume interface {
	Resource
	SizeGB() int64
	Attached() bool
	Encrypted() bool
	VolumeType() string
}

// Snapshot composes the Resource interface, and describe a snapshot
// in any CSP.
type Snapshot interface {
	Resource
	Encrypted() bool
	InUse() bool
	SizeGB() int64
}

// Bucket represents a bucket in a CSP, such as an S3 bucket in AWS
type Bucket interface {
	Resource
	LastModified() time.Time
	ObjectCount() int64
	TotalSizeGB() float64
	StorageTypeSizesGB() map[string]float64
}

// ResourceCollection encapsulates collections of multiple resources. Does not
// include buckets.
type ResourceCollection struct {
	Owner     string
	Instances []Instance
	Images    []Image
	Volumes   []Volume
	Snapshots []Snapshot
}

// CSP represent a cloud service provider, such as AWS
type CSP string

const (
	// AWS is AWS
	AWS CSP = "AWS"
	// GCP is Google Cloud Platform
	GCP CSP = "GCP"
)

// NewManager will build a new resource manager for the specified CSP
func NewManager(c CSP, accounts ...string) (ResourceManager, error) {
	switch c {
	case AWS:
		log.Println("Initializing AWS Resource Manager")
		manager := &awsResourceManager{
			accounts: accounts,
		}
		return manager, nil
	case GCP:
		log.Println("Initializing GCP Resource Manager")
		client, err := getGCPHttpClient()
		if err != nil {
			return nil, err
		}
		computeService, err := compute.New(client)
		if err != nil {
			return nil, fmt.Errorf("Could not initialize compute service: %s", err)
		}
		storageService, err := storage.New(client)
		if err != nil {
			return nil, fmt.Errorf("Coult not initialize storage service: %s", err)
		}
		manager := &gcpResourceManager{
			projects: accounts,
			compute:  computeService,
			storage:  storageService,
		}
		return manager, nil
	default:
		return nil, fmt.Errorf("Invalid CSP specified: %s", c)
	}
}

func getGCPHttpClient() (*http.Client, error) {
	credsFile, exist := os.LookupEnv(GcpCredentialsFileKey)
	if !exist {
		log.Println("No GCP credentials specified, using default")
		return oauth2.DefaultClient(context.Background(), scopeGCPCompute, scopeGCPStorage)
	}
	creds, err := ioutil.ReadFile(credsFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read GCP credentials JSON: %s", err)
	}
	conf, err := oauth2.JWTConfigFromJSON(creds, scopeGCPCompute, scopeGCPStorage)
	if err != nil {
		return nil, fmt.Errorf("Could not get GCP credentials: %s", err)
	}
	return conf.Client(context.Background()), nil
}
