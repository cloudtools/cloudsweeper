package filter

import (
	"brkt/housekeeper/cloud"
)

func (f *filter) shouldIncludeInstance(instance cloud.Instance) bool {
	for i := range f.generalRules {
		if !f.generalRules[i](instance) {
			return false
		}
	}
	for i := range f.instanceRules {
		if !f.instanceRules[i](instance) {
			return false
		}
	}
	_, isWhitelisted := instance.Tags()[WhitelistTagKey]
	return !isWhitelisted
}

func (f *filter) shouldIncludeVolume(volume cloud.Volume) bool {
	for i := range f.generalRules {
		if !f.generalRules[i](volume) {
			return false
		}
	}
	for i := range f.volumeRules {
		if !f.volumeRules[i](volume) {
			return false
		}
	}
	_, isWhitelisted := volume.Tags()[WhitelistTagKey]
	return !isWhitelisted
}

func (f *filter) shouldIncludeImage(image cloud.Image) bool {
	for i := range f.generalRules {
		if !f.generalRules[i](image) {
			return false
		}
	}
	for i := range f.imageRules {
		if !f.imageRules[i](image) {
			return false
		}
	}
	_, isWhitelisted := image.Tags()[WhitelistTagKey]
	return !isWhitelisted
}

func (f *filter) shouldIncludeSnapshot(snapshot cloud.Snapshot) bool {
	for i := range f.generalRules {
		if !f.generalRules[i](snapshot) {
			return false
		}
	}
	for i := range f.snapshotRules {
		if !f.snapshotRules[i](snapshot) {
			return false
		}
	}
	_, isWhitelisted := snapshot.Tags()[WhitelistTagKey]
	return !isWhitelisted
}
