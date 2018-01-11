package cloud

import "time"

type baseImage struct {
	id           string
	tags         map[string]string
	location     string
	public       bool
	creationTime time.Time
	name         string
}

func (i *baseImage) ID() string {
	return i.id
}

func (i *baseImage) Tags() map[string]string {
	return i.tags
}

func (i *baseImage) Location() string {
	return i.location
}

func (i *baseImage) Public() bool {
	return i.public
}

func (i *baseImage) CreationTime() time.Time {
	return i.creationTime
}

func (i *baseImage) Name() string {
	return i.name
}
