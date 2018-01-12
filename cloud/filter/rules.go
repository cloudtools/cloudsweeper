package filter

import (
	"brkt/housekeeper/cloud"
	"strings"
	"time"
)

// Below are general rules

// OlderThanXDays return a resource that is older than the
// specified amount of days
func OlderThanXDays(days time.Duration) func(cloud.Resource) bool {
	then := time.Now().Add(-(days * 24 * time.Hour))
	return func(r cloud.Resource) bool {
		return r.CreationTime().Before(then)
	}
}

// OlderThanXWeeks return a resource that is older than the
// specified amount of weeks
func OlderThanXWeeks(weeks time.Duration) func(cloud.Resource) bool {
	then := time.Now().Add(-(weeks * time.Hour * 24 * 7))
	return func(r cloud.Resource) bool {
		return r.CreationTime().Before(then)
	}
}

// OlderThanXYears return a resource that is older than the
// specified amount of years
func OlderThanXYears(years int) func(cloud.Resource) bool {
	year, month, day := time.Now().Date()
	year -= years
	return func(r cloud.Resource) bool {
		then := time.Date(year, month, day, 0, 0, 0, 0, r.CreationTime().Location())
		return r.CreationTime().Before(then)
	}
}

// NameContains checks if a resource's name contains a
// specified substring
func NameContains(contains string) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		name := ""
		if n, ok := r.Tags()["Name"]; ok {
			name = n
		}
		return strings.Contains(name, contains)
	}
}
