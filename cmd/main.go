package main

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/filter"
	hk "brkt/housekeeper/housekeeper"
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
	accountsFile = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
)

func main() {
	flag.Parse()
	//owners := parseAWSAccounts(*accountsFile)

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
