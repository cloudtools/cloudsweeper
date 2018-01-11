package res

import (
	"log"
	"time"
)

// ResourceManager is used to manage the different resources on
// a CSP. It can be used to get e.g. all instances for all accounts
// in AWS.
type ResourceManager interface {
	// InstancesPerAccount returns a mapping from account/project
	// to its associated instances
	InstancesPerAccount() map[string][]Instance
	// ImagesPerAccount returns a mapping from account/project
	// to its associated images
	ImagesPerAccount() map[string][]Image
	// VolumesPerAccount returns a mapping from account/project
	// to its associated volumes
	VolumesPerAccount() map[string][]Volume
}

// Resource represents a generic resource in any CSP. It should be
// concretizised further.
type Resource interface {
	ID() string
	Tags() map[string]string
	Location() string
	Public() bool
	CreationTime() time.Time
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

type csp int

const (
	// AWS is AWS
	AWS csp = iota
	// GCP is Google Cloud Platform
	GCP
)

// NewManager will build a new resource manager for the specified CSP
func NewManager(c csp, accounts ...string) ResourceManager {
	switch c {
	case AWS:
		log.Println("Initializing AWS Resource Manager")
		manager := &awsResourceManager{
			accounts: accounts,
		}
		return manager
	case GCP:
		log.Fatalln("Unfortunately, GCP is currently not supported")
	default:
		log.Fatalln("Invalid CSP specified")
	}
	return nil
}
