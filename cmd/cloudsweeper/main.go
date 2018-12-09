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

	"github.com/cloudtools/cloudsweeper/cloud"
	"github.com/cloudtools/cloudsweeper/cloud/billing"
	cs "github.com/cloudtools/cloudsweeper/cloudsweeper"
	"github.com/cloudtools/cloudsweeper/cloudsweeper/cleanup"
	"github.com/cloudtools/cloudsweeper/cloudsweeper/find"
	"github.com/cloudtools/cloudsweeper/cloudsweeper/notify"
	"github.com/cloudtools/cloudsweeper/cloudsweeper/setup"
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

	warningHours    = flag.String("warning-hours", "", "The number of hours in advance to warn about resource deletion")
	displayName     = flag.String("display-name", "", "Name displayed on emails sent by Cloudsweeper")
	summaryReciever = flag.String("summary-addressee", "", "Reciever of month to date summaries")
	summaryManager  = flag.String("total-sum-addressee", "", "Reciever of total cost sums")
	mailDomain      = flag.String("mail-domain", "", "The mail domain appended to usernames specified in the organization")

	setupARN = flag.String("aws-master-arn", "", "AWS ARN of role in account used by Cloudsweeper to assume roles")

	findResourceID = flag.String("resource-id", "", "ID of resource to find with find-resource command")
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
	csp := cspFromConfig(findConfig("csp"))
	log.Printf("Running against %s...\n", csp)
	switch getPositionalCmd() {
	case "cleanup":
		log.Println("Cleaning up old resources")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		cleanup.PerformCleanup(mngr)
	case "reset":
		log.Println("Resetting all tags")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		cleanup.ResetCloudsweeper(mngr)
	case "mark-for-cleanup":
		log.Println("Marking old resources for cleanup")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		cleanup.MarkForCleanup(mngr)
	case "review":
		log.Println("Sending out old resource review")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		client := initNotifyClient()
		client.OldResourceReview(mngr, org, csp)
	case "warn":
		log.Println("Sending out cleanup warning")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		client := initNotifyClient()
		client.DeletionWarning(findConfigInt("warning-hours"), mngr, org.AccountToUserMapping(csp))
	case "billing-report":
		log.Println("Generating month-to-date billing report for", csp)
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
		log.Println("Finding untagged resources")
		org := parseOrganization(findConfig("org-file"))
		mngr := initManager(csp, org)
		mapping := org.AccountToUserMapping(csp)
		client := initNotifyClient()
		client.UntaggedResourcesReview(mngr, mapping)
	case "find-resource":
		id := *findResourceID
		if id == "" {
			log.Fatalln("Must specify a resource ID to find, using --resource-id=<ID>")
		}
		log.Printf("Finding resource with ID %s", id)
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
		log.Println("Running cloudsweeper setup")
		setup.PerformSetup(findConfig("aws-master-arn"))
	default:
		log.Fatalln("Please supply a command")
	}
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
		SMTPUsername:     findConfig("smtp-username"),
		SMTPPassword:     findConfig("smtp-password"),
		SMTPServer:       findConfig("smtp-server"),
		SMTPPort:         findConfigInt("smtp-port"),
		DisplayName:      findConfig("display-name"),
		EmailDomain:      findConfig("mail-domain"),
		SummaryAddressee: findConfig("summary-addressee"),
		TotalSumAddresse: findConfig("total-sum-addressee"),
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
