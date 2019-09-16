// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

// Package find is containing functionality to find more information
// about a cloud resource given its ID.
package find

import (
	"fmt"
	"time"

	"github.com/agaridata/cloudsweeper/cloud"
	"github.com/agaridata/cloudsweeper/cloudsweeper"
)

const foundBannerTemplate = `

#############################################
               Found %s
#############################################

`

// Client is a client for finding a resource in a specific cloud
type Client interface {
	FindResource(id string) error
	CSP() cloud.CSP
}

// Init will initialize a finding Client for the given CSP
func Init(mngr cloud.ResourceManager, org *cloudsweeper.Organization, csp cloud.CSP) (Client, error) {
	if csp == cloud.AWS {
		return &awsClient{
			cloudManager: mngr,
			organization: org,
		}, nil
	}
	return nil, fmt.Errorf("Unsupported CSP: %s", csp)
}

func foundInstance(inst cloud.Instance, account string, owner *cloudsweeper.Employee) {
	fmt.Printf(foundBannerTemplate, "Instance")
	foundResource(inst, account, owner)
	fmt.Printf("Instance Type: %s\n", inst.InstanceType())
}

func foundVolume(vol cloud.Volume, account string, owner *cloudsweeper.Employee) {
	fmt.Printf(foundBannerTemplate, "Volume")
	foundResource(vol, account, owner)
	fmt.Printf("Volume Type:   %s\n", vol.VolumeType())
	fmt.Printf("Size:          %d GB\n", vol.SizeGB())
}

func foundImage(image cloud.Image, account string, owner *cloudsweeper.Employee) {
	fmt.Printf(foundBannerTemplate, "Image")
	foundResource(image, account, owner)
	var isPublic string
	if image.Public() {
		isPublic = "Yes"
	} else {
		isPublic = "No"
	}
	fmt.Printf("Is public:     %s\n", isPublic)
	fmt.Printf("Size:          %d GB\n", image.SizeGB())
}

func foundSnapshot(snap cloud.Snapshot, account string, owner *cloudsweeper.Employee) {
	fmt.Printf(foundBannerTemplate, "Snapshot")
	foundResource(snap, account, owner)
	fmt.Printf("Size:          %d GB\n", snap.SizeGB())
}

func foundResource(res cloud.Resource, account string, owner *cloudsweeper.Employee) {
	var resourceName = "<no name tag>"
	if name, ok := res.Tags()["Name"]; ok {
		resourceName = name
	}

	fmt.Printf("Account:       %s (%s)\n", owner.Username, account)
	fmt.Printf("Resource ID:   %s\n", res.ID())
	fmt.Printf("Resource name: %s\n", resourceName)
	fmt.Printf("Region:        %s\n", res.Location())
	fmt.Printf("Creation Time: %s\n", res.CreationTime().Format(time.RFC3339))
	fmt.Printf("Tags:\n")
	for key, val := range res.Tags() {
		if val != "" {
			fmt.Printf("\t\t%s: %s\n", key, val)
		} else {
			fmt.Printf("\t\t%s\n", key)
		}
	}
}
