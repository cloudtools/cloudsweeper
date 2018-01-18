package notify

import (
	"brkt/housekeeper/cloud"
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
	ownerNames := owners.IDToName()
	for owner, resources := range all {
		oldResources := cloud.ResourceCollection{}
		ownerName := convertEmailExceptions(ownerNames[owner])
		oldResources.Owner = ownerName
		fil := filter.New()
		fil.AddGeneralRule(filter.OlderThanXMonths(months))
		oldResources.Instances = fil.FilterInstances(resources.Instances)
		oldResources.Images = fil.FilterImages(resources.Images)
		oldResources.Snapshots = fil.FilterSnapshots(resources.Snapshots)
		oldResources.Volumes = fil.FilterVolumes(resources.Volumes)

		oldResourceCount := len(oldResources.Images) + len(oldResources.Instances) + len(oldResources.Snapshots) + len(oldResources.Volumes)
		if oldResourceCount > 0 {
			// Now send email
			mailClient := getMailClient()
			mailContent, err := generateMail(oldResources)
			if err != nil {
				log.Fatalln("Could not generate email:", err)
			}
			ownerMail := fmt.Sprintf("%s@brkt.com", oldResources.Owner)
			log.Printf("Notifying %s about old resources\n", ownerMail)
			title := fmt.Sprintf("You have %d old resources (%s)", oldResourceCount, time.Now().Format("2006-01-02"))
			mailClient.SendEmail("hsson@brkt.com", title, mailContent) // TODO: Use actual email
		}
	}
}

func generateMail(resources cloud.ResourceCollection) (string, error) {
	t := template.New("emailTemplate").Funcs(extraTemplateFunctions())
	t, err := t.Parse(oldResourcesTemplate)
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	err = t.Execute(&result, resources)
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
			return fmt.Sprintf("%.0f", time.Now().Sub(t).Hours()/24.0)
		},
		"even": func(num int) bool { return num%2 == 0 },
		"yesno": func(b bool) string {
			if b {
				return "Yes"
			}
			return "No"
		},
	}
}
