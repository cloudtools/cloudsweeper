// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package filter

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cloudtools/cloudsweeper/cloud"
)

const (
	// WhitelistTagKey marks a resource to not matched by filter
	WhitelistTagKey = "cloudsweeper-whitelisted"
	// LifetimeTagKey marks a resource to be cleaned up after X days
	LifetimeTagKey = "cloudsweeper-lifetime"
	// ExpiryTagKey marks a resource to be cleaned up at the specified date (YYYY-MM-DD)
	ExpiryTagKey = "cloudsweeper-expiry"
	// DeleteTagKey marks a resource for deletion. This is used internally by houskeeper
	// to keep track of resources that should be cleaned up, but was not explicitly tagged
	// by the resource owner.
	DeleteTagKey = "cloudsweeper-delete-at"
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

// TaggedForCleanup checks if resource is already tagged for cleanup
func TaggedForCleanup() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return HasTag(DeleteTagKey)(r)
	}
}

// OlderThanXHours returns a resource that is older than the
// specified amount of hours.
func OlderThanXHours(hours int) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return time.Now().After(r.CreationTime().Add(time.Duration(hours) * time.Hour))
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
		return strings.Contains(strings.ToLower(name), strings.ToLower(contains))
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
		for key := range r.Tags() {
			if strings.ToLower(key) == strings.ToLower(tagKey) {
				return true
			}
		}
		return false
	}
}

// IsUntaggedWithException checks if a resource is untagged with the exception of a specific tag
func IsUntaggedWithException(exceptionTag string) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		if len(r.Tags()) == 0 {
			return true
		} else if len(r.Tags()) == 1 {
			return HasTag(exceptionTag)(r)
		}
		return false
	}
}

// IsPublic checks if a resource is public
func IsPublic() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		return r.Public()
	}
}

// LifetimeExceeded check if a resource have the lifetime tag,
// with the format "cloudsweeper-lifetime: days-X" (where X is the amount of
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
// expiry tag has the format "cloudsweeper-expiry: 2018-06-17".
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

// DeleteWithinXHours checks if a resources is marked for deletion and if
// it's about to be deleted within the specified amount of hours. This also
// includes resources which deletion time is passed.
func DeleteWithinXHours(hours int) func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		deleteTimeString, hasDeletion := r.Tags()[DeleteTagKey]
		if !hasDeletion {
			return false
		}
		deleteTime, err := time.Parse(time.RFC3339, deleteTimeString)
		if err != nil {
			log.Printf("%s has malformed deletion tag: %s\n", r.ID(), deleteTimeString)
			return false
		}
		within := deleteTime.Add(-(time.Duration(hours) * time.Hour))
		return time.Now().After(within)
	}
}

// DeleteAtPassed checks is the delete-at time for a resource has passed. The
// delete tag has the format "cloudsweeper-delete-at: 2018-01-25T16:51:39-08:00".
func DeleteAtPassed() func(cloud.Resource) bool {
	return func(r cloud.Resource) bool {
		deleteAt, exist := r.Tags()[DeleteTagKey]
		if !exist {
			return false
		}
		deleteAtTime, err := time.Parse(time.RFC3339, deleteAt)
		if err != nil {
			log.Printf("%s has malformed deletion tag: %s\n", r.ID(), deleteAt)
			return false
		}
		return time.Now().After(deleteAtTime)
	}
}

// Below are volume rules

// IsUnattached checks if volume is not attached to an instance
func IsUnattached() func(cloud.Volume) bool {
	return func(v cloud.Volume) bool {
		return !v.Attached()
	}
}

// Below are snapshot rules

// IsInUse checks if the snapshot is currently being used by an AMI
func IsInUse() func(cloud.Snapshot) bool {
	return func(s cloud.Snapshot) bool {
		return s.InUse()
	}
}

// IsNotInUse is the opposite of IsInUse
func IsNotInUse() func(cloud.Snapshot) bool {
	return func(s cloud.Snapshot) bool {
		return !(IsInUse())(s)
	}
}

// Below are bucket rules

// NotModifiedInXDays returns bucket which have not had any modification
// to them within X days.
func NotModifiedInXDays(days int) func(cloud.Bucket) bool {
	return func(b cloud.Bucket) bool {
		return time.Now().After(b.LastModified().AddDate(0, 0, days))
	}
}
