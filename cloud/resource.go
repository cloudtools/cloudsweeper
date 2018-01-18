package cloud

import "time"

type baseResource struct {
	csp          CSP
	owner        string
	id           string
	tags         map[string]string
	location     string
	public       bool
	creationTime time.Time
}

func (r *baseResource) CSP() CSP {
	return r.csp
}

func (r *baseResource) Owner() string {
	return r.owner
}

func (r *baseResource) ID() string {
	return r.id
}

func (r *baseResource) Tags() map[string]string {
	return r.tags
}

func (r *baseResource) Location() string {
	return r.location
}

func (r *baseResource) Public() bool {
	return r.public
}

func (r *baseResource) CreationTime() time.Time {
	return r.creationTime
}
