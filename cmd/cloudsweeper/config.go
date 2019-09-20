// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/joho/godotenv"
)

const optionalDefault = "<optional>"

type lookup struct {
	confKey      string
	defaultValue string
}

var configMapping = map[string]lookup{
	// General variables
	"csp":      lookup{"CS_CSP", "aws"},
	"org-file": lookup{"CS_ORG_FILE", "organization.json"},

	// Billing related
	"billing-account":       lookup{"CS_BILLING_ACCOUNT", ""},
	"billing-bucket-region": lookup{"CS_BILLING_BUCKET_REGION", ""},
	"billing-csv-prefix":    lookup{"CS_BILLING_CSV_PREFIX", ""},
	"billing-bucket":        lookup{"CS_BILLING_BUCKET_NAME", ""},
	"billing-sort-tag":      lookup{"CS_BILLING_SORT_TAG", optionalDefault},

	// Email variables
	"smtp-username": lookup{"CS_SMTP_USER", ""},
	"smtp-password": lookup{"CS_SMTP_PASSWORD", ""},
	"smtp-server":   lookup{"CS_SMTP_SERVER", ""},
	"smtp-port":     lookup{"CS_SMTP_PORT", "587"},

	// Notifying specific variables
	"warning-hours":            lookup{"CS_WARNING_HOURS", "48"},
	"display-name":             lookup{"CS_DISPLAY_NAME", "Cloudsweeper"},
	"mail-from":                lookup{"CS_MAIL_FROM", ""},
	"billing-report-addressee": lookup{"CS_BILLING_REPORT_ADDRESSEE", ""},
	"total-sum-addressee":      lookup{"CS_TOTAL_SUM_ADDRESSEE", ""},
	"mail-domain":              lookup{"CS_EMAIL_DOMAIN", ""},

	// Setup variables
	"aws-master-arn": lookup{"CS_MASTER_ARN", ""},

	// Clean thresholds
	"clean-untagged-older-than-days":    lookup{"CLEAN_UNTAGGED_OLDER_THAN_DAYS", "30"},
	"clean-instances-older-than-days":   lookup{"CLEAN_INSTANCES_OLDER_THAN_DAYS", "182"},
	"clean-images-older-than-days":      lookup{"CLEAN_IMAGES_OLDER_THAN_DAYS", "182"},
	"clean-snapshots-older-than-days":   lookup{"CLEAN_SNAPSHOTS_OLDER_THAN_DAYS", "182"},
	"clean-unattached-older-than-days": lookup{"CLEAN_UNATTACHED_OLDER_THAN_DAYS", "30"},
	"clean-bucket-not-modified-days":    lookup{"CLEAN_BUCKET_NOT_MODIFIED_DAYS", "182"},
	"clean-bucket-older-than-days":      lookup{"CLEAN_BUCKET_OLDER_THAN_DAYS", "7"},
	"clean-keep-n-component-images":     lookup{"CLEAN_KEEP_N_COMPONENT_IMAGES", "2"},

	//  Notify thresholds
	"notify-untagged-older-than-days":   lookup{"NOTIFY_UNTAGGED_OLDER_THAN_DAYS", "14"},
	"notify-instances-older-than-days":  lookup{"NOTIFY_INSTANCES_OLDER_THAN_DAYS", "30"},
	"notify-images-older-than-days":     lookup{"NOTIFY_IMAGES_OLDER_THAN_DAYS", "30"},
	"notify-unattached-older-than-days": lookup{"NOTIFY_UNATTACHED_OLDER_THAN_DAYS", "30"},
	"notify-snapshots-older-than-days":  lookup{"NOTIFY_SNAPSHOTS_OLDER_THAN_DAYS", "30"},
	"notify-buckets-older-than-days":    lookup{"NOTIFY_BUCKETS_OLDER_THAN_DAYS", "30"},
	"notify-whitelist-older-than-days":  lookup{"NOTIFY_WHITELIST_OLDER_THAN_DAYS", "182"},
	"notify-dnd-older-than-days":        lookup{"NOTIFY_DND_OLDER_THAN_DAYS", "7"},
}

func loadConfig() {
	var err error
	config, err = godotenv.Read(configFileName)
	if err != nil {
		log.Fatalf("Could not load config file '%s': %s", configFileName, err)
	}
}

func loadThresholds() {
	for _, v := range thnames {
		thresholds[v] = findConfigInt(v)
	}
}

func findConfig(name string) string {
	if _, exist := configMapping[name]; !exist {
		log.Fatalf("Unknown config option: %s", name)
	}
	flagVal := flag.Lookup(name).Value.String()
	if flagVal != "" {
		return flagVal
	} else if confVal, ok := config[configMapping[name].confKey]; ok && confVal != "" {
		maybeNoValExit(confVal, name)
		return confVal
	} else {
		defaultVal := configMapping[name].defaultValue
		if defaultVal == optionalDefault {
			return ""
		}
		maybeNoValExit(defaultVal, name)
		return defaultVal
	}
}

func maybeNoValExit(val, name string) {
	if val == "" {
		log.Fatalf("No value specified for --%s", name)
	}
}

func findConfigInt(name string) int {
	val := findConfig(name)
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Value specified for %s is not an integer", name)
	}
	return i
}
