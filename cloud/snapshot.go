package cloud

type baseSnapshot struct {
	baseResource
	encrypted bool
	sizeGB    int64
}

func (s *baseSnapshot) Encrypted() bool {
	return s.encrypted
}

func (s *baseSnapshot) SizeGB() int64 {
	return s.sizeGB
}
