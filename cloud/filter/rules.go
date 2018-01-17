package filter

import (
	"brkt/housekeeper/cloud"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	// WhitelistTagKey marks a resource to not matched by filter
	WhitelistTagKey = "whitelisted"
	// LifetimeTagKey marks a resource to be cleaned up after X days
	LifetimeTagKey = "housekeeper-lifetime"
	// ExpiryTagKey marks a resource to be cleaned up at the specified date (YYYY-MM-DD)
	ExpiryTagKey = "housekeeper-expiry"
	// ExpiryTagValueFormat is the format to use when setting expiry date
	ExpiryTagValueFormat = "2006-01-02" // Used to parse string
)

// Below are general rules

// Negate will simply negate another rule
func Negate(funcToNegate func(r cloud.Resource) bool) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return !funcToNegate(r)
	}
}

// OlderThanXDays return a resource that is older than the
// specified amount of days
func OlderThanXDays(days int) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return time.Now().After(r.CreationTime().AddDate(0, 0, days))
	}
}

// OlderThanXMonths return a resource that is older than the
// specified amount of months
func OlderThanXMonths(months int) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return time.Now().After(r.CreationTime().AddDate(0, months, 0))
	}
}

// OlderThanXYears return a resource that is older than the
// specified amount of years
func OlderThanXYears(years int) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return time.Now().After(r.CreationTime().AddDate(years, 0, 0))
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

// IDMatches checks if a resource's ID matches any of the
// specified IDs.
func IDMatches(ids ...string) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		for i := range ids {
			if ids[i] == r.ID() {
				return true
			}
		}
		return false
	}
}

// HasTag checks if a resource have a specified tag or not
func HasTag(tagKey string) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		_, ok := r.Tags()[tagKey]
		return ok
	}
}

// IsPublic checks if a resource is public
func IsPublic() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return r.Public()
	}
}

// LifetimeExceeded check if a resource have the lifetime tag,
// with the format "housekeeper-lifetime: days-X" (where X is the amount of
// days to keep the resource). If the lifetime is passed, then
// this resource should be included in the filter.
func LifetimeExceeded() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		lifetime, hasLifetime := r.Tags()[LifetimeTagKey]
		if !hasLifetime {
			// If resource doesn't have the lifetime tag then don't include it
			return false
		}
		days := strings.Split(lifetime, "-")
		if len(days) != 2 {
			// Lifetime tag is not on the correct format
			log.Printf("%s have an incorrect lifetime tag: %s", r.ID(), lifetime)
			return false
		}
		numberOfDays, err := strconv.Atoi(days[1])
		if err != nil {
			// Lifetime tag is not on the correct format
			log.Printf("%s have an incorrect lifetime tag: %s", r.ID(), lifetime)
			return false
		}
		expiery := r.CreationTime().Add(time.Hour * 24 * time.Duration(numberOfDays))
		return time.Now().After(expiery)
	}
}

// ExpiryDatePassed checks is the expiry date for a resource has passed. The
// expiry tag has the format "housekeeper-expiry: 2018-06-17".
func ExpiryDatePassed() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		expiryVal, hasExpiry := r.Tags()[ExpiryTagKey]
		if !hasExpiry {
			// Don't include resource that doesn't have expiry tag
			return false
		}
		expiryDate, err := time.Parse(ExpiryTagValueFormat, expiryVal)
		if err != nil {
			log.Printf("%s has incorrect expiry tag:%s", r.ID(), expiryVal)
			return false
		}
		return time.Now().After(expiryDate)
	}
}
