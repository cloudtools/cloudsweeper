// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package notify

import (
	"brkt/cloudsweeper/cloud"
	"brkt/cloudsweeper/cloud/billing"
	"brkt/cloudsweeper/cloud/filter"
	hk "brkt/cloudsweeper/housekeeper"
	"fmt"
	"log"
	"time"
)

// Here is where you use credentials to send email
// note that monthToDateAddressee is intended to be sent weekly
// to your entire org.  
// the totalSumAddressee is meant to send a total report to
// the person in your org monitoring costs

const (
	smtpUserKey          = "SMTP_USER"
	smtpPassKey          = "SMTP_PASS"
	mailDisplayName      = "HouseKeeper"
	monthToDateAddressee = "eng@example.com"
	totalSumAddressee    = "ben"
)

type resourceMailData struct {
	Owner          string
	OwnerID        string
	Instances      []cloud.Instance
	Images         []cloud.Image
	Snapshots      []cloud.Snapshot
	Volumes        []cloud.Volume
	Buckets        []cloud.Bucket
	HoursInAdvance int
}

func (d *resourceMailData) ResourceCount() int {
	return len(d.Images) + len(d.Instances) + len(d.Snapshots) + len(d.Volumes) + len(d.Buckets)
}

func (d *resourceMailData) SendEmail(mailTemplate, title string, debugAddressees ...string) {
	mailClient := getMailClient()
	mailContent, err := generateMail(d, mailTemplate)
	if err != nil {
		log.Fatalln("Could not generate email:", err)
	}
	username := convertEmailExceptions(d.Owner)

// Insert a domain name into the usernames in the organization json file
	ownerMail := fmt.Sprintf("%s@.example.com", username)
	log.Printf("Sending out email to %s\n", ownerMail)
	addressees := append(debugAddressees, ownerMail)
	err = mailClient.SendEmail(title, mailContent, addressees...)
	if err != nil {
		log.Fatalf("Failed to email %s: %s\n", ownerMail, err)
	}
}

type monthToDateData struct {
	CSP              cloud.CSP
	TotalCost        float64
	SortedUsers      billing.UserList
	MinimumTotalCost float64
	MinimumCost      float64
	AccountToUser    map[string]string
}

func initTotalSummaryMailData() *resourceMailData {
	return &resourceMailData{
		Owner:     totalSumAddressee,
		Instances: []cloud.Instance{},
		Images:    []cloud.Image{},
		Snapshots: []cloud.Snapshot{},
		Volumes:   []cloud.Volume{},
		Buckets:   []cloud.Bucket{},
	}
}

func initManagerToMailDataMapping(managers hk.Employees) map[string]*resourceMailData {
	result := make(map[string]*resourceMailData)
	for _, manager := range managers {
		result[manager.Username] = &resourceMailData{
			Owner:     manager.Username,
			Instances: []cloud.Instance{},
			Images:    []cloud.Image{},
			Snapshots: []cloud.Snapshot{},
			Volumes:   []cloud.Volume{},
			Buckets:   []cloud.Bucket{},
		}
	}
	return result
}

// OldResourceReview will review (but not do any cleanup action) old resources
// that an owner might want to consider doing something about. The owner is then
// sent an email with a list of these resources. Resources are sent for review
// if they fulfil any of the following rules:
//		- Resource is older than 30 days
//		- A whitelisted resource is older than 6 months
//		- An instance marked with do-not-delete is older than a week
func OldResourceReview(mngr cloud.ResourceManager, org *hk.Organization, csp cloud.CSP) {
	allCompute := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	accountUserMapping := org.AccountToUserMapping(csp)
	userEmployeeMapping := org.UsernameToEmployeeMapping()
	totalSummaryMailData := initTotalSummaryMailData()
	managerToMailDataMapping := initManagerToMailDataMapping(org.Managers)

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

	for account, resources := range allCompute {
		log.Println("Performing old resource review in", account)
		username := accountUserMapping[account]
		employee := userEmployeeMapping[username]

		// Apply filters
		userMailData := resourceMailData{
			Owner:     username,
			Instances: filter.Instances(resources.Instances, generalFilter, whitelistFilter, dndFilter, dndFilter2),
			Images:    filter.Images(resources.Images, generalFilter, whitelistFilter),
			Volumes:   filter.Volumes(resources.Volumes, generalFilter, whitelistFilter),
			Snapshots: filter.Snapshots(resources.Snapshots, generalFilter, whitelistFilter),
			Buckets:   []cloud.Bucket{},
		}
		if buckets, ok := allBuckets[account]; ok {
			userMailData.Buckets = filter.Buckets(buckets, generalFilter, whitelistFilter)
		}

		// Add to the manager summary
		if managerSummaryMailData, ok := managerToMailDataMapping[employee.Manager.Username]; ok { // safe or org _should_ have thrown an error
			managerSummaryMailData.Instances = append(managerSummaryMailData.Instances, userMailData.Instances...)
			managerSummaryMailData.Images = append(managerSummaryMailData.Images, userMailData.Images...)
			managerSummaryMailData.Snapshots = append(managerSummaryMailData.Snapshots, userMailData.Snapshots...)
			managerSummaryMailData.Volumes = append(managerSummaryMailData.Volumes, userMailData.Volumes...)
			managerSummaryMailData.Buckets = append(managerSummaryMailData.Buckets, userMailData.Buckets...)
		} else {
			log.Fatalf("%s is not a manager??? Verify `organization.go` and the org repo itself for issues", employee.Manager.Username)
		}

		// Add to the total summary
		totalSummaryMailData.Instances = append(totalSummaryMailData.Instances, userMailData.Instances...)
		totalSummaryMailData.Images = append(totalSummaryMailData.Images, userMailData.Images...)
		totalSummaryMailData.Snapshots = append(totalSummaryMailData.Snapshots, userMailData.Snapshots...)
		totalSummaryMailData.Volumes = append(totalSummaryMailData.Volumes, userMailData.Volumes...)
		totalSummaryMailData.Buckets = append(totalSummaryMailData.Buckets, userMailData.Buckets...)

		if userMailData.ResourceCount() > 0 {
			title := fmt.Sprintf("You have %d old resources to review (%s)", userMailData.ResourceCount(), time.Now().Format("2006-01-02"))
			userMailData.SendEmail(reviewMailTemplate, title)
		}
	}

	// Send out manager emails
	for username, managerSummaryMailData := range managerToMailDataMapping {
		log.Printf("Collecting old resources to review for %s's team\n", username)
		if managerSummaryMailData.ResourceCount() > 0 {
			title := fmt.Sprintf("Your team has %d old resources to review (%s)", managerSummaryMailData.ResourceCount(), time.Now().Format("2006-01-02"))
			managerSummaryMailData.SendEmail(managerReviewMailTemplate, title)
		}
	}

	// Send out a total summary
	log.Println("Collecting old resource review for the org")
	title := fmt.Sprintf("Your org has %d old resources to review (%s)", totalSummaryMailData.ResourceCount(), time.Now().Format("2006-01-02"))
	totalSummaryMailData.SendEmail(totalReviewMailTemplate, title)
}

// UntaggedResourcesReview will look for resources without any tags, and
// send out a mail encouraging to tag tag them
func UntaggedResourcesReview(mngr cloud.ResourceManager, accountUserMapping map[string]string) {
	// We only care about untagged resources in EC2
	allCompute := mngr.AllResourcesPerAccount()
	for account, resources := range allCompute {
		log.Printf("Performing untagged resources review in %s", account)
		untaggedFilter := filter.New()
		untaggedFilter.AddGeneralRule(filter.Negate(filter.HasTag("Name")))

		// We care about un-tagged whitelisted resources too
		untaggedFilter.OverrideWhitelist = true

		username := accountUserMapping[account]
		mailData := resourceMailData{
			Owner:     username,
			OwnerID:   account,
			Instances: filter.Instances(resources.Instances, untaggedFilter),
			// Only report on instances for now
			//Images:    filter.Images(resources.Images, untaggedFilter),
			//Snapshots: filter.Snapshots(resources.Snapshots, untaggedFilter),
			//Volumes:   filter.Volumes(resources.Volumes, untaggedFilter),
			Buckets: []cloud.Bucket{},
		}

		if mailData.ResourceCount() > 0 {
			// Send mail
			// title := fmt.Sprintf("You have %d un-tagged resources to review (%s)", mailData.ResourceCount(), time.Now().Format("2006-01-02"))
			// You can add some debug email address to ensure it works
			// debugAddressees := []string{"ben@example.com"} 
			// mailData.SendEmail(untaggedMailTemplate, title, debugAddressees...)
		}
	}
}

// DeletionWarning will find resources which are about to be deleted within
// `hoursInAdvance` hours, and send an email to the owner of those resources
// with a warning. Resources explicitly tagged to be deleted are not included
// in this warning.
func DeletionWarning(hoursInAdvance int, mngr cloud.ResourceManager, accountUserMapping map[string]string) {
	allCompute := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	for account, resources := range allCompute {
		ownerName := convertEmailExceptions(accountUserMapping[account])
		fil := filter.New()
		fil.AddGeneralRule(filter.DeleteWithinXHours(hoursInAdvance))
		mailData := resourceMailData{
			ownerName,
			account,
			filter.Instances(resources.Instances, fil),
			filter.Images(resources.Images, fil),
			filter.Snapshots(resources.Snapshots, fil),
			filter.Volumes(resources.Volumes, fil),
			[]cloud.Bucket{},
			hoursInAdvance,
		}
		if buckets, ok := allBuckets[account]; ok {
			mailData.Buckets = filter.Buckets(buckets, fil)
		}

		if mailData.ResourceCount() > 0 {
			// Now send email
			// title := fmt.Sprintf("Deletion warning, %d resources are cleaned up within %d hours", mailData.ResourceCount(), hoursInAdvance)
			// debugAddressees := []string{"ben@example.com"} 
			// mailData.SendEmail(deletionWarningTemplate, title, debugAddressees...)
		}
	}
}

// MonthToDateReport sends an email to engineering with the
// Month-to-Date billing report
func MonthToDateReport(report billing.Report, accountUserMapping map[string]string) {
	mailClient := getMailClient()
	reportData := monthToDateData{report.CSP, report.TotalCost(), report.SortedUsersByTotalCost(), billing.MinimumTotalCost, billing.MinimumCost, accountUserMapping}
	mailContent, err := generateMail(reportData, monthToDateTemplate)
	if err != nil {
		log.Fatalln("Could not generate email:", err)
	}
	log.Printf("Sending the Month-to-date report to %s\n", monthToDateAddressee)
	title := fmt.Sprintf("Month-to-date %s billing report", report.CSP)
	err = mailClient.SendEmail(title, mailContent, monthToDateAddressee)
	if err != nil {
		log.Printf("Failed to email %s: %s\n", monthToDateAddressee, err)
	}
}
