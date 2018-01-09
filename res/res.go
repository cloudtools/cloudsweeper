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

// Instance inhertis the Resource interface, and descibes an instance
// in any CSP.
type Instance interface {
	Resource
	InstanceType() string
}

type csp int

const (
	// AwsCSP is AWS
	AwsCSP csp = iota
	// GcpCSP is Google Cloud Platform
	GcpCSP
)

// NewManager will build a new resource manager for the specified CSP
func NewManager(c csp, accounts ...string) ResourceManager {
	switch c {
	case AwsCSP:
		log.Println("Initializing AWS Resource Manager")
		manager := &awsResourceManager{
			accounts: accounts,
		}
		return manager
	case GcpCSP:
		log.Fatalln("Unfortunately, GCP is currently not supported")
	default:
		log.Fatalln("Invalid CSP specified")
	}
	return nil
}
