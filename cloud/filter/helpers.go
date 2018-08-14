// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package filter

import (
	"brkt/cloudsweeper/cloud"
)

func (f *ResourceFilter) includeResource(resource cloud.Resource) bool {
	for i := range f.generalRules {
		if !f.generalRules[i](resource) {
			return false
		}
	}
	return true
}

func (f *ResourceFilter) includeInstance(instance cloud.Instance) bool {
	if !f.includeResource(instance) {
		return false
	}
	for i := range f.instanceRules {
		if !f.instanceRules[i](instance) {
			return false
		}
	}
	_, isWhitelisted := instance.Tags()[WhitelistTagKey]
	return !isWhitelisted || f.OverrideWhitelist
}

func (f *ResourceFilter) includeVolume(volume cloud.Volume) bool {
	if !f.includeResource(volume) {
		return false
	}
	for i := range f.volumeRules {
		if !f.volumeRules[i](volume) {
			return false
		}
	}
	_, isWhitelisted := volume.Tags()[WhitelistTagKey]
	return !isWhitelisted || f.OverrideWhitelist
}

func (f *ResourceFilter) includeImage(image cloud.Image) bool {
	if !f.includeResource(image) {
		return false
	}
	for i := range f.imageRules {
		if !f.imageRules[i](image) {
			return false
		}
	}
	_, isWhitelisted := image.Tags()[WhitelistTagKey]
	return !isWhitelisted || f.OverrideWhitelist
}

func (f *ResourceFilter) includeSnapshot(snapshot cloud.Snapshot) bool {
	if !f.includeResource(snapshot) {
		return false
	}
	for i := range f.snapshotRules {
		if !f.snapshotRules[i](snapshot) {
			return false
		}
	}
	_, isWhitelisted := snapshot.Tags()[WhitelistTagKey]
	return !isWhitelisted || f.OverrideWhitelist
}

func (f *ResourceFilter) includeBucket(bucket cloud.Bucket) bool {
	if !f.includeResource(bucket) {
		return false
	}
	for i := range f.bucketRules {
		if !f.bucketRules[i](bucket) {
			return false
		}
	}
	_, isWhitelisted := bucket.Tags()[WhitelistTagKey]
	return !isWhitelisted || f.OverrideWhitelist
}

func or(resource cloud.Resource, filters []*ResourceFilter) bool {
	if inst, ok := resource.(cloud.Instance); ok {
		for _, filter := range filters {
			if filter.includeInstance(inst) {
				return true
			}
		}
		return false
	}

	if img, ok := resource.(cloud.Image); ok {
		for _, filter := range filters {
			if filter.includeImage(img) {
				return true
			}
		}
		return false
	}

	if vol, ok := resource.(cloud.Volume); ok {
		for _, filter := range filters {
			if filter.includeVolume(vol) {
				return true
			}
		}
		return false
	}

	if snap, ok := resource.(cloud.Snapshot); ok {
		for _, filter := range filters {
			if filter.includeSnapshot(snap) {
				return true
			}
		}
		return false
	}

	if buck, ok := resource.(cloud.Bucket); ok {
		for _, filter := range filters {
			if filter.includeBucket(buck) {
				return true
			}
		}
		return false
	}

	return false
}
