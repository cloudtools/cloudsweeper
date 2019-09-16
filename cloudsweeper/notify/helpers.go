// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/agaridata/cloudsweeper/cloud"
	"github.com/agaridata/cloudsweeper/cloud/billing"
	"github.com/agaridata/cloudsweeper/cloud/filter"
	"github.com/agaridata/cloudsweeper/mailer"
)

var emailEdgeCases = map[string]string{} // Use this map to fix bad mappings between usernames and email aliases

func generateMail(data interface{}, templateString string) (string, error) {
	t := template.New("emailTemplate").Funcs(extraTemplateFunctions())
	t, err := t.Parse(templateString)
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	err = t.Execute(&result, data)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// This function will convert some edge case emails to their proper
// email. This is useful if some user doesn't share the common org domain
func convertEmailExceptions(oldMail string) string {
	name, hasEdgeCase := emailEdgeCases[oldMail]
	if hasEdgeCase {
		return name
	}
	return oldMail
}

func getMailClient(notifyClient *Client) mailer.Client {
	username := notifyClient.config.SMTPUsername
	password := notifyClient.config.SMTPPassword
	server := notifyClient.config.SMTPServer
	port := notifyClient.config.SMTPPort
	from := notifyClient.config.MailFrom
	displayName := notifyClient.config.DisplayName
	return mailer.NewClient(username, password, displayName, from, server, port)
}

func timeUntilEarliestDeletion(resourceCollection cloud.AllResourceCollection) string {

	// Initialize to something bigger than time to deletion
	earliestTime := time.Now().AddDate(0, 0, 99999)

	resources := []cloud.Resource{}
	for _, res := range resourceCollection.Instances {
		resources = append(resources, res.(cloud.Resource))
	}
	for _, res := range resourceCollection.Images {
		resources = append(resources, res.(cloud.Resource))
	}
	for _, res := range resourceCollection.Snapshots {
		resources = append(resources, res.(cloud.Resource))
	}
	for _, res := range resourceCollection.Volumes {
		resources = append(resources, res.(cloud.Resource))
	}
	for _, res := range resourceCollection.Buckets {
		resources = append(resources, res.(cloud.Resource))
	}

	for _, res := range resources {
		tempTag, exists := res.Tags()["cloudsweeper-delete-at"]
		if !exists {
			continue
		}
		tempTime, err := time.Parse(time.RFC3339, tempTag)
		if err != nil {
			continue
		}
		if earliestTime.After(tempTime) {
			earliestTime = tempTime
		}
	}

	hours := int(time.Until(earliestTime).Hours())
	return strconv.Itoa(hours)
}

func accumulatedCost(res cloud.Resource) float64 {
	days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
	costPerDay := billing.ResourceCostPerDay(res)
	return days * costPerDay
}

func extraTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"fdate": func(t time.Time, format string) string { return t.Format(format) },
		"daysrunning": func(t time.Time) string {
			if (t == time.Time{}) {
				return "never"
			}
			days := int(time.Now().Sub(t).Hours() / 24.0)
			switch days {
			case 0:
				return "today"
			case 1:
				return "yesterday"
			default:
				return fmt.Sprintf("%d days ago", days)
			}
		},
		// TODO: this should be configurable
		"modifiedInTheLast6Months": func(t time.Time) string {
			if time.Now().Before(t.AddDate(0, 6, 0)) {
				return "true"
			}
			return "false"
		},

		"even": func(num int) bool { return num%2 == 0 },
		"yesno": func(b bool) string {
			if b {
				return "Yes"
			}
			return "No"
		},
		"whitelisted": func(res cloud.Resource) bool {
			return filter.IsWhitelisted(res)
		},
		"accucost": func(res cloud.Resource) string {
			totalCost := accumulatedCost(res)
			return fmt.Sprintf("$%.2f", totalCost)
		},
		"bucketcost": func(res cloud.Bucket) float64 {
			return billing.BucketPricePerMonth(res)
		},
		"instname": func(inst cloud.Instance) string {
			if inst.CSP() == cloud.AWS {
				name, exist := inst.Tags()["Name"]
				if exist {
					return name
				}
				return ""

			} else if inst.CSP() == cloud.GCP {
				return inst.ID()
			} else {
				return ""
			}
		},
		"productname": func(res cloud.Resource) string {
			product, exist := res.Tags()["product"]
			if exist {
				return product
			}
			return ""
		},
		"rolename": func(res cloud.Resource) string {
			role, exist := res.Tags()["role"]
			if exist {
				return role
			}
			return ""
		},
		"maybeRealName": func(account string, accountToUser map[string]string) string {
			if name, ok := accountToUser[account]; ok {
				return name
			}
			return account
		},
		"prettyTag": func(key, val string) string {
			if val == "" {
				return key
			}
			return fmt.Sprintf("%s: %s", key, val)
		},
		"deletedate": func(res cloud.Resource, format string) string {
			tag, exist := res.Tags()["cloudsweeper-delete-at"]
			if !exist {
				return ""
			}
			t, err := time.Parse(time.RFC3339, tag)
			if err != nil {
				return ""
			}
			return t.Format(format)
		},
		// TODO: This isn't pretty whatsoever
		"timeUntilDelete": func(instances []cloud.Instance, images []cloud.Image, snapshots []cloud.Snapshot, volumes []cloud.Volume, buckets []cloud.Bucket) string {
			allResources := cloud.AllResourceCollection{}
			allResources.Instances = instances
			allResources.Images = images
			allResources.Snapshots = snapshots
			allResources.Volumes = volumes
			allResources.Buckets = buckets
			return timeUntilEarliestDeletion(allResources)
		},
	}
}
