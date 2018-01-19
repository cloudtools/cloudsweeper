package notify

import (
	"brkt/housekeeper/cloud"
	"brkt/housekeeper/cloud/billing"
	"brkt/housekeeper/cloud/filter"
	"brkt/housekeeper/housekeeper"
	"brkt/housekeeper/mailer"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"
)

const (
	smtpUserKey     = "SMTP_USER"
	smtpPassKey     = "SMTP_PASS"
	mailDisplayName = "HouseKeeper"
)

// OlderThanXMonths sends out an email notification to all specified owners
// about all of their resources older than the specified amount of months.
func OlderThanXMonths(months int, csp cloud.CSP, owners housekeeper.Owners) {
	mngr := cloud.NewManager(csp, owners.AllIDs()...)
	all := mngr.AllResourcesPerAccount()
	allBuckets := mngr.BucketsPerAccount()
	ownerNames := owners.IDToName()
	for owner, resources := range all {
		ownerName := convertEmailExceptions(ownerNames[owner])
		fil := filter.New()
		fil.AddGeneralRule(filter.OlderThanXMonths(months))
		mailHolder := struct {
			Owner     string
			Instances []cloud.Instance
			Images    []cloud.Image
			Snapshots []cloud.Snapshot
			Volumes   []cloud.Volume
			Buckets   []cloud.Bucket
		}{
			ownerName,
			fil.FilterInstances(resources.Instances),
			fil.FilterImages(resources.Images),
			fil.FilterSnapshots(resources.Snapshots),
			fil.FilterVolumes(resources.Volumes),
			[]cloud.Bucket{},
		}
		if bucks, ok := allBuckets[owner]; ok {
			mailHolder.Buckets = fil.FilterBuckets(bucks)
		}

		oldResourceCount := len(mailHolder.Images) + len(mailHolder.Instances) + len(mailHolder.Snapshots) + len(mailHolder.Volumes)
		if oldResourceCount > 0 {
			// Now send email
			mailClient := getMailClient()
			mailContent, err := generateMail(mailHolder)
			if err != nil {
				log.Fatalln("Could not generate email:", err)
			}
			ownerMail := fmt.Sprintf("%s@brkt.com", mailHolder.Owner)
			log.Printf("Notifying %s about old resources\n", ownerMail)
			title := fmt.Sprintf("You have %d old resource(s) (%s)", oldResourceCount, time.Now().Format("2006-01-02"))
			mailClient.SendEmail("hsson@brkt.com", title, mailContent) // TODO: Use actual email
		}
	}
}

func generateMail(data interface{}) (string, error) {
	t := template.New("emailTemplate").Funcs(extraTemplateFunctions())
	t, err := t.Parse(oldResourcesTemplate)
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

// This function will convert some edge case names to their proper
// email alias
func convertEmailExceptions(oldName string) (newName string) {
	switch oldName {
	case "qa-solo":
		return "qa"
	default:
		return oldName
	}
}

func getMailClient() mailer.Client {
	username, exists := os.LookupEnv(smtpUserKey)
	if !exists {
		log.Fatalf("%s is required\n", smtpUserKey)
	}
	password, exists := os.LookupEnv(smtpPassKey)
	if !exists {
		log.Fatalf("%s is required\n", smtpPassKey)
	}
	return mailer.NewClient(username, password, mailDisplayName)
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
		"even": func(num int) bool { return num%2 == 0 },
		"yesno": func(b bool) string {
			if b {
				return "Yes"
			}
			return "No"
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
	}
}
