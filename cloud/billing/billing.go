// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package billing

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/agaridata/cloudsweeper/cloud"
)

const (
	dateFormatLayout = "2006-01-02"
	// MinimumTotalCost is also used in notify.MonthToDateReport
	MinimumTotalCost = 10.0
	// MinimumCost is also used in notify.MonthToDateReport
	MinimumCost = 5.0
)

// ReportItem represent a single item in a report. This is usually
// the cost for a specific service for a certain user in a certain
// account/project.
type ReportItem struct {
	Owner        string
	Description  string
	Cost         float64
	sortTagValue string
}

// User represents an User and it's TotalCost
// plus a CostList of all associated DetailedCosts
type User struct {
	Name          string
	TotalCost     float64
	DetailedCosts CostList
}

// UserList respresents a list of Users
type UserList []User

func (l UserList) Len() int           { return len(l) }
func (l UserList) Less(i, j int) bool { return l[i].TotalCost < l[j].TotalCost }
func (l UserList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// DetailedCost represents a Cost and Description for a Users expense
type DetailedCost struct {
	Cost        float64
	Description string
}

// CostList respresents a list of Costs
type CostList []DetailedCost

func (l CostList) Len() int           { return len(l) }
func (l CostList) Less(i, j int) bool { return l[i].Cost < l[j].Cost }
func (l CostList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

// Reporter is a general interface that can be implemented
// for both AWS and GCP to generate expense reports.
type Reporter interface {
	GenerateReport(start time.Time) Report
}

// NewReporterAWS will initialize a new Reporter for the AWS cloud. This
// requires specifying the account which holds the billing information,
// the bucket where the billing CSVs can be found as well as which region
// this bucket is in. None of these arguments must be empty.
func NewReporterAWS(billingAccount, bucket, bucketRegion, sortTag string) Reporter {
	if billingAccount == "" || bucket == "" || bucketRegion == "" {
		panic("Invalid arguments, must not be empty (\"\")")
	}
	return &awsReporter{
		csp:                 cloud.AWS,
		billingAccount:      billingAccount,
		billingBucket:       bucket,
		billingBucketRegion: bucketRegion,
		sortByTag:           sortTag,
	}
}

// NewReporterGCP initializes and returns a new Reporter for the GCP cloud.
// This requires specifying a bucket where the billing CSVs can be found, as
// well as the prefix of these CSV files. The prefix will be prepended to
// the date and .csv suffix (e.g. <YOUR PREFIX>-2018-10-09.csv). None of
// these argument must be empty.
func NewReporterGCP(bucket, csvPrefix string) Reporter {
	if bucket == "" || csvPrefix == "" {
		panic("Invalid argument, must not be empty")
	}
	return &gcpReporter{
		csp:           cloud.GCP,
		bucket:        bucket,
		csvNamePrefix: csvPrefix,
	}
}

// Report contains a collection of items, and some metadata
// about when the items were collected and which dates they
// span. The report struct also has methods to help work with
// all the items.
type Report struct {
	CSP   cloud.CSP
	Items []ReportItem
}

// TotalCost returns the total cost for all items
func (r *Report) TotalCost() float64 {
	total := 0.0
	for i := range r.Items {
		total += r.Items[i].Cost
	}
	return total
}

// SortedUsersByTotalCost returns a sorted list of Users by TotalCost
func (r *Report) SortedUsersByTotalCost() UserList {
	type tempUser struct {
		name          string
		totalCost     float64
		detailedCosts map[string]float64
	}
	userMap := make(map[string]*tempUser)
	// Go through all ReportItems
	for _, item := range r.Items {
		// Group by AccountId
		if user, ok := userMap[item.Owner]; ok {
			user.totalCost += item.Cost
			// Group by Description
			if cost, ok := user.detailedCosts[item.Description]; ok {
				user.detailedCosts[item.Description] = cost + item.Cost
			} else {
				user.detailedCosts[item.Description] = item.Cost
			}
		} else {
			costs := make(map[string]float64)
			costs[item.Description] = item.Cost
			userMap[item.Owner] = &tempUser{item.Owner, item.Cost, costs}
		}
	}

	userList := make(UserList, 0, len(userMap))
	for _, user := range userMap {
		// omit users with low TotalCost
		if user.totalCost < MinimumTotalCost {
			continue
		}
		// convert detailedCosts into sorted CostLists
		detailedCostList := convertCostMapToSortedList(user.detailedCosts)
		// add generated User to userList
		userList = append(userList, User{user.name, user.totalCost, detailedCostList})
	}

	sort.Sort(sort.Reverse(userList))
	return userList
}

// SortedTagsByTotalCost returns a sorted list of grouped sort tag values,
// sorted by their total cost.
func (r *Report) SortedTagsByTotalCost() UserList {
	type tempTag struct {
		name          string
		totalCost     float64
		detailedCosts map[string]float64
	}
	tagMap := make(map[string]*tempTag)
	// Iterate through all report items
	for _, item := range r.Items {
		// Group by sort tag value
		if tag, ok := tagMap[item.sortTagValue]; ok {
			tag.totalCost += item.Cost
			// Group by Description
			if cost, ok := tag.detailedCosts[item.Description]; ok {
				tag.detailedCosts[item.Description] = cost + item.Cost
			} else {
				tag.detailedCosts[item.Description] = item.Cost
			}
		} else {
			costs := make(map[string]float64)
			costs[item.Description] = item.Cost
			tagMap[item.sortTagValue] = &tempTag{item.sortTagValue, item.Cost, costs}
		}
	}

	tagList := make(UserList, 0, len(tagMap))
	for _, tag := range tagMap {
		// Omit tags with low total cost
		if tag.totalCost < MinimumTotalCost {
			continue
		}

		// Convert detailed costs into sorted cost lists
		detailedCostList := convertCostMapToSortedList(tag.detailedCosts)
		// Add generated tag to tag list
		tagList = append(tagList, User{tag.name, tag.totalCost, detailedCostList})
	}

	sort.Sort(sort.Reverse(tagList))
	return tagList
}

// FormatReport returns a simple version of the Month-to-date billing report. It
// takes a mapping form account/project ID to employee username in order to
// more easily distinguish the owner of a cost.
func (r *Report) FormatReport(accountToUserMapping map[string]string, sortedByTags bool) string {
	b := new(bytes.Buffer)
	var sorted UserList
	if sortedByTags {
		sorted = r.SortedTagsByTotalCost()
	} else {
		sorted = r.SortedUsersByTotalCost()
	}

	fmt.Fprintln(b, "\n\nSummary:")
	fmt.Fprintln(b, "Name      | Cost ($)")
	fmt.Fprintln(b, "----------------------------")
	for _, user := range sorted {
		name := user.Name
		if realName, exist := accountToUserMapping[name]; exist {
			name = realName
		} else {
			// Assume this is a support cost
			if name == "" {
				if sortedByTags {
					name = "<not tagged>"
				} else {
					name = "Support"
				}
			}
		}
		fmt.Fprintf(b, "%-12s | %8.2f\n", name, user.TotalCost)
	}

	fmt.Fprintf(b, "\nDetails:")
	for _, user := range sorted {
		name := user.Name
		if realName, exist := accountToUserMapping[name]; exist {
			name = realName
		} else {
			// Assume this is a support cost
			if name == "" {
				if sortedByTags {
					name = "<not tagged>"
				} else {
					name = "support"
				}
			}
		}
		fmt.Fprintf(b, "\n%s's costs:\n", name)
		fmt.Fprintln(b, "Cost ($) | Description")
		fmt.Fprintln(b, "---------------------------")
		for _, cost := range user.DetailedCosts {
			fmt.Fprintf(b, "%-8.2f | %s\n", cost.Cost, cost.Description)
		}
	}
	return b.String()
}

// GenerateReport generates a Month-to-date billing report for the current month
func GenerateReport(reporter Reporter) Report {
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	return reporter.GenerateReport(start)
}

func convertCostMapToSortedList(costMap map[string]float64) CostList {
	costList := make(CostList, 0, len(costMap))
	for desc, cost := range costMap {
		if cost > MinimumCost {
			costList = append(costList, DetailedCost{Description: desc, Cost: cost})
		}
	}
	sort.Sort(sort.Reverse(costList))
	return costList
}

func updateCsvHeaders(record []string) map[string]int {
	csvHeaders := make(map[string]int)
	for i, column := range record {
		csvHeaders[column] = i
	}
	return csvHeaders
}
