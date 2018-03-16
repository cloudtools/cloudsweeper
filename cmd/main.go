package main

import (
	"brkt/olga/cloud"
	"brkt/olga/cloud/billing"
	"brkt/olga/housekeeper"
	hk "brkt/olga/housekeeper"
	"brkt/olga/housekeeper/cleanup"
	"brkt/olga/housekeeper/notify"
	"brkt/olga/housekeeper/setup"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	defaultAccountsFile = "aws_accounts.json"

	sharedQAAccount     = "475063612724"
	sharedDevAWSAccount = "164337164081"
	prodAWSAccount      = "992270393355"
	secProdAWSAccount   = "108660276130"
	secDevAWSAccount    = "120690514258"
	secStageAWSAccount  = "605040402381"
	soloProdAWSAccount  = "067829456282"
	soloStageAWSAccount = "842789976943"
	soloFriendliesAgari = "139798613772"
	soloFriendliesMark  = "586683603820"

	warningHoursInAdvance = 48
)

var (
	accountsFile = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
	warningHours = flag.Int("warning-hours", warningHoursInAdvance, "The number of hours in advance to warn about resource deletion")
	cspToUse     = flag.String("csp", defaultCSPFlag, "Which CSP to run against")
)

const banner = `
  /\  /\___  _   _ ___  ___  /\ /\___  ___ _ __   ___ _ __
 / /_/ / _ \| | | / __|/ _ \/ //_/ _ \/ _ \ '_ \ / _ \ '__|
/ __  / (_) | |_| \__ \  __/ __ \  __/  __/ |_) |  __/ |
\/ /_/ \___/ \__,_|___/\___\/  \/\___|\___| .__/ \___|_|
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
	org := parseOrganization("organization.json")
	for _, emp := range org.Employees {
		if emp.Username == "hsson" || emp.Username == "olle" {
			fmt.Printf(`Name: %s
Manager: %s
Department: %s
===========
`, emp.RealName, emp.Manager.RealName, emp.Department.Name)
		}
	}

	return
	fmt.Println(banner)
	flag.Parse()
	csp := cspFromFlag(*cspToUse)
	fmt.Printf("Running against %s...\n", csp)
	switch getPositional() {
	case cmdCleanup:
		log.Println("Cleaning up old resources")
		cleanup.PerformCleanup(csp, parseAWSAccounts(*accountsFile))
	case cmdReset:
		log.Println("Resetting all tags")
		cleanup.ResetHousekeeper(csp, parseAWSAccounts(*accountsFile))
	case cmdMark:
		log.Println("Marking old resources for cleanup")
		cleanup.MarkForCleanup(csp, parseAWSAccounts(*accountsFile))
	case cmdReview:
		log.Println("Sending out old resource review")
		notify.OldResourceReview(csp, parseAWSAccounts(*accountsFile))
	case cmdWarn:
		log.Println("Sending out cleanup warning")
		notify.DeletionWarning(*warningHours, csp, parseAWSAccounts(*accountsFile))
	case cmdBilling:
		log.Println("Generating month-to-date billing report")
		owners := parseAWSAccounts(*accountsFile)
		report := billing.GenerateReport(csp, owners)
		log.Println(report.FormatReport(owners))
		notify.MonthToDateReport(report, owners)
	case cmdUntagged:
		log.Println("Finding untagged resources")
		// Only care about prod, shared-dev and QA
		owners := housekeeper.Owners{
			housekeeper.Owner{Name: "cloud-dev", ID: sharedDevAWSAccount},
			housekeeper.Owner{Name: "prod", ID: prodAWSAccount},
			housekeeper.Owner{Name: "qa", ID: sharedQAAccount},
		}
		notify.UntaggedResourcesReview(csp, owners)
	case cmdSetup:
		log.Println("Running housekeeper setup")
		setup.PerformSetup()
	default:
		// Default to setup
		log.Println("Running housekeeper setup")
		setup.PerformSetup()
	}
}

func parseAWSAccounts(inputFile string) hk.Owners {
	raw, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalln("Could not read accounts file:", err)
	}
	owners := hk.Owners{}
	err = json.Unmarshal(raw, &owners)
	if err != nil {
		log.Fatalln("Could not parse JSON:", err)
	}
	if *accountsFile == defaultAccountsFile {
		// Add any additional accounts that are not present in the accounts file
		owners = append(owners, hk.Owner{Name: "cloud-dev", ID: sharedDevAWSAccount})
		owners = append(owners, hk.Owner{Name: "prod", ID: prodAWSAccount})
		owners = append(owners, hk.Owner{Name: "sec-prod", ID: secProdAWSAccount})
		owners = append(owners, hk.Owner{Name: "sec-dev", ID: secDevAWSAccount})
		owners = append(owners, hk.Owner{Name: "sec-stage", ID: secStageAWSAccount})
		owners = append(owners, hk.Owner{Name: "solo-prod", ID: soloProdAWSAccount})
		owners = append(owners, hk.Owner{Name: "solo-stage", ID: soloStageAWSAccount})
		owners = append(owners, hk.Owner{Name: "solo-friendlies-agari", ID: soloFriendliesAgari})
		owners = append(owners, hk.Owner{Name: "solo-friendlies-mark", ID: soloFriendliesMark})
	}
	return owners
}

func parseOrganization(inputFile string) *hk.Organization {
	raw, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Could not read organization file: %s\n", err)
	}
	org, err := hk.InitOrganization(raw)
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
