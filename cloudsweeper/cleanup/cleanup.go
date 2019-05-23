// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cleanup

import (
	"log"
	"time"

	"github.com/cloudtools/cloudsweeper/cloud"
	"github.com/cloudtools/cloudsweeper/cloud/billing"
	"github.com/cloudtools/cloudsweeper/cloud/filter"
)

const (
	releaseTag         = "Release"
	totalCostThreshold = 10.0
)

// MarkForCleanup will look for resources that should be automatically
// cleaned up. These resources are not deleted directly, but are given
// a tag that will delete the resources 4 days from now. The rules
// for marking a resource for cleanup are the following:
// 		- unattached volumes > 30 days old
//		- unused/unaccessed buckets > 6 months (182 days)
// 		- non-whitelisted AMIs > 6 months
// 		- non-whitelisted snapshots > 6 months
// 		- non-whitelisted volumes > 6 months
//		- untagged resources > 30 days (this should take care of instances)
func MarkForCleanup(mngr cloud.ResourceManager, thresholds map[string]int) {
	allResources := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()

	for owner, res := range allResources {
		log.Println("Marking resources for cleanup in", owner)
		untaggedFilter := filter.New()
		untaggedFilter.AddGeneralRule(func(r cloud.Resource) bool {
			return len(r.Tags()) == 0
		})
		untaggedFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-untagged-older-than-days"]))
		untaggedFilter.AddSnapshotRule(filter.IsNotInUse())
		untaggedFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		instanceFilter := filter.New()
		instanceFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-instances-older-than-days"]))
		instanceFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		instanceFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		snapshotFilter := filter.New()
		instanceFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-snapshots-older-than-days"]))
		snapshotFilter.AddSnapshotRule(filter.IsNotInUse())
		snapshotFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		snapshotFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		imageFilter := filter.New()
		imageFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-images-older-than-days"]))
		imageFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		imageFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		volumeFilter := filter.New()
		volumeFilter.AddVolumeRule(filter.IsUnattached())
		volumeFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-unattatched-older-than-days"]))
		volumeFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		volumeFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		bucketFilter := filter.New()
		bucketFilter.AddBucketRule(filter.NotModifiedInXDays(thresholds["clean-bucket-not-modified-days"]))
		bucketFilter.AddGeneralRule(filter.OlderThanXDays(thresholds["clean-bucket-older-than-days"]))
		bucketFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		bucketFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		timeToDelete := time.Now().AddDate(0, 0, 4)

		resourcesToTag := []cloud.Resource{}
		totalCost := 0.0

		// Tag instances
		for _, res := range filter.Instances(res.Instances, instanceFilter, untaggedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag volumes
		for _, res := range filter.Volumes(res.Volumes, volumeFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag snapshots
		for _, res := range filter.Snapshots(res.Snapshots, snapshotFilter, untaggedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag images
		for _, res := range filter.Images(res.Images, imageFilter, untaggedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		if buck, ok := allBuckets[owner]; ok {
			for _, res := range filter.Buckets(buck, bucketFilter) {
				resourcesToTag = append(resourcesToTag, res)
				totalCost += billing.BucketPricePerMonth(res)
			}
		}

		if totalCost >= totalCostThreshold {
			for _, res := range resourcesToTag {
				err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
				if err != nil {
					log.Printf("%s: Failed to tag %s for deletion: %s\n", owner, res.ID(), err)
				} else {
					log.Printf("%s: Marked %s for deletion at %s\n", owner, res.ID(), timeToDelete)
				}
			}
		} else {
			log.Printf("%s: Skipping the tagging of resources, total cost $%.2f is less than $%.2f", owner, totalCost, totalCostThreshold)
		}
	}
}

// PerformCleanup will run different cleanup functions which all
// do some sort of rule based cleanup
func PerformCleanup(mngr cloud.ResourceManager) {
	// Cleanup all resources with a lifetime tag that has passed. This
	// includes both the lifetime and the expiry tag
	cleanupLifetimePassed(mngr)
}

func cleanupLifetimePassed(mngr cloud.ResourceManager) {
	allResources := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	for owner, resources := range allResources {
		log.Println("Performing lifetime check in", owner)
		lifetimeFilter := filter.New()
		lifetimeFilter.AddGeneralRule(filter.LifetimeExceeded())

		expiryFilter := filter.New()
		expiryFilter.AddGeneralRule(filter.ExpiryDatePassed())

		deleteAtFilter := filter.New()
		deleteAtFilter.AddGeneralRule(filter.DeleteAtPassed())

		err := mngr.CleanupInstances(filter.Instances(resources.Instances, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup instances in %s, err:\n%s", owner, err)
		}
		err = mngr.CleanupImages(filter.Images(resources.Images, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup images in %s, err:\n%s", owner, err)
		}
		err = mngr.CleanupVolumes(filter.Volumes(resources.Volumes, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup volumes in %s, err:\n%s", owner, err)
		}
		err = mngr.CleanupSnapshots(filter.Snapshots(resources.Snapshots, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup snapshots in %s, err:\n%s", owner, err)
		}
		if bucks, ok := allBuckets[owner]; ok {
			err = mngr.CleanupBuckets(filter.Buckets(bucks, lifetimeFilter, expiryFilter, deleteAtFilter))
			if err != nil {
				log.Printf("Could not cleanup buckets in %s, err:\n%s", owner, err)
			}
		}
	}
}

// ResetCloudsweeper will remove any cleanup tags existing in the accounts
// associated with the provided resource manager
func ResetCloudsweeper(mngr cloud.ResourceManager) {
	allResources := mngr.AllResourcesPerAccount()

	for owner, res := range allResources {
		log.Println("Resetting Cloudsweeper tags in", owner)
		taggedFilter := filter.New()
		taggedFilter.AddGeneralRule(filter.HasTag(filter.DeleteTagKey))

		// Un-Tag instances
		for _, res := range filter.Instances(res.Instances, taggedFilter) {
			err := res.RemoveTag(filter.DeleteTagKey)
			if err != nil {
				log.Printf("Failed to remove tag on %s: %s\n", res.ID(), err)
			} else {
				log.Printf("Removed cleanup tag on %s\n", res.ID())
			}
		}

		// Un-Tag volumes
		for _, res := range filter.Volumes(res.Volumes, taggedFilter) {
			err := res.RemoveTag(filter.DeleteTagKey)
			if err != nil {
				log.Printf("Failed to remove tag on %s: %s\n", res.ID(), err)
			} else {
				log.Printf("Removed cleanup tag on %s\n", res.ID())
			}
		}

		// Un-Tag snapshots
		for _, res := range filter.Snapshots(res.Snapshots, taggedFilter) {
			err := res.RemoveTag(filter.DeleteTagKey)
			if err != nil {
				log.Printf("Failed to remove tag on %s: %s\n", res.ID(), err)
			} else {
				log.Printf("Removed cleanup tag on %s\n", res.ID())
			}
		}

		// Un-Tag images
		for _, res := range filter.Images(res.Images, taggedFilter) {
			err := res.RemoveTag(filter.DeleteTagKey)
			if err != nil {
				log.Printf("Failed to remove tag on %s: %s\n", res.ID(), err)
			} else {
				log.Printf("Removed cleanup tag on %s\n", res.ID())
			}
		}
	}
}
