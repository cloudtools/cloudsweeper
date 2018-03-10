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
	"os"
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
)

const banner = `
  /\  /\___  _   _ ___  ___  /\ /\___  ___ _ __   ___ _ __
 / /_/ / _ \| | | / __|/ _ \/ //_/ _ \/ _ \ '_ \ / _ \ '__|
/ __  / (_) | |_| \__ \  __/ __ \  __/  __/ |_) |  __/ |
\/ /_/ \___/ \__,_|___/\___\/  \/\___|\___| .__/ \___|_|
                                          |_|
										`

const (
	cmdCleanup = "cleanup"
	cmdReset   = "reset"
	cmdMark    = "mark-for-cleanup"
	cmdReview  = "review"
	cmdSetup   = "setup"
	cmdWarn    = "warn"
	cmdBilling = "billing-report"
)

func main() {
	fmt.Println(banner)
	flag.Parse()
	switch getPositional() {
	case cmdCleanup:
		log.Println("Cleaning up old resources")
		cleanup.PerformCleanup(cloud.AWS, parseAWSAccounts(*accountsFile))
	case cmdReset:
		log.Println("Resetting all tags")
		cleanup.ResetHousekeeper(cloud.AWS, parseAWSAccounts(*accountsFile))
	case cmdMark:
		log.Println("Marking old resources for cleanup")
		cleanup.MarkForCleanup(cloud.AWS, parseAWSAccounts(*accountsFile))
	case cmdReview:
		log.Println("Sending out old resource review")
		notify.OldResourceReview(cloud.AWS, parseAWSAccounts(*accountsFile))
	case cmdWarn:
		log.Println("Sending out cleanup warning")
		notify.DeletionWarning(*warningHours, cloud.AWS, parseAWSAccounts(*accountsFile))
	case cmdBilling:
		log.Println("Generating month-to-date billing report")
		owners := parseAWSAccounts(*accountsFile)
		report := billing.GenerateReport(cloud.AWS, owners)
		log.Println(report.FormatReport(owners))
		notify.MonthToDateReport(report, owners)
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

func getPositional() string {
	n := len(os.Args)
	if n <= 1 {
		return ""
	}
	return os.Args[n-1]
}
