package cloud

type baseInstance struct {
	baseResource
	instanceType string
}

func (i *baseInstance) InstanceType() string {
	return i.instanceType
}
