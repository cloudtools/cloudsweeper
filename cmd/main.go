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
	performNotify  = flag.Bool("notify", false, "Specify if notifications should be sent out")
	performBuckets = flag.Bool("buckets", false, "Include buckets, this can take some time")
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
	//notify.DeletionWarning(4, cloud.AWS, []hk.Owner{hk.Owner{Name: "qa", ID: sharedQAAccount}})
	//return
	/*
		reporter := billing.NewReporter(cloud.AWS)
		t1, _ := time.Parse("2006-01-02", "2017-12-01")
		t2, _ := time.Parse("2006-01-02", "2017-12-31")
		rep := reporter.GenerateReport(t1, t2)
		for own, val := range rep.TotalPerOwner() {
			fmt.Printf("%s\t$%.3f\n", own, val)
		}
		fmt.Println(rep.TotalCost())
		return
	*/
	/*
		mngr := cloud.NewManager(cloud.AWS, sharedQAAccount)
		for acc, bucks := range mngr.BucketsPerAccount() {
			fmt.Println(acc)
			superTotal := 0.0
			for _, buck := range bucks {
				fmt.Printf("name: %s\nsize: %.5fGB\ncount: %d\nlast: %s\ncreated: %s\n", buck.ID(), buck.TotalSizeGB(), buck.ObjectCount(), buck.LastModified().Format("2006-01-02"), buck.CreationTime().Format("2006-01-02"))
				months := time.Now().Sub(buck.CreationTime()).Hours() / 24 / 30
				total := billing.BucketPricePerMonth(buck) * months
				superTotal += total
				fmt.Printf("Accumulated cost: $%.3f\n\n", total)
			}
			fmt.Printf("Total accumulated cost: $%.5f\n", superTotal)

			fil := filter.New()
			fil.AddGeneralRule(filter.IDMatches("cf-templates-158hsj9iwexre-us-east-1"))
			for _, buck := range fil.FilterBuckets(bucks) {
				fmt.Println(buck.ID())
				err := buck.Cleanup()
				if err != nil {
					panic(err)
				}
			}
		}
		return
	*/

	flag.Parse()

	if *performSetup {
		setup.PerformSetup()
		return
	}

	if *performWarning {
		log.Println("Warning about cleanup")
		notify.DeletionWarning(*warningHours, cloud.AWS, []hk.Owner{hk.Owner{Name: "qa", ID: sharedQAAccount}})
		didAction = true
	}

	if *performCleanup {
		owners := parseAWSAccounts(*accountsFile)
		// The shared dev account is not in the imported accounts file
		owners = append(owners, hk.Owner{Name: "cloud-dev", ID: sharedDevAWSAccount})
		log.Println("Running cleanup")
		cleanup.PerformCleanup(cloud.AWS, owners)
		didAction = true
	}

	if *performNotify {
		log.Println("Notifying")
		notify.OlderThanXMonths(3, cloud.AWS, []hk.Owner{hk.Owner{Name: "qa", ID: sharedQAAccount}})
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
	return owners
}
