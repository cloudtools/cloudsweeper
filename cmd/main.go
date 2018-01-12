package main

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/filter"
	"fmt"
)

const (
	sharedQAaccount = "475063612724"
)

func main() {
	asd := []string{sharedQAaccount}
	mngr := cloud.NewManager(cloud.AWS, asd...)
	instances := mngr.InstancesPerAccount()
	for _, val := range instances {
		fil := filter.New()
		fil.AddGeneralRule(filter.NameContains("NatGateway"))
		fil.AddGeneralRule(filter.OlderThanXYears(1))
		newInstances := fil.FilterInstances(val)
		for i := range newInstances {
			fmt.Println(newInstances[i].Tags()["Name"])
		}
	}

	snapshots := mngr.SnapshotsPerAccount()
	for _, val := range snapshots {
		fil := filter.New()
		fil.AddGeneralRule(filter.OlderThanXDays(60))

		snaps := fil.FilterSnapshots(val)
		for i := range snaps {
			fmt.Println(snaps[i].Tags()["Name"])
		}
	}
}
