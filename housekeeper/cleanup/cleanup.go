// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cleanup

import (
	"brkt/cloudsweeper/cloud"
	"brkt/cloudsweeper/cloud/billing"
	"brkt/cloudsweeper/cloud/filter"
	"log"
	"time"
)

const (
	releaseTag          = "Release"
	sharedDevAWSAccount = "164337164081"
	totalCostThreshold  = 10.0
)

// MarkForCleanup will look for resources that should be automatically
// cleaned up. These resources are not deleted directly, but are given
// a tag that will delete the resources 4 days from now. The rules
// for marking a resource for cleanup are the following:
// 		- unattached volumes > 30 days old
//		- unused/unaccessed buckets > 120 days old
// 		- non-whitelisted AMIs > 6 months
// 		- non-whitelisted snapshots > 6 months
// 		- non-whitelisted volumes > 6 months
//		- untagged resources > 30 days (this should take care of instances)
func MarkForCleanup(mngr cloud.ResourceManager) {
	allResources := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()

	for owner, res := range allResources {
		log.Println("Marking resources for cleanup in", owner)
		untaggedFilter := filter.New()
		untaggedFilter.AddGeneralRule(func(r cloud.Resource) bool {
			return len(r.Tags()) == 0
		})
		untaggedFilter.AddGeneralRule(filter.OlderThanXDays(30))
		untaggedFilter.AddSnapshotRule(filter.IsNotInUse())
		untaggedFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		oldFilter := filter.New()
		oldFilter.AddGeneralRule(filter.OlderThanXMonths(6))
		// Don't cleanup resources tagged for release
		oldFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		oldFilter.AddSnapshotRule(filter.IsNotInUse())
		oldFilter.AddVolumeRule(filter.IsUnattached())
		oldFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		unattachedFilter := filter.New()
		unattachedFilter.AddVolumeRule(filter.IsUnattached())
		unattachedFilter.AddGeneralRule(filter.OlderThanXDays(30))
		unattachedFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		unattachedFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		bucketFilter := filter.New()
		bucketFilter.AddBucketRule(filter.NotModifiedInXDays(120))
		bucketFilter.AddGeneralRule(filter.OlderThanXDays(7))
		bucketFilter.AddGeneralRule(filter.Negate(filter.HasTag(releaseTag)))
		bucketFilter.AddGeneralRule(filter.Negate(filter.TaggedForCleanup()))

		timeToDelete := time.Now().AddDate(0, 0, 4)

		resourcesToTag := []cloud.Resource{}
		totalCost := 0.0

		// Tag instances
		for _, res := range filter.Instances(res.Instances, untaggedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag volumes
		for _, res := range filter.Volumes(res.Volumes, oldFilter, unattachedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag snapshots
		for _, res := range filter.Snapshots(res.Snapshots, oldFilter, untaggedFilter) {
			resourcesToTag = append(resourcesToTag, res)
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			totalCost += days * costPerDay
		}

		// Tag images
		for _, res := range filter.Images(res.Images, oldFilter, untaggedFilter) {
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

	// This will cleanup old released AMIs if they're older than a year
	cleanupReleaseImagesAWS()
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

// This function will look for released images. If the image is older
// than 6 months they will be made private and set to be de-registered
// after another 6 months have passed
func cleanupReleaseImagesAWS() {
	mngr, err := cloud.NewManager(cloud.AWS, sharedDevAWSAccount)
	if err != nil {
		log.Printf("Could not initalize resource manager for release image cleanup: %s", err)
		return
	}
	allImages := mngr.ImagesPerAccount()
	for owner, images := range allImages {
		log.Println("Performing release image cleanup in", owner)
		err := cleanupReleaseImagesHelper(mngr, images)
		if err != nil {
			log.Printf("Release cleanup for \"%s\" failed:\n%s\n", owner, err)
		}
	}
}

func cleanupReleaseImagesHelper(mngr cloud.ResourceManager, images []cloud.Image) error {
	// First find public images older than 6 months. Make these images
	// private and add an expiry to them.
	filPub := filter.New()
	filPub.AddGeneralRule(filter.HasTag(releaseTag))
	filPub.AddGeneralRule(filter.IsPublic())
	filPub.AddGeneralRule(filter.OlderThanXMonths(6))
	// Images shoudln't have an expiry tag already
	filPub.AddGeneralRule(filter.Negate(filter.HasTag(filter.ExpiryTagKey)))
	imagesToMakePrivate := filter.Images(images, filPub)
	// Make images private and add expiry tag
	for i := range imagesToMakePrivate {
		err := imagesToMakePrivate[i].MakePrivate()

		if err != nil {
			log.Printf("Failed to make release image '%s' private\n", imagesToMakePrivate[i].ID())
			return err
		}
		expiryDateString := time.Now().AddDate(0, 6, 0).Format(filter.ExpiryTagValueFormat)
		err = imagesToMakePrivate[i].SetTag(filter.ExpiryTagKey, expiryDateString, true)
		if err != nil {
			log.Println("Failed to set expiry tag on image", imagesToMakePrivate[i].ID())
			return err
		}
	}

	// Then check for private images that are expired
	filPriv := filter.New()
	filPriv.AddGeneralRule(filter.HasTag(releaseTag))
	filPriv.AddGeneralRule(filter.Negate(filter.IsPublic()))
	filPriv.AddGeneralRule(filter.ExpiryDatePassed())
	imagesToCleanup := filter.Images(images, filPriv)
	// Cleanup expired images
	err := mngr.CleanupImages(imagesToCleanup)
	if err != nil {
		log.Println("Failed to cleanup expired release images")
		return err
	}
	return nil
}

// ResetHousekeeper will remove any cleanup tags existing in the accounts
// associated with the provided resource manager
func ResetHousekeeper(mngr cloud.ResourceManager) {
	allResources := mngr.AllResourcesPerAccount()

	for owner, res := range allResources {
		log.Println("Resetting housekeeper tags in", owner)
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
