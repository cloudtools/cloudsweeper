// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/cloudtools/cloudsweeper/cloud"
	"github.com/cloudtools/cloudsweeper/cloud/billing"
	"github.com/cloudtools/cloudsweeper/cloud/filter"
	"github.com/cloudtools/cloudsweeper/mailer"
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
			days := time.Now().Sub(res.CreationTime()).Hours() / 24.0
			costPerDay := billing.ResourceCostPerDay(res)
			return fmt.Sprintf("$%.2f", days*costPerDay)
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
	}
}
