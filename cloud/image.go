package cloud

type baseImage struct {
	baseResource
	name   string
	sizeGB int64
}

func (i *baseImage) Name() string {
	return i.name
}

func (i *baseImage) SizeGB() int64 {
	return i.sizeGB
}
