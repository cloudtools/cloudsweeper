package main

import (
	"brkt/housekeeper/cloud"
	hk "brkt/housekeeper/housekeeper"
	"brkt/housekeeper/housekeeper/cleanup"
	"brkt/housekeeper/housekeeper/notify"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

const (
	defaultAccountsFile = "aws_accounts.json"

	sharedQAAccount     = "475063612724"
	sharedDevAWSAccount = "164337164081"
)

var (
	accountsFile   = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
	performCleanup = flag.Bool("cleanup", false, "Specify if cleanup should be performed")
	performNotify  = flag.Bool("notify", false, "Specify if notifications should be sent out")
)

func main() {
	flag.Parse()
	owners := parseAWSAccounts(*accountsFile)
	// The shared dev account is not in the imported accounts file
	owners = append(owners, hk.Owner{Name: "cloud-dev", ID: sharedDevAWSAccount})
	if *performCleanup {
		log.Println("Running cleanup")
		cleanup.PerformCleanup(cloud.AWS, owners)
	}

	if *performNotify {
		log.Println("Notifying")
		notify.OlderThanXMonths(3, cloud.AWS, []hk.Owner{hk.Owner{Name: "qa", ID: sharedQAAccount}})
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
	return owners
}
