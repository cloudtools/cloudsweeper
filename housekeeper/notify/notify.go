package notify

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/filter"
	"brkt/housekeeper/housekeeper"
	"fmt"
	"log"
	"time"
)

const (
	smtpUserKey     = "SMTP_USER"
	smtpPassKey     = "SMTP_PASS"
	mailDisplayName = "HouseKeeper"
)

type resourceMailData struct {
	Owner     string
	Instances []cloud.Instance
	Images    []cloud.Image
	Snapshots []cloud.Snapshot
	Volumes   []cloud.Volume
	Buckets   []cloud.Bucket
}

func (d *resourceMailData) ResourceCount() int {
	return len(d.Images) + len(d.Instances) + len(d.Snapshots) + len(d.Volumes) + len(d.Buckets)
}

// OldResourceReview will review (but not do any cleanup action) old resources
// that an owner might want to consider doing something about. The owner is then
// sent an email with a list of these resources. Resources are sent for review
// if they fulfil any of the following rules:
//		- Resource is older than 30 days
//		- A whitelisted resource is older than 6 months
//		- An instance marked with do-not-delete is older than a week
func OldResourceReview(csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	allCompute := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	ownerNames := owners.IDToName()
	for owner, resources := range allCompute {
		log.Println("Performing old resource review in", owner)
		ownerName := convertEmailExceptions(ownerNames[owner])

		// Create filters
		generalFilter := filter.New()
		generalFilter.AddGeneralRule(filter.OlderThanXDays(30))

		whitelistFilter := filter.New()
		whitelistFilter.OverrideWhitelist = true
		whitelistFilter.AddGeneralRule(filter.OlderThanXMonths(6))

		// These only apply to instances
		dndFilter := filter.New()
		dndFilter.AddGeneralRule(filter.HasTag("no-not-delete"))
		dndFilter.AddGeneralRule(filter.OlderThanXDays(7))

		dndFilter2 := filter.New()
		dndFilter2.AddGeneralRule(filter.NameContains("do-not-delete"))
		dndFilter2.AddGeneralRule(filter.OlderThanXDays(7))

		// Apply filters
		mailHolder := resourceMailData{
			Owner:     ownerName,
			Instances: filter.Instances(resources.Instances, generalFilter, whitelistFilter, dndFilter, dndFilter2),
			Images:    filter.Images(resources.Images, generalFilter, whitelistFilter),
			Volumes:   filter.Volumes(resources.Volumes, generalFilter, whitelistFilter),
			Snapshots: filter.Snapshots(resources.Snapshots, generalFilter, whitelistFilter),
			Buckets:   []cloud.Bucket{},
		}
		if bucks, ok := allBuckets[owner]; ok {
			mailHolder.Buckets = filter.Buckets(bucks, generalFilter, whitelistFilter)
		}

		if mailHolder.ResourceCount() > 0 {
			// Now send email
			mailClient := getMailClient()
			mailContent, err := generateMail(mailHolder, reviewMailTemplate)
			if err != nil {
				log.Fatalln("Could not generate email:", err)
			}
			ownerMail := fmt.Sprintf("%s@brkt.com", mailHolder.Owner)
			log.Printf("Sending out old resource review to %s\n", ownerMail)
			title := fmt.Sprintf("You have %d old resources to review (%s)", mailHolder.ResourceCount(), time.Now().Format("2006-01-02"))
			err = mailClient.SendEmail(title, mailContent, "hsson@brkt.com") // TODO: Use correct email
			if err != nil {
				log.Printf("Failed to email %s: %s\n", ownerMail, err)
			}
		}
	}
}

// DeletionWarning will find resources which are about to be deleted within
// `hoursInAdvance` hours, and send an email to the owner of those resources
// with a warning. Resources explicitly tagged to be deleted are not included
// in this warning.
func DeletionWarning(hoursInAdvance int, csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	allCompute := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	ownerNames := owners.IDToName()
	for owner, resources := range allCompute {
		ownerName := convertEmailExceptions(ownerNames[owner])
		fil := filter.New()
		fil.AddGeneralRule(filter.DeleteWithinXHours(hoursInAdvance))
		mailHolder := struct {
			resourceMailData
			Hours int
		}{
			resourceMailData{
				ownerName,
				filter.Instances(resources.Instances, fil),
				filter.Images(resources.Images, fil),
				filter.Snapshots(resources.Snapshots, fil),
				filter.Volumes(resources.Volumes, fil),
				[]cloud.Bucket{},
			},
			hoursInAdvance,
		}
		if bucks, ok := allBuckets[owner]; ok {
			mailHolder.Buckets = filter.Buckets(bucks, fil)
		}

		if mailHolder.ResourceCount() > 0 {
			// Now send email
			mailClient := getMailClient()
			mailContent, err := generateMail(mailHolder, deletionWarningTemplate)
			if err != nil {
				log.Fatalln("Could not generate email:", err)
			}
			ownerMail := fmt.Sprintf("%s@brkt.com", mailHolder.Owner)
			log.Printf("Warning %s about resource deletion\n", ownerMail)
			title := fmt.Sprintf("Deletion warning, %d resources are cleaned up within %d hours", mailHolder.ResourceCount(), hoursInAdvance)
			err = mailClient.SendEmail(title, mailContent, "hsson@brkt.com") // TODO: Use correct email
			if err != nil {
				log.Printf("Failed to email %s: %s\n", ownerMail, err)
			}
		}
	}
}
