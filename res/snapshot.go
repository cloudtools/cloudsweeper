package res

import "time"

type baseSnapshot struct {
	id           string
	tags         map[string]string
	location     string
	public       bool
	creationTime time.Time
	encrypted    bool
	sizeGB       int64
}

func (s *baseSnapshot) ID() string {
	return s.id
}

func (s *baseSnapshot) Tags() map[string]string {
	return s.tags
}

func (s *baseSnapshot) Location() string {
	return s.location
}

func (s *baseSnapshot) Public() bool {
	return s.public
}

func (s *baseSnapshot) CreationTime() time.Time {
	return s.creationTime
}

func (s *baseSnapshot) Encrypted() bool {
	return s.encrypted
}

func (s *baseSnapshot) SizeGB() int64 {
	return s.sizeGB
}
