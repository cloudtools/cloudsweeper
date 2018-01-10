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
	mngr := res.NewManager(res.AwsCSP, asd...)
	t := mngr.ImagesPerAccount()
	for key, value := range t {
		fmt.Println(key)
		fmt.Println(len(value))
		fmt.Println(value[0].ID())
	}
}
