package res

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// awsResourceManager uses the AWS Go SDK. Docs can be found at:
// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/
type awsResourceManager struct {
	accounts []string
}

type awsInstance struct {
	baseInstance
}

type awsImage struct {
	baseImage
}

type awsVolume struct {
	baseVolume
}

type awsSnapshot struct {
	baseSnapshot
}

const (
	assumeRoleARNTemplate = "arn:aws:iam::%s:role/brkt-HouseKeeper"

	accessDeniedErrorCode = "AccessDenied"
)

var (
	instanceStateFilterName = "instance-state-name"
	instanceStateRunning    = ec2.InstanceStateNameRunning

	awsOwnerIDSelfValue = "self"
)

func (m *awsResourceManager) InstancesPerAccount() map[string][]Instance {
	resultMap := make(map[string][]Instance)
	getAllResources(m.accounts, func(client *ec2.EC2, account string) {
		instances, err := getAWSInstances(client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(instances) > 0 {
			resultMap[account] = append(resultMap[account], instances...)
		}
	})
	return resultMap
}

func (m *awsResourceManager) ImagesPerAccount() map[string][]Image {
	resultMap := make(map[string][]Image)
	getAllResources(m.accounts, func(client *ec2.EC2, account string) {
		images, err := getAWSImages(client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(images) > 0 {
			resultMap[account] = append(resultMap[account], images...)
		}
	})
	return resultMap
}

func (m *awsResourceManager) VolumesPerAccount() map[string][]Volume {
	resultMap := make(map[string][]Volume)
	getAllResources(m.accounts, func(client *ec2.EC2, account string) {
		volumes, err := getAWSVolumes(client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(volumes) > 0 {
			resultMap[account] = append(resultMap[account], volumes...)
		}
	})
	return resultMap
}

func (m *awsResourceManager) SnapshotsPerAccount() map[string][]Snapshot {
	resultMap := make(map[string][]Snapshot)
	getAllResources(m.accounts, func(client *ec2.EC2, account string) {
		snapshots, err := getAWSSnapshots(client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(snapshots) > 0 {
			resultMap[account] = append(resultMap[account], snapshots...)
		}
	})
	return resultMap
}

// getAWSInstances will get all running instances using an already
// set-up client for a specific credential and region.
func getAWSInstances(client *ec2.EC2) ([]Instance, error) {
	// We're only interested in running instances
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{&ec2.Filter{
			Name:   aws.String(instanceStateFilterName),
			Values: aws.StringSlice([]string{instanceStateRunning})}},
	}
	awsReservations, err := client.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	result := []Instance{}
	for _, reservation := range awsReservations.Reservations {
		for _, instance := range reservation.Instances {
			inst := awsInstance{baseInstance{
				id:           *instance.InstanceId,
				location:     *client.Config.Region,
				launchTime:   *instance.LaunchTime,
				public:       instance.PublicIpAddress != nil,
				tags:         convertAWSTags(instance.Tags),
				instanceType: *instance.InstanceType,
			}}
			result = append(result, &inst)
		}
	}
	return result, nil
}

// getAWSImages will get all AMIs owned by the current account
func getAWSImages(client *ec2.EC2) ([]Image, error) {
	input := &ec2.DescribeImagesInput{
		Owners: aws.StringSlice([]string{awsOwnerIDSelfValue}),
	}
	awsImages, err := client.DescribeImages(input)
	if err != nil {
		return nil, err
	}
	result := []Image{}
	for _, ami := range awsImages.Images {
		ti, err := time.Parse(time.RFC3339, *ami.CreationDate)
		if err != nil {
			return nil, err
		}
		img := awsImage{baseImage{
			id:           *ami.ImageId,
			location:     *client.Config.Region,
			creationTime: ti,
			public:       *ami.Public,
			tags:         convertAWSTags(ami.Tags),
			name:         *ami.Name,
		}}
		result = append(result, &img)
	}
	return result, nil
}

// getAWSVolumes will get all volumes (both attached and un-attached)
// in the current account
func getAWSVolumes(client *ec2.EC2) ([]Volume, error) {
	input := new(ec2.DescribeVolumesInput)
	awsVolumes, err := client.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	result := []Volume{}
	for _, volume := range awsVolumes.Volumes {
		vol := awsVolume{baseVolume{
			id:           *volume.VolumeId,
			location:     *client.Config.Region,
			creationTime: *volume.CreateTime,
			public:       false,
			tags:         convertAWSTags(volume.Tags),
			sizeGB:       *volume.Size,
			attached:     len(volume.Attachments) > 0,
			encrypted:    *volume.Encrypted,
			volumeType:   *volume.VolumeType,
		}}
		result = append(result, &vol)
	}
	return result, nil
}

// getAWSSnapshots will get all snapshots in AWS owned
// by the current account
func getAWSSnapshots(client *ec2.EC2) ([]Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: aws.StringSlice([]string{awsOwnerIDSelfValue}),
	}
	awsSnapshots, err := client.DescribeSnapshots(input)
	if err != nil {
		return nil, err
	}
	result := []Snapshot{}
	for _, snapshot := range awsSnapshots.Snapshots {
		snap := awsSnapshot{baseSnapshot{
			id:           *snapshot.SnapshotId,
			location:     *client.Config.Region,
			creationTime: *snapshot.StartTime,
			public:       false,
			tags:         convertAWSTags(snapshot.Tags),
			sizeGB:       *snapshot.VolumeSize,
			encrypted:    *snapshot.Encrypted,
		}}
		result = append(result, &snap)
	}
	return result, nil
}

func getAllResources(accounts []string, funcToRun func(client *ec2.EC2, account string)) {
	sess := session.Must(session.NewSession())
	forEachAccount(accounts, sess, func(account string, cred *credentials.Credentials) {
		log.Println("Accessing account", account)
		forEachAWSRegion(func(region string) {
			client := ec2.New(sess, &aws.Config{
				Credentials: cred,
				Region:      aws.String(region),
			})
			funcToRun(client, account)
		})
	})
}

// forEachAccount is a higher order function that will, for
// every account, create credentials and call the specified
// function with those creds
func forEachAccount(accounts []string, sess *session.Session, funcToRun func(account string, cred *credentials.Credentials)) {
	var wg sync.WaitGroup
	for i := range accounts {
		wg.Add(1)
		go func(x int) {
			creds := stscreds.NewCredentials(sess, fmt.Sprintf(assumeRoleARNTemplate, accounts[x]))
			funcToRun(accounts[x], creds)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// forEachAWSRegion is a higher order function that will, for
// every available AWS region, run the specified function
func forEachAWSRegion(funcToRun func(region string)) {
	regions, exists := endpoints.RegionsForService(endpoints.DefaultPartitions(), endpoints.AwsPartitionID, endpoints.Ec2ServiceID)
	if !exists {
		panic("The regions for EC2 in the standard partition should exist")
	}
	var wg sync.WaitGroup
	for regionID := range regions {
		wg.Add(1)
		go func(x string) {
			funcToRun(x)
			wg.Done()
		}(regionID)
	}
	wg.Wait()
}

func handleAWSAccessDenied(account string, err error) {
	// Cast err to awserr.Error to handle specific AWS errors
	aerr, ok := err.(awserr.Error)
	if ok && aerr.Code() == accessDeniedErrorCode {
		// The account does not have the role setup correctly
		log.Printf("The account '%s' denied access\n", account)
	} else if ok {
		// Some other AWS error occured
		log.Fatalln(aerr)
	} else {
		//Some other non-AWS error occured
		log.Fatalln(err)
	}
}

func convertAWSTags(tags []*ec2.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[*tag.Key] = *tag.Value
	}
	return result
}
