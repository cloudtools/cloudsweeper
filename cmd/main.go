package main

import (
	"brkt/housekeeper/res"
	"fmt"
)

const (
	sharedQAaccount = "475063612724"
)

func main() {
	asd := []string{sharedQAaccount}
	mngr := res.NewManager(res.AWS, asd...)
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
