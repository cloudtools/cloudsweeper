package billing

import (
	"brkt/housekeeper/cloud"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

const (
	awsPricingURL = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json"
)

var (
	awsPrices priceMap
)

var awsRegionNameToIDMap = map[string]string{
	"US East (Ohio)":            "us-east-2",
	"US East (N. Virginia)":     "us-east-1",
	"US West (N. California)":   "us-west-1",
	"US West (Oregon)":          "us-west-2",
	"Asia Pacific (Tokyo)":      "ap-northeast-1",
	"Asia Pacific (Seoul)":      "ap-northeast-2",
	"Asia Pacific (Mumbai)":     "ap-south-1",
	"Asia Pacific (Singapore)":  "ap-southeast-1",
	"Asia Pacific (Sydney)":     "ap-southeast-2",
	"Canada (Central)":          "ca-central-1",
	"China (Beijing)":           "cn-north-1",
	"China (Ningxia)":           "cn-northwest-1",
	"EU (Frankfurt)":            "eu-central-1",
	"EU (Ireland)":              "eu-west-1",
	"EU (London)":               "eu-west-2",
	"EU (Paris)":                "eu-west-3",
	"South America (Sao Paulo)": "sa-east-1",
	"AWS GovCloud (US)":         "unknown",
}

// Storage cost per GB per day
var awsStorageCostMap = map[string]float64{
	"gp2":      0.1 / 30.0,
	"io1":      0.125 / 30.0,
	"st1":      0.045 / 30.0,
	"sc1":      0.025 / 30.0,
	"snapshot": 0.05 / 30.0,
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
	}
	log.Panicln("Unsupported CSP:", snapshot.CSP())
	return 0.0
}

// ImageCostPerDay returns the daily cost in USD for a
// certain image
func ImageCostPerDay(image cloud.Image) float64 {
	if image.CSP() == cloud.AWS {
		return awsStorageCostMap["snapshot"] * float64(image.SizeGB())
	}
	log.Panicln("Unsupported CSP:", image.CSP())
	return 0.0
}

// InstancePricePerHour will return the hourly price in USD for a
// specified instance.
func InstancePricePerHour(instance cloud.Instance) float64 {
	if instance.CSP() == cloud.AWS {
		return awsInstancePricePerHour(instance.Location(), instance.InstanceType())
	}
	log.Panicln("Unsupported CSP:", instance.CSP())
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
