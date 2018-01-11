package main

import (
	"brkt/housekeeper/cloud"
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
		fmt.Println(val[0].ID())
	}
	resources := mngr.AllResourcesPerAccount()
	for account, resource := range resources {
		fmt.Println("Account:", account)
		fmt.Println("Instances:", len(resource.Instances))
		fmt.Println("Images:", len(resource.Images))
	}
}
