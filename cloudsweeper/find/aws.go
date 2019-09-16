// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package find

import (
	"fmt"
	"log"
	"strings"

	"github.com/agaridata/cloudsweeper/cloudsweeper"

	"github.com/agaridata/cloudsweeper/cloud"
)

type awsResourceType int

const (
	awsTypeInstance awsResourceType = iota
	awsTypeVolume
	awsTypeSnapshop
	awsTypeImage
)

type awsClient struct {
	cloudManager cloud.ResourceManager
	organization *cloudsweeper.Organization
}

func (c *awsClient) CSP() cloud.CSP {
	return cloud.AWS
}

func (c *awsClient) FindResource(id string) error {
	resourceType, err := c.determineResourceType(id)
	if err != nil {
		return err
	}

	for account, resources := range c.cloudManager.AllResourcesPerAccount() {
		log.Printf("Looking for %s in account %s\n", id, account)
		switch resourceType {
		case awsTypeInstance:
			for _, inst := range resources.Instances {
				if inst.ID() == id {
					// Found instance
					log.Printf("Found instance in account %s", account)
					employee, err := c.getEmployee(account)
					if err != nil {
						return err
					}
					foundInstance(inst, account, employee)
					return nil
				}
			}
		case awsTypeVolume:
			for _, vol := range resources.Volumes {
				if vol.ID() == id {
					// Found volume
					employee, err := c.getEmployee(account)
					if err != nil {
						return err
					}
					foundVolume(vol, account, employee)
					return nil
				}
			}
		case awsTypeImage:
			for _, ami := range resources.Images {
				if ami.ID() == id {
					// Found AMI
					employee, err := c.getEmployee(account)
					if err != nil {
						return err
					}
					foundImage(ami, account, employee)
					return nil
				}
			}
		case awsTypeSnapshop:
			for _, snap := range resources.Snapshots {
				if snap.ID() == id {
					// Found snapshot
					employee, err := c.getEmployee(account)
					if err != nil {
						return err
					}
					foundSnapshot(snap, account, employee)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("Resource %s not found in any account", id)
}

func (c *awsClient) determineResourceType(id string) (awsResourceType, error) {
	idParts := strings.Split(id, "-")
	if len(idParts) != 2 {
		return -1, fmt.Errorf("Looks like ID %s is not a valid AWS resource id", id)
	}
	prefix := idParts[0]
	switch prefix {
	case "i":
		log.Println("Resource is an instance")
		return awsTypeInstance, nil
	case "vol":
		log.Println("Resource is a volume")
		return awsTypeVolume, nil
	case "ami":
		log.Println("Resource is an image/AMI")
		return awsTypeImage, nil
	case "snap":
		log.Println("Resource is a snapshot")
		return awsTypeSnapshop, nil
	default:
		return -1, fmt.Errorf("Unsupported resource type, must be one of either instance, volume, AMI, or snapshot")
	}
}

func (c *awsClient) getEmployee(accountID string) (*cloudsweeper.Employee, error) {
	users := c.organization.AccountToUserMapping(cloud.AWS)
	employees := c.organization.UsernameToEmployeeMapping()
	if user, ok := users[accountID]; ok {
		if employee, ok := employees[user]; ok {
			return employee, nil
		}
	}
	return nil, fmt.Errorf("Could not find information about account %s", accountID)
}
