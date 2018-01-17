package filter

import (
	"brkt/housekeeper/cloud"
)

// ResourceFilter is a dynamic filter that can have any amount
// of rules. The rules are used to determine which resources
// are kept when performing the filtering
type ResourceFilter interface {
	AddGeneralRule(func(cloud.Resource) bool)
	AddInstanceRule(func(cloud.Instance) bool)
	AddImageRule(func(cloud.Image) bool)
	AddVolumeRule(func(cloud.Volume) bool)
	AddSnapshotRule(func(cloud.Snapshot) bool)

	FilterInstances([]cloud.Instance) []cloud.Instance
	FilterVolumes([]cloud.Volume) []cloud.Volume
	FilterImages([]cloud.Image) []cloud.Image
	FilterSnapshots([]cloud.Snapshot) []cloud.Snapshot
}

// New will create a new resource filter ready to use
func New() ResourceFilter {
	return &filter{
		generalRules: []func(cloud.Resource) bool{},
	}
}

type filter struct {
	generalRules  []func(cloud.Resource) bool
	instanceRules []func(cloud.Instance) bool
	imageRules    []func(cloud.Image) bool
	volumeRules   []func(cloud.Volume) bool
	snapshotRules []func(cloud.Snapshot) bool
}

func (f *filter) AddGeneralRule(rule func(cloud.Resource) bool) {
	f.generalRules = append(f.generalRules, rule)
}

func (f *filter) AddInstanceRule(rule func(cloud.Instance) bool) {
	f.instanceRules = append(f.instanceRules, rule)
}

func (f *filter) AddImageRule(rule func(cloud.Image) bool) {
	f.imageRules = append(f.imageRules, rule)
}

func (f *filter) AddVolumeRule(rule func(cloud.Volume) bool) {
	f.volumeRules = append(f.volumeRules, rule)
}

func (f *filter) AddSnapshotRule(rule func(cloud.Snapshot) bool) {
	f.snapshotRules = append(f.snapshotRules, rule)
}

func (f *filter) FilterInstances(instances []cloud.Instance) []cloud.Instance {
	resultList := []cloud.Instance{}
	for _, instance := range instances {
		if f.shouldIncludeInstance(instance) {
			resultList = append(resultList, instance)
		}
	}
	return resultList
}

func (f *filter) FilterImages(images []cloud.Image) []cloud.Image {
	resultList := []cloud.Image{}
	for _, image := range images {
		if f.shouldIncludeImage(image) {
			resultList = append(resultList, image)
		}
	}
	return resultList
}

func (f *filter) FilterVolumes(volumes []cloud.Volume) []cloud.Volume {
	resultList := []cloud.Volume{}
	for _, volume := range volumes {
		if f.shouldIncludeVolume(volume) {
			resultList = append(resultList, volume)
		}
	}
	return resultList
}

func (f *filter) FilterSnapshots(snapshots []cloud.Snapshot) []cloud.Snapshot {
	resultList := []cloud.Snapshot{}
	for _, snapshot := range snapshots {
		if f.shouldIncludeSnapshot(snapshot) {
			resultList = append(resultList, snapshot)
		}
	}
	return resultList
}
