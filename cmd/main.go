package main

import (
	"brkt/housekeeper/cloud"
	hk "brkt/housekeeper/housekeeper"
	"brkt/housekeeper/housekeeper/cleanup"
	"brkt/housekeeper/housekeeper/notify"
	"brkt/housekeeper/housekeeper/setup"
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

	warningHoursInAdvance = 48
)

var (
	accountsFile   = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
	performCleanup = flag.Bool("cleanup", false, "Specify if cleanup should be performed")
	performMarking = flag.Bool("mark-for-cleanup", false, "Specify if resources should be marked")
	performReview  = flag.Bool("review", false, "Specify if review of old resources should be sent out")
	performSetup   = flag.Bool("setup", false, "Setup AWS account to allow housekeeping")
	performWarning = flag.Bool("warning", false, "Send out warning about resource cleanup")
	warningHours   = flag.Int("warning-hours", warningHoursInAdvance, "The number of hours in advance to warn about resource deletion")

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

	if *performReview {
		log.Println("Reviewing old resources")
		if owners == nil {
			owners = parseAWSAccounts(*accountsFile)
		}
		notify.OldResourceReview(cloud.AWS, owners)
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
	// The shared dev account is not in the imported accounts file
	owners = append(owners, hk.Owner{Name: "cloud-dev", ID: sharedDevAWSAccount})
	return owners
}
