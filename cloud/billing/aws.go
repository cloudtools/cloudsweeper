// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package billing

import (
	"archive/zip"
	"brkt/cloudsweeper/cloud"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	awsBillingAccount      = "992270393355"
	awsBillingBucket       = "aws-prod-billing"
	awsBillingBucketRegion = "us-east-1"
	awsCSVDateFormat       = "2006-01-02"
	awsCSVNameFormat       = "%s-aws-billing-detailed-line-items-%d-%02d.csv.zip"
)

type awsReporter struct {
	csp cloud.CSP
}

func (r *awsReporter) GenerateReport(start time.Time) Report {
	report := Report{}
	report.CSP = r.csp

	name := fmt.Sprintf(awsCSVNameFormat, awsBillingAccount, start.Year(), start.Month())
	csvFile, err := getCSVFromS3(name)
	if err != nil {
		log.Println("Failed to get", name, ":", err)
	}
	err = processAwsCsv(&report, csvFile, true)
	if err != nil {
		log.Println("Failed to process CSV", name)
	}

	return report
}

func processAwsCsv(report *Report, csvFile *csv.Reader, allowFailed bool) error {
	csvHeaders := make(map[string]int)
	line := 0
	for {
		record, err := csvFile.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if allowFailed {
				log.Printf("Failed reading line %d, continuing...\n%s", line, err)
			} else {
				return err
			}
		}
		if line == 0 {
			csvHeaders = updateCsvHeaders(record)
			line++
			continue
		}
		if record[csvHeaders["RecordType"]] != "LineItem" {
			// Ignore lines with AccountTotal (so we don't count it twice)
			line++
			continue
		}

		reportItem := ReportItem{}
		reportItem.Owner = record[csvHeaders["LinkedAccountId"]]
		reportItem.Description = record[csvHeaders["ItemDescription"]]
		cost := record[csvHeaders["UnBlendedCost"]]
		cost = strings.Replace(cost, ",", "", -1)
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
		line++
	}
}

func getCSVFromS3(name string) (*csv.Reader, error) {
	tmpZip := filepath.Join(os.TempDir(), name)
	f, err := os.Create(tmpZip)
	if err != nil {
		log.Println("Could not create file in temp directory")
		return nil, err
	}
	sess := session.Must(session.NewSession())
	sess.Config.Region = aws.String(awsBillingBucketRegion)
	downloader := s3manager.NewDownloader(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(awsBillingBucket),
		Key:    aws.String(name),
	}
	_, err = downloader.Download(f, input)
	if err != nil {
		log.Println("Could not find bucket")
		return nil, err
	}
	reader, err := zip.OpenReader(tmpZip)
	if err != nil {
		log.Println("Could not read ZIP file")
		return nil, err
	}
	//defer reader.Close()
	if len(reader.File) == 0 {
		return nil, errors.New("Zip file was empty")
	}
	file := reader.File[0]
	log.Println("Using", file.Name)
	rc, err := file.Open()
	if err != nil {
		log.Println("Billing CSV is corrupt:", err)
		return nil, err
	}
	return csv.NewReader(rc), nil
}
