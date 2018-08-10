// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"errors"
	"log"
	"sync"
	"time"
)

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

func cleanupResources(resources []Resource) error {
	failed := false
	var wg sync.WaitGroup
	wg.Add(len(resources))
	for i := range resources {
		go func(index int) {
			err := resources[index].Cleanup()
			if err != nil {
				log.Printf("Cleaning up %s for owner %s failed\n%s\n", resources[index].ID(), resources[index].Owner(), err)
				failed = true
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if failed {
		return errors.New("One or more resource cleanups failed")
	}
	return nil
}
