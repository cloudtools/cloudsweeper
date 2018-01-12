package cloud

type baseImage struct {
	baseResource
	name string
}

func (i *baseImage) Name() string {
	return i.name
}
