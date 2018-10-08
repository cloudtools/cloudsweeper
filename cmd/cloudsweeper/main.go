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
	"github.com/cloudtools/cloudsweeper/cloudsweeper/notify"
	"github.com/cloudtools/cloudsweeper/cloudsweeper/setup"
)

const (
	defaultOrgFile = "organization.json"

	sharedQAAccount     = "475063612724"
	sharedDevAWSAccount = "164337164081"
	prodAWSAccount      = "992270393355"

	warningHoursInAdvance = 48
)

var (
	orgFile      = flag.String("org-file", defaultOrgFile, "Specify where to find the JSON with organization information")
	warningHours = flag.Int("warning-hours", warningHoursInAdvance, "The number of hours in advance to warn about resource deletion")
	cspToUse     = flag.String("csp", defaultCSPFlag, "Which CSP to run against")
)

const banner = `
   ___ _                 _                                       
  / __\ | ___  _   _  __| |_____      _____  ___ _ __   ___ _ __ 
 / /  | |/ _ \| | | |/ _` + "`" + ` / __\ \ /\ / / _ \/ _ \ '_ \ / _ \ '__|
/ /___| | (_) | |_| | (_| \__ \\ V  V /  __/  __/ |_) |  __/ |   
\____/|_|\___/ \__,_|\__,_|___/ \_/\_/ \___|\___| .__/ \___|_|   
                                                |_|
`

const (
	cmdCleanup  = "cleanup"
	cmdReset    = "reset"
	cmdMark     = "mark-for-cleanup"
	cmdReview   = "review"
	cmdSetup    = "setup"
	cmdWarn     = "warn"
	cmdBilling  = "billing-report"
	cmdUntagged = "find-untagged"

	defaultCSPFlag = cspFlagAWS
	cspFlagAWS     = "aws"
	cspFlagGCP     = "gcp"
)

func main() {
	fmt.Println(banner)
	flag.Parse()
	csp := cspFromFlag(*cspToUse)
	fmt.Printf("Running against %s...\n", csp)
	switch getPositional() {
	case cmdCleanup:
		log.Println("Cleaning up old resources")
		org := parseOrganization(*orgFile)
		mngr := initManager(csp, org)
		cleanup.PerformCleanup(mngr)
	case cmdReset:
		log.Println("Resetting all tags")
		org := parseOrganization(*orgFile)
		mngr := initManager(csp, org)
		cleanup.ResetCloudsweeper(mngr)
	case cmdMark:
		log.Println("Marking old resources for cleanup")
		org := parseOrganization(*orgFile)
		mngr := initManager(csp, org)
		cleanup.MarkForCleanup(mngr)
	case cmdReview:
		log.Println("Sending out old resource review")
		org := parseOrganization(*orgFile)
		mngr := initManager(csp, org)
		notify.OldResourceReview(mngr, org, csp)
	case cmdWarn:
		log.Println("Sending out cleanup warning")
		org := parseOrganization(*orgFile)
		mngr := initManager(csp, org)
		notify.DeletionWarning(*warningHours, mngr, org.AccountToUserMapping(csp))
	case cmdBilling:
		log.Println("Generating month-to-date billing report for", csp)
		reporter, err := billing.NewReporter(csp)
		if err != nil {
			log.Fatal(err)
			return
		}
		report := billing.GenerateReport(reporter)
		org := parseOrganization(*orgFile)
		mapping := org.AccountToUserMapping(csp)
		log.Println(report.FormatReport(mapping))
		notify.MonthToDateReport(report, mapping)
	case cmdUntagged:
		log.Println("Finding untagged resources")
		// Only care about prod, shared-dev and QA
		mngr, err := cloud.NewManager(csp, sharedDevAWSAccount, prodAWSAccount, sharedQAAccount)
		if err != nil {
			log.Fatal(err)
		}
		mapping := map[string]string{sharedDevAWSAccount: "cloud-dev", prodAWSAccount: "prod", sharedQAAccount: "qa"}
		notify.UntaggedResourcesReview(mngr, mapping)
	case cmdSetup:
		log.Println("Running cloudsweeper setup")
		setup.PerformSetup()
	default:
		// Default to setup
		log.Println("Running cloudsweeper setup")
		setup.PerformSetup()
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

func getPositional() string {
	n := len(os.Args)
	if n <= 1 {
		return ""
	}
	return os.Args[n-1]
}

func cspFromFlag(rawFlag string) cloud.CSP {
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
