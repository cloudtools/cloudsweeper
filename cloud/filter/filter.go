// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package filter

import (
	"github.com/agaridata/cloudsweeper/cloud"
)

// New will create a new resource filter ready to use
func New() *ResourceFilter {
	return &ResourceFilter{
		generalRules:  []func(cloud.Resource) bool{},
		instanceRules: []func(cloud.Instance) bool{},
		volumeRules:   []func(cloud.Volume) bool{},
		imageRules:    []func(cloud.Image) bool{},
		snapshotRules: []func(cloud.Snapshot) bool{},
		bucketRules:   []func(cloud.Bucket) bool{},

		OverrideWhitelist: false,
	}
}

// ResourceFilter is a dynamic filter that can have any amount
// of rules. The rules are used to determine which resources
// are kept when performing the filtering
type ResourceFilter struct {
	generalRules  []func(cloud.Resource) bool
	instanceRules []func(cloud.Instance) bool
	imageRules    []func(cloud.Image) bool
	volumeRules   []func(cloud.Volume) bool
	snapshotRules []func(cloud.Snapshot) bool
	bucketRules   []func(cloud.Bucket) bool

	OverrideWhitelist bool
}

// AddGeneralRule adds a generic resource rule, which is not specific to
// any particular type of resource.
func (f *ResourceFilter) AddGeneralRule(rule func(cloud.Resource) bool) {
	f.generalRules = append(f.generalRules, rule)
}

// AddInstanceRule adds an instance specific rule to the filter chain
func (f *ResourceFilter) AddInstanceRule(rule func(cloud.Instance) bool) {
	f.instanceRules = append(f.instanceRules, rule)
}

// AddImageRule adds an image specific rule to the filter chain
func (f *ResourceFilter) AddImageRule(rule func(cloud.Image) bool) {
	f.imageRules = append(f.imageRules, rule)
}

// AddVolumeRule adds a volume specific rule to the filter chain
func (f *ResourceFilter) AddVolumeRule(rule func(cloud.Volume) bool) {
	f.volumeRules = append(f.volumeRules, rule)
}

// AddSnapshotRule adds a snapshot specific rule to the filter chain
func (f *ResourceFilter) AddSnapshotRule(rule func(cloud.Snapshot) bool) {
	f.snapshotRules = append(f.snapshotRules, rule)
}

// AddBucketRule adds a bucket specific rule to the filter chain
func (f *ResourceFilter) AddBucketRule(rule func(cloud.Bucket) bool) {
	f.bucketRules = append(f.bucketRules, rule)
}

// Instances will filter the specified instances using the specified filters and
// return the instances which match. A boolean OR is performed between every specified
// filter.
func Instances(instances []cloud.Instance, filters ...*ResourceFilter) []cloud.Instance {
	resultList := []cloud.Instance{}
	for i := range instances {
		if or(instances[i], filters) {
			resultList = append(resultList, instances[i])
		}
	}
	return resultList
}

// Images will filter the specified images using the specified filters and
// return the images which match. A boolean OR is performed between every specified
// filter.
func Images(images []cloud.Image, filters ...*ResourceFilter) []cloud.Image {
	resultList := []cloud.Image{}
	for i := range images {
		if or(images[i], filters) {
			resultList = append(resultList, images[i])
		}
	}
	return resultList
}

// Volumes will filter the specified volumes using the specified filters and
// return the volumes which match. A boolean OR is performed between every specified
// filter.
func Volumes(volumes []cloud.Volume, filters ...*ResourceFilter) []cloud.Volume {
	resultList := []cloud.Volume{}
	for i := range volumes {
		if or(volumes[i], filters) {
			resultList = append(resultList, volumes[i])
		}
	}
	return resultList
}

// Snapshots will filter the specified snapshots using the specified filters and
// return the snapshots which match. A boolean OR is performed between every specified
// filter.
func Snapshots(snapshots []cloud.Snapshot, filters ...*ResourceFilter) []cloud.Snapshot {
	resultList := []cloud.Snapshot{}
	for i := range snapshots {
		if or(snapshots[i], filters) {
			resultList = append(resultList, snapshots[i])
		}
	}
	return resultList
}

// Buckets will filter the specified buckets using the specified filters and
// return the buckets which match. A boolean OR is performed between every specified
// filter.
func Buckets(buckets []cloud.Bucket, filters ...*ResourceFilter) []cloud.Bucket {
	resultList := []cloud.Bucket{}
	for i := range buckets {
		if or(buckets[i], filters) {
			resultList = append(resultList, buckets[i])
		}
	}
	return resultList
}
