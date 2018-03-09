package main

import (
	"brkt/olga/cloud"
	"brkt/olga/cloud/billing"
	hk "brkt/olga/housekeeper"
	"brkt/olga/housekeeper/cleanup"
	"brkt/olga/housekeeper/notify"
	"brkt/olga/housekeeper/setup"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	accountsFile   = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
	performCleanup = flag.Bool("cleanup", false, "Specify if cleanup should be performed")
	performReset   = flag.Bool("reset", false, "Remove deletion tag from resources")
	performMarking = flag.Bool("mark-for-cleanup", false, "Specify if resources should be marked")
	performReview  = flag.Bool("review", false, "Specify if review of old resources should be sent out")
	performSetup   = flag.Bool("setup", false, "Setup AWS account to allow housekeeping")
	performWarning = flag.Bool("warning", false, "Send out warning about resource cleanup")
	warningHours   = flag.Int("warning-hours", warningHoursInAdvance, "The number of hours in advance to warn about resource deletion")
	performReport  = flag.Bool("billing-report", false, "Generate a Month-to-date billing report")

	didAction = false
)

const banner = `
  /\  /\___  _   _ ___  ___  /\ /\___  ___ _ __   ___ _ __
 / /_/ / _ \| | | / __|/ _ \/ //_/ _ \/ _ \ '_ \ / _ \ '__|
/ __  / (_) | |_| \__ \  __/ __ \  __/  __/ |_) |  __/ |
\/ /_/ \___/ \__,_|___/\___\/  \/\___|\___| .__/ \___|_|
                                          |_|
										`

func main() {
	fmt.Println(banner)
	flag.Parse()
	var owners hk.Owners

	if *performSetup {
		setup.PerformSetup()
		return
	}

	if *performMarking {
		log.Println("Marking old resources for cleanup")
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		cleanup.MarkForCleanup(cloud.AWS, owners)
		didAction = true
	}

	if *performWarning {
		log.Println("Warning about cleanup")
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		notify.DeletionWarning(*warningHours, cloud.AWS, owners)
		didAction = true
	}

	if *performCleanup {
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		log.Println("Running cleanup")
		cleanup.PerformCleanup(cloud.AWS, owners)
		didAction = true
	}

	if *performReset {
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		log.Println("Resetting tags")
		cleanup.ResetHousekeeper(cloud.AWS, owners)
		didAction = true
	}

	if *performReview {
		log.Println("Reviewing old resources")
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		notify.OldResourceReview(cloud.AWS, owners)
		didAction = true
	}

	if *performReport {
		log.Println("Generating Month-to-date billing report")
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		report := billing.GenerateReport(cloud.AWS, owners)
		log.Println(report.FormatReport(owners))
		notify.MonthToDateReport(report, owners)
		didAction = true
	}

	// Perform default action (no flags specified)
	if !didAction {
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
