package main

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/billing"
	"brkt/housekeeper/cloud/filter"
	hk "brkt/housekeeper/housekeeper"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

const (
	defaultAccountsFile = "aws_accounts.json"

	sharedQAAccount = "475063612724"
)

var (
	accountsFile = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
)

func main() {
	rep := billing.NewReporter(cloud.AWS)
	start, _ := time.Parse("2006-1-2", "2017-12-01")
	then, _ := time.Parse("2006-1-2", "2017-12-31")
	report := rep.GenerateReport(start, then)
	flag.Parse()
	owners := parseAWSAccounts(*accountsFile)
	asd := owners.IDToName()
	for key, val := range report.TotalPerOwner() {
		fmt.Printf("%s:%s:\t\t$%.3f\n", key, asd[key], val)
	}
	fmt.Println("Total:", report.TotalCost())
	return

	mngr := cloud.NewManager(cloud.AWS, []string{sharedQAAccount}...)
	instances := mngr.InstancesPerAccount()
	for _, val := range instances {
		fil := filter.New()
		fil.AddGeneralRule(filter.NameContains("alexander"))
		//fil.AddGeneralRule(filter.OlderThanXYears(1))
		newInstances := fil.FilterInstances(val)
		err := mngr.CleanupInstances(newInstances)
		if err != nil {
			panic(err)
		}
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
