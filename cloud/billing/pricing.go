// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package billing

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/private/protocol"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/cloudtools/cloudsweeper/cloud"
)

const (
	gcpBucketPerGBMonth = 0.026

	assumeRoleARNTemplate = "arn:aws:iam::%s:role/Cloudsweeper"
)

type instanceKeyPair struct {
	Region, InstanceType string
}

type priceMap map[instanceKeyPair]float64

var (
	awsPrices priceMap
)

var generalInstanceFilters = []*pricing.Filter{
	{
		Field: aws.String("operatingSystem"),
		Type:  aws.String("TERM_MATCH"),
		Value: aws.String("Linux"),
	},
	{
		Field: aws.String("operation"),
		Type:  aws.String("TERM_MATCH"),
		Value: aws.String("RunInstances"),
	},
	{
		Field: aws.String("capacitystatus"),
		Type:  aws.String("TERM_MATCH"),
		Value: aws.String("Used"),
	},
	{
		Field: aws.String("tenancy"),
		Type:  aws.String("TERM_MATCH"),
		Value: aws.String("Shared"),
	},
}

var awsRegionIDToNameMap = map[string]string{
	"us-east-2":      "US East (Ohio)",
	"us-east-1":      "US East (N. Virginia)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-northeast-3": "Asia Pacific (Osaka-Local)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ca-central-1":   "Canada (Central)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "EU (Frankfurt)",
	"eu-west-1":      "EU (Ireland)",
	"eu-west-2":      "EU (London)",
	"eu-west-3":      "EU (Paris)",
	"eu-north-1":     "EU (Stockholm)",
	"sa-east-1":      "South America (Sao Paulo)",
	"us-gov-east-1":  "AWS GovCloud (US-East)",
	"us-gov-west-1":  "AWS GovCloud (US-West)",
}

var awsS3StorageCostMap = map[string]float64{
	"StandardStorage":             0.023,
	"IntelligentTieringFAStorage": 0.023,
	"IntelligentTieringIAStorage": 0.0125,
	"StandardIAStorage":           0.0125,
	"OneZoneIAStorage":            0.01,
	"ReducedRedundancyStorage":    0.023, // TODO: double check this
	"GlacierStorage":              0.004,
}

// Storage cost per GB per day
var awsStorageCostMap = map[string]float64{
	"standard": 0.05 / 30.0,
	"gp2":      0.1 / 30.0,
	"io1":      0.125 / 30.0,
	"st1":      0.045 / 30.0,
	"sc1":      0.025 / 30.0,
	"snapshot": 0.05 / 30.0,
}

// Storage cost per GB per day
var gcpStorageCostGBDayMap = map[string]float64{
	"pd-ssd":      0.170 / 30.0,
	"pd-standard": 0.040 / 30.0,
	"snapshot":    0.026 / 30.0,
}

var gcpInstanceCostPerHourMap = map[string]float64{
	"n1-standard-1":  0.0475,
	"n1-standard-2":  0.0950,
	"n1-standard-4":  0.1900,
	"n1-standard-8":  0.3800,
	"n1-standard-16": 0.7600,
	"n1-standard-32": 1.5200,
	"n1-standard-64": 3.0400,
	"n1-standard-96": 4.5600,

	"n1-highmem-2":  0.1184,
	"n1-highmem-4":  0.2368,
	"n1-highmem-8":  0.4736,
	"n1-highmem-16": 0.9472,
	"n1-highmem-32": 1.8944,
	"n1-highmem-64": 3.7888,
	"n1-highmem-96": 5.6832,

	"n1-highcpu-2":  0.0709,
	"n1-highcpu-4":  0.1418,
	"n1-highcpu-8":  0.2836,
	"n1-highcpu-16": 0.5672,
	"n1-highcpu-32": 1.1344,
	"n1-highcpu-64": 2.2688,
	"n1-highcpu-96": 3.4020,

	"f1-micro": 0.0076,
	"g1-small": 0.0257,

	"n1-megamem-96": 10.6740,
}

// ResourceCostPerDay returns the daily cost of a resource in USD
func ResourceCostPerDay(resource cloud.Resource) float64 {
	if inst, ok := resource.(cloud.Instance); ok {
		return InstancePricePerHour(inst) * 24.0
	} else if vol, ok := resource.(cloud.Volume); ok {
		return VolumeCostPerDay(vol)
	} else if img, ok := resource.(cloud.Image); ok {
		return ImageCostPerDay(img)
	} else if snap, ok := resource.(cloud.Snapshot); ok {
		return SnapshotCostPerDay(snap)
	} else {
		log.Println("Resource was neither instance, volume, image or snapshot")
		return 0.0
	}
}

// VolumeCostPerDay returns the daily cost in USD for a
// certain volume
func VolumeCostPerDay(volume cloud.Volume) float64 {
	if volume.CSP() == cloud.AWS {
		price, ok := awsStorageCostMap[volume.VolumeType()]
		if !ok {
			log.Fatalf("Could not find price for %s in AWS", volume.VolumeType())
			return 0.0
		}
		return price * float64(volume.SizeGB())
	} else if volume.CSP() == cloud.GCP {
		price, ok := gcpStorageCostGBDayMap[volume.VolumeType()]
		if !ok {
			log.Fatalf("Could not find price for %s in GCP", volume.VolumeType())
			return 0.0
		}
		return price * float64(volume.SizeGB())
	}
	log.Panicln("Unsupported CSP:", volume.CSP())
	return 0.0
}

// SnapshotCostPerDay returns the daily cost in USD for a
// certain snapshot
func SnapshotCostPerDay(snapshot cloud.Snapshot) float64 {
	if snapshot.CSP() == cloud.AWS {
		return awsStorageCostMap["snapshot"] * float64(snapshot.SizeGB())
	} else if snapshot.CSP() == cloud.GCP {
		price := gcpStorageCostGBDayMap["snapshot"]
		return price * float64(snapshot.SizeGB())
	}
	log.Panicln("Unsupported CSP:", snapshot.CSP())
	return 0.0
}

// ImageCostPerDay returns the daily cost in USD for a
// certain image
func ImageCostPerDay(image cloud.Image) float64 {
	if image.CSP() == cloud.AWS {
		return awsStorageCostMap["snapshot"] * float64(image.SizeGB())
	} else if image.CSP() == cloud.GCP {
		price := gcpStorageCostGBDayMap["snapshot"]
		return price * float64(image.SizeGB())
	}
	log.Panicln("Unsupported CSP:", image.CSP())
	return 0.0
}

// InstancePricePerHour will return the hourly price in USD for a
// specified instance.
func InstancePricePerHour(instance cloud.Instance) float64 {
	if instance.CSP() == cloud.AWS {
		return awsInstancePricePerHour(instance)
	} else if instance.CSP() == cloud.GCP {
		price, ok := gcpInstanceCostPerHourMap[instance.InstanceType()]
		if !ok {
			log.Fatalf("Could not find price for %s in GCP", instance.InstanceType())
			return 0.0
		}
		return price
	}
	log.Panicln("Unsupported CSP:", instance.CSP())
	return 0.0
}

// BucketPricePerMonth will return the monthly price in USD for a
// specified bucket. It will not take any account wide discounts
// that might have been collected for using a certain amount of
// storage every month.
func BucketPricePerMonth(bucket cloud.Bucket) float64 {
	if bucket.CSP() == cloud.AWS {
		price := 0.0
		for storageType, size := range bucket.StorageTypeSizesGB() {
			price += awsS3StorageCostMap[storageType] * size
		}
		return price
	} else if bucket.CSP() == cloud.GCP {
		return gcpBucketPerGBMonth * bucket.TotalSizeGB()
	}
	log.Panicln("Unsupported CSP:", bucket.CSP())
	return 0.0
}

// awsInstancePricePerHour will return the hourly price in USD for a
// specified instance type in a specified AWS region.
func awsInstancePricePerHour(instance cloud.Instance) float64 {
	if awsPrices == nil {
		awsPrices = make(priceMap)
	}
	// The price for this instance type/region has already been fetched before
	price, exist := awsPrices[instanceKeyPair{instance.Location(), instance.InstanceType()}]
	if exist {
		return price
	}

	sess := session.Must(session.NewSession())
	creds := stscreds.NewCredentials(sess, fmt.Sprintf(assumeRoleARNTemplate, instance.Owner()))
	svc := pricing.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String("us-east-1"), // pricing API is only available here
	})

	specificFilters := []*pricing.Filter{
		{
			Field: aws.String("instanceType"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String(instance.InstanceType()),
		},
		{
			Field: aws.String("location"),
			Type:  aws.String("TERM_MATCH"),
			Value: aws.String(awsRegionIDToNameMap[instance.Location()]),
		},
	}
	filters := append(generalInstanceFilters, specificFilters...)
	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String("AmazonEC2"),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}
	result, err := svc.GetProducts(input)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var listPrice rawAWSPrice
	rawListPriceJSON, err := protocol.EncodeJSONValue(result.PriceList[0], protocol.NoEscape)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = json.Unmarshal([]byte(rawListPriceJSON), &listPrice)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, term := range listPrice.Terms.OnDemand {
		for _, price := range term.PriceDimensions {
			key := instanceKeyPair{
				Region:       instance.Location(),
				InstanceType: instance.InstanceType(),
			}
			usd, err := strconv.ParseFloat(price.PricePerUnit.USD, 64)
			if err != nil {
				log.Fatalln("Could not convert price from AWS JSON", err)
			}
			if usd == 0.00 {
				log.Println("Price for", instance.InstanceType(), "in", instance.Location(), "is $0.00. Needs investigation!")
			}
			awsPrices[key] = usd
			continue
		}
	}

	price, exist = awsPrices[instanceKeyPair{instance.Location(), instance.InstanceType()}]
	if !exist {
		log.Fatalln("Could not fetch price for", instance.InstanceType(), "in", instance.Location())
	}
	return price
}

// Helper structs for parsing the JSON from AWS
type rawAWSPrice struct {
	Terms struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}
