// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/agaridata/cloudsweeper/cloud"
	"github.com/agaridata/cloudsweeper/cloud/billing"
	cs "github.com/agaridata/cloudsweeper/cloudsweeper"
	"github.com/agaridata/cloudsweeper/cloudsweeper/cleanup"
	"github.com/agaridata/cloudsweeper/cloudsweeper/find"
	"github.com/agaridata/cloudsweeper/cloudsweeper/notify"
	"github.com/agaridata/cloudsweeper/cloudsweeper/setup"
)

const (
	configFileName = "config.conf"
	cspFlagAWS     = "aws"
	cspFlagGCP     = "gcp"
)

var (
	config map[string]string

	cspToUse = flag.String("csp", "", "Which CSP to run against")
	orgFile  = flag.String("org-file", "", "Specify where to find the JSON with organization information")

	awsBillingAccount      = flag.String("billing-account", "", "Specify AWS billing account id (e.g. 1234661312)")
	awsBillingBucketRegion = flag.String("billing-bucket-region", "", "Specify AWS region where --billing-bucket is location")
	gcpBillingCSVPrefix    = flag.String("billing-csv-prefix", "", "Specify name prefix of GCP billing CSV files")
	billingBucket          = flag.String("billing-bucket", "", "Specify bucket with billing CSVs")
	awsBillingSortTag      = flag.String("billing-sort-tag", "", "Specify a tag to sort on when creating report")

	mailUser     = flag.String("smtp-username", "", "SMTP username used to send email")
	mailPassword = flag.String("smtp-password", "", "SMTP password used to send email")
	mailServer   = flag.String("smtp-server", "", "SMTP server used to send mail")
	mailPort     = flag.String("smtp-port", "", "SMTP port used to send mail")

	warningHours          = flag.String("warning-hours", "", "The number of hours in advance to warn about resource deletion")
	displayName           = flag.String("display-name", "", "Name displayed on emails sent by Cloudsweeper")
	mailFrom              = flag.String("mail-from", "", "'From Email' displayed on emails sent by Cloudsweeper")
	billingReportReceiver = flag.String("billing-report-addressee", "", "Receiver of month to date billing report")
	summaryManager        = flag.String("total-sum-addressee", "", "Receiver of total cost sums")
	mailDomain            = flag.String("mail-domain", "", "The mail domain appended to usernames specified in the organization")

	setupARN = flag.String("aws-master-arn", "", "AWS ARN of role in account used by Cloudsweeper to assume roles")

	findResourceID = flag.String("resource-id", "", "ID of resource to find with find-resource command")

	dryRun = flag.Bool("marking-dry-run", false, "Whether to perform a dry run for mark and delete (nothing will actually be marked)")

	// Thresholds
	thresholds = make(map[string]int)
	thnames    = []string{
		"clean-untagged-older-than-days",
		"clean-instances-older-than-days",
		"clean-images-older-than-days",
		"clean-snapshots-older-than-days",
		"clean-unattached-older-than-days",
		"clean-bucket-not-modified-days",
		"clean-bucket-older-than-days",
		"clean-keep-n-component-images",
		"notify-untagged-older-than-days",
		"notify-instances-older-than-days",
		"notify-images-older-than-days",
		"notify-unattached-older-than-days",
		"notify-snapshots-older-than-days",
		"notify-buckets-older-than-days",
		"notify-whitelist-older-than-days",
		"notify-dnd-older-than-days",
	}

	// Clean thresholds
	cleanUntaggedOlderThanDays   = flag.String("clean-untagged-older-than-days", "", "Clean untagged resources if older than X days (default: 30)")
	cleanInstancesOlderThanDays  = flag.String("clean-instances-older-than-days", "", "Clean if instance is older than X days (default: 182)")
	cleanImagesOlderThanDays     = flag.String("clean-images-older-than-days", "", "Clean if image is older than X days (default: 182)")
	cleanSnapshotsOlderThanDays  = flag.String("clean-snapshots-older-than-days", "", "Clean if snapshot is older than X days (default: 182)")
	cleanUnattachedOlderThanDays = flag.String("clean-unattached-older-than-days", "", "Clean unattached volumes older than X days (default: 30)")
	cleanBucketNotModifiedDays   = flag.String("clean-bucket-not-modified-days", "", "Clean s3 bucket if not modified for more than X days (default: 182)")
	cleanBucketOlderThanDays     = flag.String("clean-bucket-older-than-days", "", "Clean s3 bucket if older than X days (default: 7)")
	cleanKeepNComponentImages    = flag.String("clean-keep-n-component-images", "", "Clean images with component-date naming that are older than the N most recent ones (default: 2)")

	//  Notify thresholds
	notifyUntaggedOlderThanDays  = flag.String("notify-untagged-older-than-days", "", "Notify if untagged resource is older than X days (default: 14)")
	notifyInstancesOlderThanDays = flag.String("notify-instances-older-than-days", "", "Notify if instances is older than X days (default: 30)")
	notifyImagesOlderThanDays    = flag.String("notify-images-older-than-days", "", "Notify if image is older than X days (default: 30)")
	notifyVolumesOlderThanDays   = flag.String("notify-unattached-older-than-days", "", "Notify if volume is older than X days (default: 30)")
	notifySnapshotsOlderThanDays = flag.String("notify-snapshots-older-than-days", "", "Notify if snapshot is older than X days (default: 30)")
	notifyBucketsOlderThanDays   = flag.String("notify-buckets-older-than-days", "", "Notify if bucket is older than X days (default: 30)")
	notifyWhitelistOlderThanDays = flag.String("notify-whitelist-older-than-days", "", "Notify if whitelisted is older than X days (default: 182)")
	notifyDndOlderThanDays       = flag.String("notify-dnd-older-than-days", "", "Do not delete older than X days (default: 7)")
)

const banner = `
   ___ _                 _
  / __\ | ___  _   _  __| |_____      _____  ___ _ __   ___ _ __
 / /  | |/ _ \| | | |/ _` + "`" + ` / __\ \ /\ / / _ \/ _ \ '_ \ / _ \ '__|
/ /___| | (_) | |_| | (_| \__ \\ V  V /  __/  __/ |_) |  __/ |
\____/|_|\___/ \__,_|\__,_|___/ \_/\_/ \___|\___| .__/ \___|_|
                                                |_|
`

func main() {
	fmt.Println(banner)
	loadConfig()
	flag.Parse()
	loadThresholds()
	csp := cspFromConfig(findConfig("csp"))
	log.Printf("Running against %s...\n", csp)
	switch getPositionalCmd() {
	case "cleanup":
		log.Println("Entering cleanup mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		cleanup.PerformCleanup(mngr)
	case "reset":
		log.Println("Entering reset mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		cleanup.ResetCloudsweeper(mngr)
	case "mark-for-cleanup":
		log.Println("Entering 'mark-for-cleanup' mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		taggedResources := cleanup.MarkForCleanup(mngr, thresholds, *dryRun)
		if *dryRun {
			client := initNotifyClient()
			client.MarkingDryRunReport(taggedResources, org.AccountToUserMapping(csp))
		} else {
			log.Println("Not sending marking report since this was not a dry run")
		}
	case "review":
		log.Println("Entering 'review' mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		client := initNotifyClient()
		client.OldResourceReview(mngr, org, csp, thresholds)
	case "warn":
		log.Println("Entering 'warn' mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		client := initNotifyClient()
		client.DeletionWarning(findConfigInt("warning-hours"), mngr, org.AccountToUserMapping(csp))
	case "billing-report":
		log.Println("Entering 'billing-report' mode", csp)
		var reporter billing.Reporter
		if csp == cloud.AWS {
			billingAccount := findConfig("billing-account")
			bucket := findConfig("billing-bucket")
			region := findConfig("billing-bucket-region")
			sortTag := findConfig("billing-sort-tag")
			reporter = billing.NewReporterAWS(billingAccount, bucket, region, sortTag)
		} else if csp == cloud.GCP {
			bucket := findConfig("billing-bucket")
			prefix := findConfig("billing-csv-prefix")
			reporter = billing.NewReporterGCP(bucket, prefix)
		} else {
			log.Fatalf("Invalid CSP specified")
			return
		}
		report := billing.GenerateReport(reporter)
		org := parseOrganization(findConfig("org-file"))
		mapping := org.AccountToUserMapping(csp)
		sortTagKey := findConfig("billing-sort-tag")
		log.Println(report.FormatReport(mapping, sortTagKey != ""))
		client := initNotifyClient()
		client.MonthToDateReport(report, mapping, sortTagKey != "")
	case "find-untagged":
		log.Println("Entering 'find-untagged' mode")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		mapping := org.AccountToUserMapping(csp)
		client := initNotifyClient()
		client.UntaggedResourcesReview(mngr, mapping)
	case "find-resource":
		id := *findResourceID
		if id == "" {
			log.Fatalln("Must specify a resource ID to find using --resource-id=<ID>")
		}
		log.Printf("Entering 'find-resource' mode (Resource ID: %s)", id)
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		client, err := find.Init(mngr, org, csp)
		if err != nil {
			log.Fatalf("Could not initalize find client: %s", err)
		}
		err = client.FindResource(id)
		if err != nil {
			log.Fatal(err)
		}
	case "setup":
		log.Println("Running Cloudsweeper setup")
		setup.PerformSetup(findConfig("aws-master-arn"))
	default:
		log.Fatalln("Please supply a command")
	}
	log.Println("Finished running")
}

func initManager(csp cloud.CSP, org *cs.Organization) cloud.ResourceManager {
	manager, err := cloud.NewManager(csp, org.EnabledAccounts(csp)...)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return manager
}

func initNotifyClient() *notify.Client {
	config := &notify.Config{
		SMTPUsername:           findConfig("smtp-username"),
		SMTPPassword:           findConfig("smtp-password"),
		SMTPServer:             findConfig("smtp-server"),
		SMTPPort:               findConfigInt("smtp-port"),
		DisplayName:            findConfig("display-name"),
		MailFrom:               findConfig("mail-from"),
		EmailDomain:            findConfig("mail-domain"),
		BillingReportAddressee: findConfig("billing-report-addressee"),
		TotalSumAddresse:       findConfig("total-sum-addressee"),
	}
	return notify.Init(config)
}

func parseOrganization(inputFile string) *cs.Organization {
	raw, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Could not read organization file: %s\n", err)
	}
	org, err := cs.InitOrganization(raw)
	if err != nil {
		log.Fatalf("Failed to initalize organization: %s\n", err)
	}
	return org
}

func getPositionalCmd() string {
	n := len(os.Args)
	if n <= 1 {
		return ""
	}
	return os.Args[n-1]
}

func cspFromConfig(rawFlag string) cloud.CSP {
	flagVal := strings.ToLower(rawFlag)
	switch flagVal {
	case cspFlagAWS:
		return cloud.AWS
	case cspFlagGCP:
		return cloud.GCP
	default:
		fmt.Fprintf(os.Stderr, "Invalid CSP flag \"%s\" specified\n", rawFlag)
		os.Exit(1)
		return cloud.AWS
	}
}
