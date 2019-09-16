// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package filter

import (
	"fmt"
	"testing"
	"time"

	"github.com/agaridata/cloudsweeper/cloud"
)

func TestAddingFilters(t *testing.T) {
	fil := New()
	fil.AddGeneralRule(func(r cloud.Resource) bool { return true })
	if len(fil.generalRules) != 1 {
		t.Error("General rule not added")
	}
	fil.AddInstanceRule(func(r cloud.Instance) bool { return true })
	if len(fil.instanceRules) != 1 {
		t.Error("Instance rule not added")
	}
	fil.AddVolumeRule(func(r cloud.Volume) bool { return true })
	if len(fil.volumeRules) != 1 {
		t.Error("Volume rule not added")
	}
	fil.AddImageRule(func(r cloud.Image) bool { return true })
	if len(fil.imageRules) != 1 {
		t.Error("Image rule not added")
	}
	fil.AddSnapshotRule(func(r cloud.Snapshot) bool { return true })
	if len(fil.snapshotRules) != 1 {
		t.Error("Snapshot rule not added")
	}
	fil.AddBucketRule(func(r cloud.Bucket) bool { return true })
	if len(fil.bucketRules) != 1 {
		t.Error("Bucket rule not added")
	}
}

type testInstance struct {
	testResource
	instType string
}

func (i *testInstance) InstanceType() string {
	return i.instType
}

// Testing using a single filter and multiple filters for the same
// resource type is identical for all instance types, so the tests
// here only do cloud.Instance, but should cover all resource types.
// This does not cover the case of mixing different resource types.
func TestSingleInstanceFilter(t *testing.T) {
	inst1 := &testInstance{}
	inst1.creationTime = time.Now().AddDate(0, 0, -5)
	inst2 := &testInstance{}
	inst2.creationTime = time.Now()

	fil := New()
	fil.AddGeneralRule(OlderThanXDays(2))

	filtered := Instances([]cloud.Instance{inst1, inst2}, fil)
	if len(filtered) != 1 {
		t.Error("Failed filtering")
	}
}

func TestMultipleInstanceFilter(t *testing.T) {
	inst1 := &testInstance{}
	inst1.creationTime = time.Now().AddDate(0, 0, -5)

	inst2 := &testInstance{}
	inst2.creationTime = time.Now()
	inst2.instType = "instance-type"

	fil1 := New()
	fil1.AddGeneralRule(OlderThanXDays(2))

	fil2 := New()
	fil2.AddInstanceRule(func(i cloud.Instance) bool {
		return i.InstanceType() == "instance-type"
	})

	filtered := Instances([]cloud.Instance{inst1, inst2}, fil1, fil2)
	if len(filtered) != 2 {
		for _, inst := range filtered {
			fmt.Println(inst)
		}
		t.Error("Failed filtering with multiple filters")
	}
}

type testImg struct {
	testResource
}

func (i *testImg) Name() string       { return "test-img" }
func (i *testImg) SizeGB() int64      { return 10 }
func (i *testImg) MakePrivate() error { return nil }

// This will test the filters being used when marking resources for
// cleanup. These are:
// 		- unattached volumes > 30 days old
//		- unused/unaccessed buckets > 6 months (182 days)
// 		- non-whitelisted AMIs > 6 months
// 		- non-whitelisted snapshots > 6 months
// 		- non-whitelisted volumes > 6 months
//		- untagged resources > 30 days (this should take care of instances)
func TestCleanupRulesFilter(t *testing.T) {

	// Setup the filters used
	untaggedFilter := New()
	untaggedFilter.AddGeneralRule(func(r cloud.Resource) bool {
		return len(r.Tags()) == 0
	})
	untaggedFilter.AddGeneralRule(OlderThanXDays(30))
	untaggedFilter.AddSnapshotRule(IsNotInUse())
	untaggedFilter.AddGeneralRule(Negate(TaggedForCleanup()))

	oldFilter := New()
	oldFilter.AddGeneralRule(OlderThanXMonths(6))
	// Don't cleanup resources tagged for release
	oldFilter.AddGeneralRule(Negate(HasTag("Release")))
	oldFilter.AddSnapshotRule(IsNotInUse())
	oldFilter.AddGeneralRule(Negate(TaggedForCleanup()))

	unattachedFilter := New()
	unattachedFilter.AddVolumeRule(IsUnattached())
	unattachedFilter.AddGeneralRule(OlderThanXDays(30))
	unattachedFilter.AddGeneralRule(Negate(HasTag("Release")))
	unattachedFilter.AddGeneralRule(Negate(TaggedForCleanup()))

	bucketFilter := New()
	bucketFilter.AddBucketRule(NotModifiedInXDays(182))
	bucketFilter.AddGeneralRule(OlderThanXDays(7))
	bucketFilter.AddGeneralRule(Negate(HasTag("Release")))
	bucketFilter.AddGeneralRule(Negate(TaggedForCleanup()))

	// Create some helper tag maps
	someTags := map[string]string{"test-key": "test-value"}
	whitelistTags := map[string]string{"cloudsweeper-whitelisted": ""}

	// Test instances
	// No
	inst1 := &testInstance{}
	inst1.creationTime = time.Now().AddDate(0, -3, 0)
	inst1.tags = someTags

	// Yes
	inst2 := &testInstance{}
	inst2.creationTime = time.Now().AddDate(0, -4, 0)

	// No
	inst3 := &testInstance{}
	inst3.creationTime = time.Now().AddDate(-5, 0, 0)
	inst3.tags = whitelistTags

	// No
	inst4 := &testInstance{}
	inst4.creationTime = time.Now()

	filInst := Instances([]cloud.Instance{inst1, inst2, inst3, inst4}, untaggedFilter)
	if len(filInst) != 1 {
		t.Error("Failed to filter instances")
	}

	// Test images
	// No
	img1 := &testImg{}
	img1.creationTime = time.Now()

	// No
	img2 := &testImg{}
	img2.creationTime = time.Now().AddDate(-3, 0, 0)
	img2.tags = whitelistTags

	// Yes
	img3 := &testImg{}
	img3.creationTime = time.Now().AddDate(0, -6, -3)

	// No
	img4 := &testImg{}
	img4.creationTime = time.Now().AddDate(0, 0, -3)
	img4.tags = someTags

	// No
	img5 := &testImg{}
	img5.creationTime = time.Now().AddDate(-3, 0, -5)
	img5.tags = map[string]string{"release": ""}

	// Yes
	img6 := &testImg{}
	img6.creationTime = time.Now().AddDate(0, 0, -32)

	filImg := Images([]cloud.Image{img1, img2, img3, img4, img5, img6}, oldFilter, untaggedFilter)
	if len(filImg) != 2 {
		t.Error("Failed to filter images")
	}

	// Test volumes
	// No
	vol1 := &testVolume{}
	vol1.creationTime = time.Now()
	vol1.attached = false

	// Yes
	vol2 := &testVolume{}
	vol2.attached = true
	vol2.creationTime = time.Now().AddDate(0, -6, -1)

	// No
	vol3 := &testVolume{}
	vol3.attached = true
	vol3.creationTime = time.Now().AddDate(0, 0, -32)
	vol3.tags = someTags

	// No
	vol4 := &testVolume{}
	vol4.creationTime = time.Now().AddDate(0, 0, -2)
	vol4.attached = false

	// No
	vol5 := &testVolume{}
	vol5.creationTime = time.Now().AddDate(0, -9, 0)
	vol5.attached = false
	vol5.tags = whitelistTags

	filVol := Volumes([]cloud.Volume{vol1, vol2, vol3, vol4, vol5}, oldFilter, unattachedFilter, untaggedFilter)
	if len(filVol) != 1 {
		t.Error("Failed to filter volumes")
	}

	// Test snapshots
	// No
	snap1 := &testSnap{}
	snap1.creationTime = time.Now().AddDate(0, 0, -3)

	// Yes
	snap2 := &testSnap{}
	snap2.creationTime = time.Now().AddDate(-4, 0, 0)
	snap2.tags = someTags

	// Yes
	snap3 := &testSnap{}
	snap3.creationTime = time.Now().AddDate(0, 0, -40)

	// No
	snap4 := &testSnap{}
	snap4.creationTime = time.Now().AddDate(0, -8, 4)
	snap4.tags = whitelistTags

	filSnaps := Snapshots([]cloud.Snapshot{snap1, snap2, snap3, snap4}, oldFilter, untaggedFilter)
	if len(filSnaps) != 2 {
		t.Error("Failed to filter snapshots")
	}

	// Test buckets
	// No
	buck1 := &testBucket{}
	buck1.creationTime = time.Now().AddDate(0, -8, 0)
	buck1.lastModified = time.Now().AddDate(0, 0, -2)

	// Yes
	buck2 := &testBucket{}
	buck2.creationTime = time.Now().AddDate(-7, 0, 0)
	buck2.lastModified = time.Now().AddDate(-2, 0, 0)
	buck2.tags = someTags

	// No
	buck3 := &testBucket{}
	buck3.creationTime = time.Now().AddDate(0, 0, -45)
	buck3.lastModified = time.Now()

	filBucks := Buckets([]cloud.Bucket{buck1, buck2, buck3}, bucketFilter)
	if len(filBucks) != 1 {
		t.Error("Failed to filter buckets")
	}
}
