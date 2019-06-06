// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package filter

import (
	"testing"
	"time"

	"github.com/cloudtools/cloudsweeper/cloud"
)

const (
	testOwner    = "475063612724"
	testID       = "some-resource-id"
	testLocation = "us-west-2"
	testCSP      = cloud.AWS
	testPublic   = false

	testSize       = 10
	testEncrypted  = false
	testVolumeType = "volume-type"
)

type testResource struct {
	creationTime time.Time
	tags         map[string]string
}

func (r *testResource) CSP() cloud.CSP                                 { return testCSP }
func (r *testResource) Owner() string                                  { return testOwner }
func (r *testResource) ID() string                                     { return testID }
func (r *testResource) Tags() map[string]string                        { return r.tags }
func (r *testResource) Location() string                               { return testLocation }
func (r *testResource) Public() bool                                   { return testPublic }
func (r *testResource) CreationTime() time.Time                        { return r.creationTime }
func (r *testResource) SetTag(key, value string, overwrite bool) error { return nil }
func (r *testResource) RemoveTag(key string) error                     { return nil }
func (r *testResource) Cleanup() error                                 { return nil }

func TestNegate(t *testing.T) {
	foo := &testResource{time.Now(), map[string]string{}}
	fun := Negate(func(r cloud.Resource) bool {
		return true
	})
	res := fun(foo)
	if res != false {
		t.Error("Failed to negate, got true when expected false")
	}
}

func TestAlreadyTaggedForDelete(t *testing.T) {
	foo := &testResource{time.Now(), map[string]string{}}
	foo.tags = map[string]string{DeleteTagKey: time.Now().Format(time.RFC3339)}
	fun := TaggedForCleanup()
	res := fun(foo)
	if !res {
		t.Error("The resource should be tagged for cleanup")
	}
}

func TestOlderHours(t *testing.T) {
	oldTime := time.Now().Add(-(10 * time.Hour))
	foo := &testResource{oldTime, map[string]string{}}

	if !OlderThanXHours(5)(foo) {
		t.Error("Resource is older than 5 hours")
	}

	foo.creationTime = time.Now()
	if OlderThanXHours(5)(foo) {
		t.Error("Resource is not older than 5 hours")
	}
}

func TestOlderDays(t *testing.T) {
	oldTime := time.Now().Add(-(100 * time.Hour))
	foo := &testResource{oldTime, map[string]string{}}

	if !OlderThanXDays(2)(foo) {
		t.Error("Resource is older than 2 days")
	}

	foo.creationTime = time.Now()
	if OlderThanXDays(2)(foo) {
		t.Error("Resource is not older than 2 days")
	}
}

func TestOlderMonths(t *testing.T) {
	oldTime := time.Now().AddDate(0, -5, 0)
	foo := &testResource{oldTime, map[string]string{}}

	if !OlderThanXMonths(2)(foo) {
		t.Error("Resource is older than 2 months")
	}

	foo.creationTime = time.Now()

	if OlderThanXMonths(2)(foo) {
		t.Error("Resource is not older than 2 months")
	}
}

func TestOlderYears(t *testing.T) {
	oldTime := time.Now().AddDate(-10, 0, 0)
	foo := &testResource{oldTime, map[string]string{}}

	if !OlderThanXYears(4)(foo) {
		t.Error("Resource is older than 4 years")
	}

	foo.creationTime = time.Now()
	if OlderThanXYears(4)(foo) {
		t.Error("Resource is not older than 4 years")
	}
}

func TestNames(t *testing.T) {
	tags := make(map[string]string)

	tags["Name"] = "SomeCoolName"

	foo := &testResource{time.Now(), tags}

	if !NameContains("SomeCoolName")(foo) {
		t.Error("Resource should contain name")
	}

	if !NameContains("Cool")(foo) {
		t.Error("Resource should contain subset of name")
	}

	foo.tags = map[string]string{}
	if NameContains("SomeCoolName")(foo) {
		t.Error("Resource does not have name")
	}

}

func TestIDMatch(t *testing.T) {
	foo := &testResource{time.Now(), map[string]string{}}

	if !IDMatches(testID)(foo) {
		t.Error("Resource ID should match")
	}

	if IDMatches("not-a-good-id")(foo) {
		t.Error("Resource ID should not match")
	}
}

func TestHasTag(t *testing.T) {
	tags := make(map[string]string)
	tags["some-tag-key"] = "some-tag-value"

	foo := &testResource{time.Now(), tags}

	if !HasTag("some-tag-key")(foo) {
		t.Error("Resource should have tag")
	}

	if HasTag("some-tag")(foo) {
		t.Error("Resource does not have tag")
	}
}

func TestPublic(t *testing.T) {
	foo := &testResource{time.Now(), map[string]string{}}

	if IsPublic()(foo) != testPublic {
		t.Error("Resource public value wrong")
	}
}

func TestLifetimeExceeded(t *testing.T) {
	tags := make(map[string]string)

	foo := &testResource{time.Now(), tags}

	if LifetimeExceeded()(foo) {
		t.Error("Resource doesn't have tag")
	}

	tags[LifetimeTagKey] = "days-5"

	oldTime := time.Now().AddDate(0, 0, -6)

	foo.creationTime = oldTime
	foo.tags = tags

	if !LifetimeExceeded()(foo) {
		t.Error("Lifetime should be exceeded")
	}

	foo.tags[LifetimeTagKey] = "invalidtag"

	if LifetimeExceeded()(foo) {
		t.Error("Tag value is malformed")
	}

	foo.tags[LifetimeTagKey] = "days-five"

	if LifetimeExceeded()(foo) {
		t.Error("Tag value is malformed")
	}

	foo.tags[LifetimeTagKey] = "days-7"

	if LifetimeExceeded()(foo) {
		t.Error("Lifetime is not exceeded")
	}
}

func TestExpiryPassed(t *testing.T) {
	tags := make(map[string]string)

	foo := &testResource{time.Now(), tags}

	if ExpiryDatePassed()(foo) {
		t.Error("Resource have no expiry tag")
	}

	foo.tags[ExpiryTagKey] = time.Now().AddDate(0, 0, -5).Format("2006-01-02")

	if !ExpiryDatePassed()(foo) {
		t.Error("Expiry should have passed")
	}

	foo.tags[ExpiryTagKey] = "malformed-tag"

	if ExpiryDatePassed()(foo) {
		t.Error("Tag is malformed")
	}

	foo.tags[ExpiryTagKey] = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	if ExpiryDatePassed()(foo) {
		t.Error("Resource is not expired")
	}
}

func TestDeleteWithin(t *testing.T) {
	deleteTime := time.Now().AddDate(0, 0, 2).Format(time.RFC3339)
	tags := make(map[string]string)
	foo := &testResource{time.Now(), tags}

	if DeleteWithinXHours(72)(foo) {
		t.Error("Resource has no delete tag")
	}

	foo.tags[DeleteTagKey] = deleteTime

	if !DeleteWithinXHours(72)(foo) {
		t.Error("Should be deleted within 72 hours")
	}

	if DeleteWithinXHours(5)(foo) {
		t.Error("Should not be deleted within 5 hours")
	}

	foo.tags[DeleteTagKey] = "malformed"

	if DeleteWithinXHours(72)(foo) {
		t.Error("Tag is malformed")
	}
}

func TestDeletePassed(t *testing.T) {
	deleteTime := time.Now().AddDate(0, 0, -2).Format(time.RFC3339)
	tags := make(map[string]string)
	foo := &testResource{time.Now(), tags}

	if DeleteAtPassed()(foo) {
		t.Error("Resource has no delete tag")
	}

	foo.tags[DeleteTagKey] = deleteTime

	if !DeleteAtPassed()(foo) {
		t.Error("Delete time should be passed")
	}

	foo.tags[DeleteTagKey] = time.Now().AddDate(0, 0, 2).Format(time.RFC3339)

	if DeleteAtPassed()(foo) {
		t.Error("Delete time is not passed")
	}

	foo.tags[DeleteTagKey] = "malformed"

	if DeleteAtPassed()(foo) {
		t.Error("Malformed tag value")
	}
}

type testVolume struct {
	testResource
	attached bool
}

func (v *testVolume) SizeGB() int64      { return testSize }
func (v *testVolume) Attached() bool     { return v.attached }
func (v *testVolume) Encrypted() bool    { return testEncrypted }
func (v *testVolume) VolumeType() string { return testVolumeType }

func TestAttached(t *testing.T) {
	foo := &testVolume{
		testResource{time.Now(), map[string]string{}},
		false,
	}

	foo.attached = true

	if IsUnattached()(foo) {
		t.Error("Should be attached")
	}

	foo.attached = false

	if !IsUnattached()(foo) {
		t.Error("Should not be attached")
	}
}

type testBucket struct {
	testResource
	lastModified time.Time
}

func (b *testBucket) LastModified() time.Time                { return b.lastModified }
func (b *testBucket) ObjectCount() int64                     { return 10 }
func (b *testBucket) TotalSizeGB() float64                   { return 5.13 }
func (b *testBucket) StorageTypeSizesGB() map[string]float64 { return make(map[string]float64) }

func TestNotModified(t *testing.T) {
	foo := &testBucket{
		testResource{time.Now(), map[string]string{}},
		time.Now(),
	}

	if NotModifiedInXDays(5)(foo) {
		t.Error("Has been modified within 5 days")
	}

	foo.lastModified = time.Now().AddDate(0, -5, 0)

	if !NotModifiedInXDays(5)(foo) {
		t.Error("Not modified within 5 days")
	}
}

type testSnap struct {
	testResource
	inUse bool
}

func (s *testSnap) Encrypted() bool { return false }
func (s *testSnap) SizeGB() int64   { return 5 }
func (s *testSnap) InUse() bool     { return s.inUse }

func TestInUse(t *testing.T) {
	foo := &testSnap{
		testResource{time.Now(), map[string]string{}},
		false,
	}

	if IsInUse()(foo) {
		t.Error("Snapshot is not in use")
	}

	foo.inUse = true

	if IsNotInUse()(foo) {
		t.Error("Snapshot is in use")
	}
}
