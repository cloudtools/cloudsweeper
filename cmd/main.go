package main

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/filter"
	hk "brkt/housekeeper/housekeeper"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	defaultAccountsFile = "aws_accounts.json"
)

var (
	accountsFile = flag.String("accounts-file", defaultAccountsFile, "Specify where to find the JSON with all accounts")
)

func main() {
	flag.Parse()
	owners := parseAWSAccounts(*accountsFile)

	mngr := cloud.NewManager(cloud.AWS, owners.AllIDs()...)
	instances := mngr.InstancesPerAccount()
	for _, val := range instances {
		fil := filter.New()
		//fil.AddGeneralRule(filter.NameContains("NatGateway"))
		fil.AddGeneralRule(filter.OlderThanXYears(1))
		newInstances := fil.FilterInstances(val)
		for i := range newInstances {
			fmt.Println(newInstances[i].Tags()["Name"])
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
