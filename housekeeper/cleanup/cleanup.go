package cleanup

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/filter"
	"brkt/housekeeper/housekeeper"
	"log"
	"time"
)

const (
	releaseTag = "Release"

	sharedDevAWSAccount = "164337164081"
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
func MarkForCleanup(csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	allResources := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()

	for owner, res := range allResources {
		log.Println("Marking resources for cleanup in", owner)
		untaggedFilter := filter.New()
		untaggedFilter.AddGeneralRule(func(r cloud.Resource) bool {
			return len(r.Tags()) == 0
		})

		oldFilter := filter.New()
		oldFilter.AddGeneralRule(filter.OlderThanXMonths(6))

		unattachedFilter := filter.New()
		unattachedFilter.AddVolumeRule(filter.IsUnattached())

		bucketFilter := filter.New()
		bucketFilter.AddBucketRule(filter.NotModifiedInXDays(120))

		timeToDelete := time.Now().AddDate(0, 0, 4)

		// Tag instances
		for _, res := range filter.Instances(res.Instances, untaggedFilter) {
			err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
			if err != nil {
				log.Printf("Failed to tag %s for deletion: %s\n", res.ID(), err)
			} else {
				log.Printf("Marked %s for deletion at %s\n", res.ID(), timeToDelete)
			}
		}

		// Tag volumes
		for _, res := range filter.Volumes(res.Volumes, oldFilter, unattachedFilter) {
			err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
			if err != nil {
				log.Printf("Failed to tag %s for deletion: %s\n", res.ID(), err)
			} else {
				log.Printf("Marked %s for deletion at %s\n", res.ID(), timeToDelete)
			}
		}

		// Tag snapshots
		for _, res := range filter.Snapshots(res.Snapshots, oldFilter, unattachedFilter) {
			err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
			if err != nil {
				log.Printf("Failed to tag %s for deletion: %s\n", res.ID(), err)
			} else {
				log.Printf("Marked %s for deletion at %s\n", res.ID(), timeToDelete)
			}
		}

		// Tag images
		for _, res := range filter.Images(res.Images, oldFilter, unattachedFilter) {
			err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
			if err != nil {
				log.Printf("Failed to tag %s for deletion: %s\n", res.ID(), err)
			} else {
				log.Printf("Marked %s for deletion at %s\n", res.ID(), timeToDelete)
			}
		}

		if buck, ok := allBuckets[owner]; ok {
			for _, res := range filter.Buckets(buck, oldFilter, unattachedFilter) {
				err := res.SetTag(filter.DeleteTagKey, timeToDelete.Format(time.RFC3339), true)
				if err != nil {
					log.Printf("Failed to tag %s for deletion: %s\n", res.ID(), err)
				} else {
					log.Printf("Marked %s for deletion at %s\n", res.ID(), timeToDelete)
				}
			}
		}
	}
}

// PerformCleanup will run different cleanup functions which all
// do some sort of rule based cleanup
func PerformCleanup(csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	// Cleanup all resources with a lifetime tag that has passed. This
	// includes both the lifetime and the expiry tag
	cleanupLifetimePassed(mngr)

	// This will cleanup old released AMIs if they're older than a year
	cleanupReleaseImages(csp)
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
			continue
		}
		err = mngr.CleanupImages(filter.Images(resources.Images, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup images in %s, err:\n%s", owner, err)
			continue
		}
		err = mngr.CleanupVolumes(filter.Volumes(resources.Volumes, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup volumes in %s, err:\n%s", owner, err)
			continue
		}
		err = mngr.CleanupSnapshots(filter.Snapshots(resources.Snapshots, lifetimeFilter, expiryFilter, deleteAtFilter))
		if err != nil {
			log.Printf("Could not cleanup snapshots in %s, err:\n%s", owner, err)
			continue
		}
		if bucks, ok := allBuckets[owner]; ok {
			err = mngr.CleanupBuckets(filter.Buckets(bucks, lifetimeFilter, expiryFilter, deleteAtFilter))
			if err != nil {
				log.Printf("Could not cleanup buckets in %s, err:\n%s", owner, err)
				continue
			}
		}
	}
}

// This function will look for released images. If the image is older
// than 6 months they will be made private and set to be de-registered
// after another 6 months have passed
func cleanupReleaseImages(csp cloud.CSP) {
	// TODO: Change when GCP is supported
	mngr := cloud.NewManager(csp, sharedDevAWSAccount)
	allImages := mngr.ImagesPerAccount()
	for owner, images := range allImages {
		log.Println("Performing release image cleanup in", owner)
		err := cleanCleanupReleaseImagesHelper(mngr, images)
		if err != nil {
			log.Printf("Release cleanup for \"%s\" failed:\n%s\n", owner, err)
		}
	}
}

func cleanCleanupReleaseImagesHelper(mngr cloud.ResourceManager, images []cloud.Image) error {
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
