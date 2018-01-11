package cloud

import "time"

type baseVolume struct {
	id           string
	tags         map[string]string
	location     string
	public       bool
	creationTime time.Time
	sizeGB       int64
	attached     bool
	encrypted    bool
	volumeType   string
}

func (v *baseVolume) ID() string {
	return v.id
}

func (v *baseVolume) Tags() map[string]string {
	return v.tags
}

func (v *baseVolume) Location() string {
	return v.location
}

func (v *baseVolume) Public() bool {
	return v.public
}

func (v *baseVolume) CreationTime() time.Time {
	return v.creationTime
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
