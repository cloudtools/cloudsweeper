package billing

import (
	"brkt/olga/cloud"
	"bytes"
	"errors"
	"fmt"
	"sort"
	"time"
)

const (
	dateFormatLayout = "2006-01-02"
	// MinimumTotalCost is also used in notify.MonthToDateReport
	MinimumTotalCost = 100.0
	// MinimumCost is also used in notify.MonthToDateReport
	MinimumCost = 5.0
)

// ReportItem represent a single item in a report. This is usually
// the cost for a specific service for a certain user in a certain
// account/project.
type ReportItem struct {
	Owner       string
	Description string
	Cost        float64
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
	GenerateReport(startDate, endDate time.Time) Report
}

// NewReporter intializes a new billing reporter for the specified CSP
func NewReporter(csp cloud.CSP) (Reporter, error) {
	switch csp {
	case cloud.AWS:
		return &awsReporter{
			csp: cloud.AWS,
		}, nil
	case cloud.GCP:
		return nil, errors.New("GCP is not supported yet")
	default:
		return nil, errors.New("Invalid CSP specified")
	}
}

// Report contains a collection of items, and some metadata
// about when the items were collected and which dates they
// span. The report struct also has methods to help work with
// all the items.
type Report struct {
	CSP          cloud.CSP
	StartDate    time.Time
	EndDate      time.Time
	CreationDate time.Time
	Items        []ReportItem
}

// FormatStartDate will return the StartDate formatted into
// YYYY-MM-DD, e.g. 2017-01-16
func (r *Report) FormatStartDate() string {
	return r.StartDate.Format(dateFormatLayout)
}

// FormatEndDate will return the EndDate formatted into
// YYYY-MM-DD, e.g. 2017-01-16
func (r *Report) FormatEndDate() string {
	return r.StartDate.Format(dateFormatLayout)
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

// FormatReport returns a simple version of the Month-to-date billing report. It
// takes a mapping form account/project ID to employee username in order to
// more easily distinguish the owner of a cost.
func (r *Report) FormatReport(accountToUserMapping map[string]string) string {
	b := new(bytes.Buffer)
	sortedUsersByTotalCost := r.SortedUsersByTotalCost()

	fmt.Fprintln(b, "\n\nSummary:")
	fmt.Fprintln(b, "Account      | Cost ($)")
	fmt.Fprintln(b, "----------------------------")
	for _, user := range sortedUsersByTotalCost {
		name := user.Name
		if realName, exist := accountToUserMapping[name]; exist {
			name = realName
		}
		fmt.Fprintf(b, "%-12s | %8.2f\n", name, user.TotalCost)
	}

	fmt.Fprintf(b, "\nDetails:")
	for _, user := range sortedUsersByTotalCost {
		name := user.Name
		if realName, exist := accountToUserMapping[name]; exist {
			name = realName
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
	startDate := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	return reporter.GenerateReport(startDate, today)
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

// DaysBetween return all days between two given dates (inclusive)
func DaysBetween(startTime, endTime time.Time) []time.Time {
	//  date.Year() != endTime.Year() && date.Month() != endTime.Month() && date.Day() != endTime.Day()
	sameDates := func(t1, t2 time.Time) bool {
		y1, m1, d1 := t1.Date()
		y2, m2, d2 := t2.Date()
		return y1 == y2 && m1 == m2 && d1 == d2
	}

	result := []time.Time{}
	for date := startTime; !sameDates(date, endTime); date = date.AddDate(0, 0, 1) {
		result = append(result, date)
	}
	// Add the last date too so that list is inclusive
	result = append(result, endTime)
	return result
}

// MonthsBetween return all months between two given dates (inclusive)
func MonthsBetween(startTime, endTime time.Time) []time.Time {
	result := []time.Time{}
	for date := startTime; date.Year() != endTime.Year() && date.Month() != endTime.Month(); date = date.AddDate(0, 1, 0) {
		result = append(result, date)
	}
	// Add the last date too so that list is inclusive
	result = append(result, endTime)
	return result
}
