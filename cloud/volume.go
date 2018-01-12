package cloud

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
