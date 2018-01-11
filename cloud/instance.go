package cloud

import "time"

type baseInstance struct {
	id           string
	tags         map[string]string
	location     string
	public       bool
	launchTime   time.Time
	instanceType string
}

func (i *baseInstance) ID() string {
	return i.id
}

func (i *baseInstance) Tags() map[string]string {
	return i.tags
}

func (i *baseInstance) Location() string {
	return i.location
}

func (i *baseInstance) Public() bool {
	// An instance being public in this case means it has a public IP
	return i.public
}

func (i *baseInstance) CreationTime() time.Time {
	return i.launchTime
}

func (i *baseInstance) InstanceType() string {
	return i.instanceType
}
