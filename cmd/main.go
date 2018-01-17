package main

import (
	"brkt/housekeeper/cloud"
	hk "brkt/housekeeper/housekeeper"
	"brkt/housekeeper/housekeeper/cleanup"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

const (
	defaultAccountsFile = "aws_accounts.json"

	sharedQAAccount = "475063612724"
)

var (
	accountsFile   = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
	performCleanup = flag.Bool("cleanup", false, "Specify if cleanup should be performed")
)

func main() {
	flag.Parse()
	owners := parseAWSAccounts(*accountsFile)
	if *performCleanup {
		log.Println("Running cleanup")
		cleanup.PerformCleanup(cloud.AWS, owners)
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
