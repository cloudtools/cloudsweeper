// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package billing

import (
	"brkt/cloudsweeper/cloud"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

const (
	awsPricingURL       = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json"
	s3BucketPerGBMonth  = 0.023
	gcpBucketPerGBMonth = 0.026
)

var (
	awsPrices priceMap
)

var awsRegionNameToIDMap = map[string]string{
	"US East (Ohio)":             "us-east-2",
	"US East (N. Virginia)":      "us-east-1",
	"US West (N. California)":    "us-west-1",
	"US West (Oregon)":           "us-west-2",
	"Asia Pacific (Tokyo)":       "ap-northeast-1",
	"Asia Pacific (Seoul)":       "ap-northeast-2",
	"Asia Pacific (Osaka-Local)": "ap-northeast-3",
	"Asia Pacific (Mumbai)":      "ap-south-1",
	"Asia Pacific (Singapore)":   "ap-southeast-1",
	"Asia Pacific (Sydney)":      "ap-southeast-2",
	"Canada (Central)":           "ca-central-1",
	"China (Beijing)":            "cn-north-1",
	"China (Ningxia)":            "cn-northwest-1",
	"EU (Frankfurt)":             "eu-central-1",
	"EU (Ireland)":               "eu-west-1",
	"EU (London)":                "eu-west-2",
	"EU (Paris)":                 "eu-west-3",
	"South America (Sao Paulo)":  "sa-east-1",
	"AWS GovCloud (US)":          "unknown",
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
		return awsInstancePricePerHour(instance.Location(), instance.InstanceType())
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
		return s3BucketPerGBMonth * bucket.TotalSizeGB()
	} else if bucket.CSP() == cloud.GCP {
		return gcpBucketPerGBMonth * bucket.TotalSizeGB()
	}
	log.Panicln("Unsupported CSP:", bucket.CSP())
	return 0.0
}

// awsInstancePricePerHour will return the hourly price in USD for a
// specified instance type in a specified AWS region. If the specified
// region/type pair does not exist, $0.0 will be returned.
func awsInstancePricePerHour(region, instanceType string) float64 {
	if awsPrices != nil {
		// Prices have already been fetched before
		price, exist := awsPrices[instanceKeyPair{region, instanceType}]
		if !exist {
			return 0.0
		}
		return price
	}
	log.Println("Fetching current instance prices from AWS")
	rawPrices := getRawAWSData()
	filteredProducts := filterRelevantAWSProducts(&rawPrices)
	awsPrices = make(priceMap)

	for _, product := range filteredProducts {
		for _, term := range rawPrices.Terms.OnDemand[product.SKU] {
			for _, price := range term.PriceDimensions {
				regionID, ok := awsRegionNameToIDMap[product.Region]
				if !ok {
					log.Fatalln("Got an unknown region from AWS")
				}
				key := instanceKeyPair{
					Region:       regionID,
					InstanceType: product.InstanceType,
				}
				usd, err := strconv.ParseFloat(price.PricePerUnit.USD, 64)
				if err != nil {
					log.Println("Could not convert price from AWS JSON", err)
				}
				awsPrices[key] = usd
				continue
			}
		}
	}
	price, exist := awsPrices[instanceKeyPair{region, instanceType}]
	if !exist {
		return 0.0
	}
	return price
}

// Helper structs for parsing the JSON from AWS
type rawAWSPricing struct {
	Products map[string]rawAWSProduct `json:"products"`
	Terms    struct {
		OnDemand map[string]rawAWSTerm `json:"OnDemand"`
	} `json:"terms"`
}
type rawAWSProduct struct {
	Sku        string `json:"sku"`
	Attributes struct {
		Location        string `json:"location"`
		LocationType    string `json:"locationType"`
		InstanceType    string `json:"instanceType"`
		Tenancy         string `json:"tenancy"`
		OperatingSystem string `json:"operatingSystem"`
	} `json:"attributes"`
}

type rawAWSTerm map[string]struct {
	OfferTermCode   string `json:"offerTermCode"`
	Sku             string `json:"sku"`
	PriceDimensions map[string]struct {
		PricePerUnit struct {
			USD string `json:"USD"`
		} `json:"pricePerUnit"`
	} `json:"priceDimensions"`
}

type awsSimpleProduct struct {
	SKU          string
	Region       string
	InstanceType string
}

type instanceKeyPair struct {
	Region, InstanceType string
}

type priceMap map[instanceKeyPair]float64

// Fetch raw pricing data from AWS and decode it
func getRawAWSData() rawAWSPricing {
	resp, err := http.Get(awsPricingURL)
	if err != nil {
		log.Fatalln("Could not download Pricing JSON", err)
	}
	decoder := json.NewDecoder(resp.Body)
	res := rawAWSPricing{}

	err = decoder.Decode(&res)
	if err != nil {
		log.Panicln("Could not decode JSON from AWS", err)
	}
	return res
}

// We're only interested in some of the prodcuts fetched from AWS
func filterRelevantAWSProducts(raw *rawAWSPricing) []awsSimpleProduct {
	filteredProducts := []awsSimpleProduct{}
	for sku, product := range raw.Products {
		attr := product.Attributes
		if attr.Tenancy == "Shared" && attr.LocationType == "AWS Region" && attr.OperatingSystem == "Linux" {
			simple := awsSimpleProduct{
				SKU:          sku,
				Region:       attr.Location,
				InstanceType: attr.InstanceType,
			}
			filteredProducts = append(filteredProducts, simple)
		}
	}
	return filteredProducts
}
