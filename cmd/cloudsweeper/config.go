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
	"warning-hours":       lookup{"CS_WARNING_HOURS", "48"},
	"display-name":        lookup{"CS_DISPLAY_NAME", "Cloudsweeper"},
	"summary-addressee":   lookup{"CS_SUMMARY_ADDRESSEE", ""},
	"total-sum-addressee": lookup{"CS_TOTAL_SUM_ADDRESSEE", ""},
	"mail-domain":         lookup{"CS_EMAIL_DOMAIN", ""},

	// Setup variables
	"aws-master-arn": lookup{"CS_MASTER_ARN", ""},
}

func loadConfig() {
	var err error
	config, err = godotenv.Read(configFileName)
	if err != nil {
		log.Fatalf("Could not load config file '%s': %s", configFileName, err)
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
