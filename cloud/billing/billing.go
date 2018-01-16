package billing

import (
	"brkt/housekeeper/cloud"
	"log"
	"time"
)

const (
	dateFormatLayout = "2006-01-02"
)

// Reporter is a general interface that can be implemented
// for both AWS and GCP to generate expense reports.
type Reporter interface {
	GenerateReport(startDate, endDate time.Time) Report
}

// Report contains a collection of items, and some metadata
// about when the items were collected and which dates they
// span. The report struct also has methods to help work with
// all the items.
type Report struct {
	StartDate    time.Time
	EndDate      time.Time
	CreationDate time.Time
	Items        []ReportItem
}

// TotalCost returns the total cost for all items
func (r *Report) TotalCost() float64 {
	total := 0.0
	for i := range r.Items {
		total += r.Items[i].Cost
	}
	return total
}

// TotalPerOwner return a map with the total cost for each owner.
func (r *Report) TotalPerOwner() map[string]float64 {
	result := make(map[string]float64)
	for i := range r.Items {
		result[r.Items[i].OwnerID] += r.Items[i].Cost
	}
	return result
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

// ReportItem represent a single item in a report. This is usually
// the cost for a specific service for a certain user in a certain
// account/project.
type ReportItem struct {
	OwnerID     string
	Description string
	Cost        float64
}

// NewReporter initializes a new billing reporter for the specified CSP
func NewReporter(c cloud.CSP) Reporter {
	switch c {
	case cloud.AWS:
		log.Println("Initializing AWS billing reporter")
		reporter := &awsReporter{}
		return reporter
	case cloud.GCP:
		log.Fatalln("Unfortunately, GCP is currently not supported")
	default:
		log.Fatalln("Invalid CSP specified")
	}
	return nil
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
