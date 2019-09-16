// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package billing

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/agaridata/cloudsweeper/cloud"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	gcpCSVNameFormat = "%s-%d-%02d-%02d.csv"
)

type gcpReporter struct {
	csp           cloud.CSP
	bucket        string
	csvNamePrefix string
}

func (r *gcpReporter) GenerateReport(start time.Time) Report {
	report := Report{}
	report.CSP = r.csp

	ctx := context.Background()
	credsFilePath, exist := os.LookupEnv(cloud.GcpCredentialsFileKey)
	if !exist {
		log.Fatalln("No GCP credentials specified!")
	}
	if _, err := os.Stat(credsFilePath); os.IsNotExist(err) {
		log.Fatalln(credsFilePath, "is not a file!")
	}
	opt := option.WithServiceAccountFile(credsFilePath)
	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		log.Printf("Could not initialize storage service:\n%s\n", err)
		return report
	}

	for d := start; d.Month() == start.Month(); d = d.AddDate(0, 0, 1) {
		name := fmt.Sprintf(gcpCSVNameFormat, r.csvNamePrefix, start.Year(), start.Month(), d.Day())
		log.Println("Getting", name)
		obj := client.Bucket(r.bucket).Object(name)
		if err := processObjectHandle(ctx, obj, &report, true); err != nil {
			log.Println(err, "- skipping...")
			break
		}
	}
	return report
}

func processObjectHandle(ctx context.Context, obj *storage.ObjectHandle, report *Report, allowFailed bool) error {
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()
	csvFile := csv.NewReader(reader)
	i := 0
	csvHeaders := make(map[string]int)
	for {
		record, err := csvFile.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if allowFailed {
				log.Printf("Failed reading line %d, continuing...\n%s", i, err)
			} else {
				return err
			}
		}
		if i == 0 {
			csvHeaders = updateCsvHeaders(record)
			i++
			continue
		}

		reportItem := ReportItem{}
		reportItem.Owner = record[csvHeaders["Project ID"]]
		reportItem.Description = record[csvHeaders["Description"]]
		cost := record[csvHeaders["Cost"]]
		costNumber, err := strconv.ParseFloat(cost, 64)
		if err != nil {
			if allowFailed {
				log.Println("Could not convert cost to float:", cost)
			} else {
				return err
			}
		}
		reportItem.Cost = costNumber
		report.Items = append(report.Items, reportItem)
		i++
	}
}
