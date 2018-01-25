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

// PerformCleanup will run different cleanup functions which all
// do some sort of rule based cleanup
func PerformCleanup(csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	// Cleanup all resources with a lifetime tag that has passed
	cleanupLifetimePassed(mngr)

	// This will cleanup old released AMIs if they're older than a year
	cleanupReleaseImages(csp)
}

func cleanupLifetimePassed(mngr cloud.ResourceManager) {
	allResources := mngr.AllResourcesPerAccount()
	for owner, resources := range allResources {
		log.Println("Performing lifetime check in", owner)
		fil := filter.New()
		fil.AddGeneralRule(filter.LifetimeExceeded())
		err := mngr.CleanupInstances(filter.Instances(resources.Instances, fil))
		if err != nil {
			log.Printf("Could not cleanup instances in %s, err:\n%s", owner, err)
			continue
		}
		err = mngr.CleanupImages(filter.Images(resources.Images, fil))
		if err != nil {
			log.Printf("Could not cleanup images in %s, err:\n%s", owner, err)
			continue
		}
		err = mngr.CleanupVolumes(filter.Volumes(resources.Volumes, fil))
		if err != nil {
			log.Printf("Could not cleanup volumes in %s, err:\n%s", owner, err)
			continue
		}
		err = mngr.CleanupSnapshots(filter.Snapshots(resources.Snapshots, fil))
		if err != nil {
			log.Printf("Could not cleanup snapshots in %s, err:\n%s", owner, err)
			continue
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
