// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	defaultAWSRegion = "us-west-2"
	gbDivider        = 1024.0 * 1024.0 * 1024.0
	awsStateInUse    = "in-use"
)

// awsResourceManager uses the AWS Go SDK. Docs can be found at:
// https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/
type awsResourceManager struct {
	accounts []string
}

func (m *awsResourceManager) Owners() []string {
	return m.accounts
}

const (
	assumeRoleARNTemplate = "arn:aws:iam::%s:role/Cloudsweeper"

	accessDeniedErrorCode = "AccessDenied"
	unauthorizedErrorCode = "UnauthorizedOperation"
	notFoundErrorOcde     = "NotFound"
	requestLimitErrorCode = "RequestLimitExceeded"

	snapshotIDFilterName = "block-device-mapping.snapshot-id"

	awsMaxRequestRetries = 6
)

var (
	instanceStateFilterName = "instance-state-name"
	instanceStateRunning    = ec2.InstanceStateNameRunning

	awsOwnerIDSelfValue = "self"

	errAWSRequestLimit = errors.New("aws request limit hit")
)

var awsS3StorageTypes = []string{
	"StandardStorage",
	"IntelligentTieringFAStorage",
	"IntelligentTieringIAStorage",
	"StandardIAStorage",
	"OneZoneIAStorage",
	"ReducedRedundancyStorage",
	"GlacierStorage",
}

func (m *awsResourceManager) InstancesPerAccount() map[string][]Instance {
	log.Println("Getting instances in all accounts")
	resultMap := make(map[string][]Instance)
	var resultMutext sync.Mutex
	getAllEC2Resources(m.accounts, func(client *ec2.EC2, account string) {
		instances, err := getAWSInstances(account, client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(instances) > 0 {
			resultMutext.Lock()
			resultMap[account] = append(resultMap[account], instances...)
			resultMutext.Unlock()
		}
	})
	return resultMap
}

func (m *awsResourceManager) ImagesPerAccount() map[string][]Image {
	log.Println("Getting images in all accounts")
	resultMap := make(map[string][]Image)
	var resultMutext sync.Mutex
	getAllEC2Resources(m.accounts, func(client *ec2.EC2, account string) {
		images, err := getAWSImages(account, client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(images) > 0 {
			resultMutext.Lock()
			resultMap[account] = append(resultMap[account], images...)
			resultMutext.Unlock()
		}
	})
	return resultMap
}

func (m *awsResourceManager) VolumesPerAccount() map[string][]Volume {
	log.Println("Getting volumes in all accounts")
	resultMap := make(map[string][]Volume)
	var resultMutext sync.Mutex
	getAllEC2Resources(m.accounts, func(client *ec2.EC2, account string) {
		volumes, err := getAWSVolumes(account, client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(volumes) > 0 {
			resultMutext.Lock()
			resultMap[account] = append(resultMap[account], volumes...)
			resultMutext.Unlock()
		}
	})
	return resultMap
}

func (m *awsResourceManager) SnapshotsPerAccount() map[string][]Snapshot {
	log.Println("Getting snapshots in all accounts")
	resultMap := make(map[string][]Snapshot)
	var resultMutext sync.Mutex
	getAllEC2Resources(m.accounts, func(client *ec2.EC2, account string) {
		snapshots, err := getAWSSnapshots(account, client)
		if err != nil {
			handleAWSAccessDenied(account, err)
		} else if len(snapshots) > 0 {
			resultMutext.Lock()
			resultMap[account] = append(resultMap[account], snapshots...)
			resultMutext.Unlock()
		}
	})
	return resultMap
}

func (m *awsResourceManager) AllResourcesPerAccount() map[string]*ResourceCollection {
	log.Println("Getting all resources in all accounts")
	resultMap := make(map[string]*ResourceCollection)
	var resultMutext sync.Mutex
	for i := range m.accounts {
		resultMap[m.accounts[i]] = new(ResourceCollection)
	}
	// TODO: Smarter error handling. If one request get access denied, then might as
	// well abort. The rest are going to fail too.
	getAllEC2Resources(m.accounts, func(client *ec2.EC2, account string) {
		result := resultMap[account]
		result.Owner = account
		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			snapshots, err := getAWSSnapshots(account, client)
			if err != nil {
				log.Printf("Snapshot error when getting all resources in %s", account)
				handleAWSAccessDenied(account, err)
			}
			result.Snapshots = append(result.Snapshots, snapshots...)
			wg.Done()
		}()
		go func() {
			instances, err := getAWSInstances(account, client)
			if err != nil {
				log.Printf("Instance error when getting all resources in %s", account)
				handleAWSAccessDenied(account, err)
			}
			result.Instances = append(result.Instances, instances...)
			wg.Done()
		}()
		go func() {
			images, err := getAWSImages(account, client)
			if err != nil {
				log.Printf("Image error when getting all resources in %s", account)
				handleAWSAccessDenied(account, err)
			}
			result.Images = append(result.Images, images...)
			wg.Done()
		}()
		go func() {
			volumes, err := getAWSVolumes(account, client)
			if err != nil {
				log.Printf("Volume error when getting all resources in %s", account)
				handleAWSAccessDenied(account, err)
			}
			result.Volumes = append(result.Volumes, volumes...)
			wg.Done()
		}()
		wg.Wait()
		resultMutext.Lock()
		resultMap[account] = result
		resultMutext.Unlock()
	})
	return resultMap
}

func (m *awsResourceManager) BucketsPerAccount() map[string][]Bucket {
	log.Println("Getting all buckets in all accounts")
	sess := session.Must(session.NewSession())
	resultMap := make(map[string][]Bucket)
	var resultMutext sync.Mutex
	forEachAccount(m.accounts, sess, func(account string, cred *credentials.Credentials) {
		s3Client := s3.New(sess, &aws.Config{
			Credentials: cred,
			Region:      aws.String(defaultAWSRegion),
		})
		awsBuckets, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
		if err != nil {
			log.Printf("Bucket error when getting buckets in %s", account)
			handleAWSAccessDenied(account, err)
		} else if len(awsBuckets.Buckets) > 0 {
			bucketCount := len(awsBuckets.Buckets)
			buckChan := make(chan *awsBucket)
			for _, bu := range awsBuckets.Buckets {
				go func(bu *s3.Bucket, resChan chan *awsBucket) {
					region, err := s3manager.GetBucketRegion(context.Background(), sess, *bu.Name, defaultAWSRegion)
					if err != nil {
						bucketCount--
						log.Printf("Couldn't determine bucket region in %s for bucket %s", account, *bu.Name)
						handleAWSAccessDenied(account, err)
						buckChan <- nil
						return
					}
					bucketClient := s3.New(sess, &aws.Config{
						Credentials: cred,
						Region:      aws.String(region),
					})
					buTags, err := bucketClient.GetBucketTagging(&s3.GetBucketTaggingInput{
						Bucket: bu.Name,
					})
					tags := make(map[string]string)
					if err == nil {
						tags = convertAWSS3Tags(buTags.TagSet)
					}

					cw := cloudwatch.New(sess, &aws.Config{
						Credentials: cred,
						Region:      aws.String(region)})
					storageTypeSizesGB := make(map[string]float64)
					numberOfObjects := int64(0)

					var input cloudwatch.GetMetricStatisticsInput
					input.Namespace = aws.String("AWS/S3")
					input.MetricName = aws.String("BucketSizeBytes")
					input.StartTime = aws.Time(time.Now().Add(time.Duration(-48*60) * time.Minute))
					input.EndTime = aws.Time(time.Now())
					input.Period = aws.Int64(24 * 60 * 60)
					input.Statistics = []*string{aws.String("Average")}
					input.Unit = aws.String("Bytes")
					dimensionNameFilter := cloudwatch.Dimension{
						Name:  aws.String("BucketName"),
						Value: bu.Name,
					}

					// Get sizes for all storage types
					numBucketSizeDatapoints := 0
					for _, storageType := range awsS3StorageTypes {
						dimensionBucketSizeFilter := cloudwatch.Dimension{
							Name:  aws.String("StorageType"),
							Value: aws.String(storageType),
						}
						input.Dimensions = []*cloudwatch.Dimension{
							&dimensionNameFilter, &dimensionBucketSizeFilter,
						}
						bucketSizeMetrics, err := cw.GetMetricStatistics(&input)
						if err != nil {
							fmt.Println("Error", err)
						}
						if bucketSizeMetrics != nil {
							var minimumTimeDifference float64
							var timeDifference float64
							var averageValue *float64
							minimumTimeDifference = -1
							for _, datapoint := range bucketSizeMetrics.Datapoints {
								timeDifference = time.Since(*datapoint.Timestamp).Seconds()
								if minimumTimeDifference == -1 {
									minimumTimeDifference = timeDifference
									averageValue = datapoint.Average
								} else if timeDifference < minimumTimeDifference {
									minimumTimeDifference = timeDifference
									averageValue = datapoint.Average
								}
							}
							if averageValue != nil {
								storageTypeSizesGB[storageType] = float64(*averageValue) / gbDivider
							}
							numBucketSizeDatapoints += len(bucketSizeMetrics.Datapoints)
						}
					}

					// Update input to get numberOfObjects instead
					input.MetricName = aws.String("NumberOfObjects")
					dimensionNumberOfObjectsFilter := cloudwatch.Dimension{
						Name:  aws.String("StorageType"),
						Value: aws.String("AllStorageTypes"),
					}
					input.Dimensions = []*cloudwatch.Dimension{
						&dimensionNameFilter, &dimensionNumberOfObjectsFilter,
					}
					input.Unit = aws.String("Count")
					numberOfObjectsMetrics, err := cw.GetMetricStatistics(&input)
					if err != nil {
						fmt.Println("Error", err)
					}
					if numBucketSizeDatapoints == 0 && len(numberOfObjectsMetrics.Datapoints) != 0 {
						fmt.Println("Warning: Got 0 datapoints from: ", *bu.Name)
					}
					if numberOfObjectsMetrics != nil {
						var minimumTimeDifference float64
						var timeDifference float64
						var averageValue *float64
						minimumTimeDifference = -1
						for _, datapoint := range numberOfObjectsMetrics.Datapoints {
							timeDifference = time.Since(*datapoint.Timestamp).Seconds()
							if minimumTimeDifference == -1 {
								minimumTimeDifference = timeDifference
								averageValue = datapoint.Average
							} else if timeDifference < minimumTimeDifference {
								minimumTimeDifference = timeDifference
								averageValue = datapoint.Average
							}
						}
						if averageValue != nil {
							numberOfObjects = int64(*averageValue)
						}
					}

					// TODO: this should be configurable instead of hardcoded to 6 + 1 months
					lastMod := time.Now().AddDate(0, -7, 0)
					err = bucketClient.ListObjectsV2Pages(&s3.ListObjectsV2Input{
						Bucket: bu.Name, EncodingType: aws.String("url"),
					}, func(output *s3.ListObjectsV2Output, lastPage bool) bool {
						for _, object := range output.Contents {
							// if object has been modified in the last 6 months
							if time.Now().Before(object.LastModified.AddDate(0, 6, 0)) {
								lastMod = time.Now().AddDate(0, -5, 0)
								// exit early
								return false
							}
						}
						return !lastPage
					})
					if err != nil {
						bucketCount--
						log.Printf("Failed to list contents in bucket %s, account %s", *bu.Name, account)
						handleAWSAccessDenied(account, err)
						buckChan <- nil
						return
					}

					totalSizeGB := 0.0
					for _, size := range storageTypeSizesGB {
						totalSizeGB += size
					}

					buck := awsBucket{baseBucket{
						baseResource: baseResource{
							csp:          AWS,
							owner:        account,
							location:     region,
							id:           *bu.Name,
							creationTime: *bu.CreationDate,
							tags:         tags,
						},
						lastModified:       lastMod,
						objectCount:        numberOfObjects,
						totalSizeGB:        totalSizeGB,
						storageTypeSizesGB: storageTypeSizesGB,
					}}
					buckChan <- &buck
				}(bu, buckChan)
			}
			for i := 0; i < bucketCount; i++ {
				buck := <-buckChan
				if buck != nil {
					resultMutext.Lock()
					resultMap[account] = append(resultMap[account], buck)
					resultMutext.Unlock()
				}
			}
		}
	})
	return resultMap
}

func (m *awsResourceManager) CleanupInstances(instances []Instance) error {
	return cleanupInstances(instances)
}

func (m *awsResourceManager) CleanupImages(images []Image) error {
	return cleanupImages(images)
}

func (m *awsResourceManager) CleanupVolumes(volumes []Volume) error {
	return cleanupVolumes(volumes)
}

func (m *awsResourceManager) CleanupSnapshots(snapshots []Snapshot) error {
	return cleanupSnapshots(snapshots)
}

func (m *awsResourceManager) CleanupBuckets(buckets []Bucket) error {
	return cleanupBuckets(buckets)
}

// getAWSInstances will get all running instances using an already
// set-up client for a specific credential and region.
func getAWSInstances(account string, client *ec2.EC2) ([]Instance, error) {
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
				baseResource: baseResource{
					csp:          AWS,
					owner:        account,
					id:           *instance.InstanceId,
					location:     *client.Config.Region,
					creationTime: *instance.LaunchTime,
					public:       instance.PublicIpAddress != nil,
					tags:         convertAWSTags(instance.Tags)},
				instanceType: *instance.InstanceType,
			}}
			result = append(result, &inst)
		}
	}
	return result, nil
}

// getAWSImages will get all AMIs owned by the current account
func getAWSImages(account string, client *ec2.EC2) ([]Image, error) {
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
			baseResource: baseResource{
				csp:          AWS,
				owner:        account,
				id:           *ami.ImageId,
				location:     *client.Config.Region,
				creationTime: ti,
				public:       *ami.Public,
				tags:         convertAWSTags(ami.Tags),
			},
			name: *ami.Name,
		}}
		for _, mapping := range ami.BlockDeviceMappings {
			if mapping != nil && (*mapping).Ebs != nil && (*(*mapping).Ebs).VolumeSize != nil {
				img.baseImage.sizeGB += *mapping.Ebs.VolumeSize
			}
		}
		result = append(result, &img)
	}
	return result, nil
}

// getAWSVolumes will get all volumes (both attached and un-attached)
// in the current account
func getAWSVolumes(account string, client *ec2.EC2) ([]Volume, error) {
	input := new(ec2.DescribeVolumesInput)
	awsVolumes, err := client.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	result := []Volume{}
	for _, volume := range awsVolumes.Volumes {
		inUse := len(volume.Attachments) > 0 || *volume.State == awsStateInUse
		vol := awsVolume{baseVolume{
			baseResource: baseResource{
				csp:          AWS,
				owner:        account,
				id:           *volume.VolumeId,
				location:     *client.Config.Region,
				creationTime: *volume.CreateTime,
				public:       false,
				tags:         convertAWSTags(volume.Tags),
			},
			sizeGB:     *volume.Size,
			attached:   inUse,
			encrypted:  *volume.Encrypted,
			volumeType: *volume.VolumeType,
		}}
		result = append(result, &vol)
	}
	return result, nil
}

// getAWSSnapshots will get all snapshots in AWS owned
// by the current account
func getAWSSnapshots(account string, client *ec2.EC2) ([]Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: aws.StringSlice([]string{awsOwnerIDSelfValue}),
	}
	awsSnapshots, err := client.DescribeSnapshots(input)
	if err != nil {
		return nil, err
	}
	result := []Snapshot{}
	snapshotsInUse := getSnapshotsInUse(client)
	for _, snapshot := range awsSnapshots.Snapshots {
		_, inUse := snapshotsInUse[*snapshot.SnapshotId]
		snap := awsSnapshot{baseSnapshot{
			baseResource: baseResource{
				csp:          AWS,
				owner:        account,
				id:           *snapshot.SnapshotId,
				location:     *client.Config.Region,
				creationTime: *snapshot.StartTime,
				public:       false,
				tags:         convertAWSTags(snapshot.Tags),
			},
			sizeGB:    *snapshot.VolumeSize,
			encrypted: *snapshot.Encrypted,
			inUse:     inUse,
		}}
		result = append(result, &snap)
	}
	return result, nil
}

func getSnapshotsInUse(client *ec2.EC2) map[string]struct{} {
	result := make(map[string]struct{})
	input := &ec2.DescribeImagesInput{
		Owners: aws.StringSlice([]string{awsOwnerIDSelfValue}),
	}
	images, err := client.DescribeImages(input)
	if err != nil {
		log.Printf("Could not determine snapshots in use:\n%s\n", err)
		return result
	}
	for _, imgs := range images.Images {
		for _, mapping := range imgs.BlockDeviceMappings {
			if mapping != nil && mapping.Ebs != nil && mapping.Ebs.SnapshotId != nil {
				result[*mapping.Ebs.SnapshotId] = struct{}{}
			}
		}
	}
	return result
}

func getAllEC2Resources(accounts []string, funcToRun func(client *ec2.EC2, account string)) {
	sess := session.Must(session.NewSession())
	forEachAccount(accounts, sess, func(account string, cred *credentials.Credentials) {
		log.Println("Accessing account", account)
		forEachAWSRegion(func(region string) {
			// Check if region is enabled by making a call that we should always have permissions for
			stsClient := sts.New(sess, &aws.Config{
				Credentials: cred,
				Region:      aws.String(region),
			})
			_, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if err != nil {
				// Ensure that we can make the default call, otherwise we have other problems
				stsClient = sts.New(sess, &aws.Config{
					Credentials: cred,
				})
				_, err = stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
				if err == nil {
					log.Printf("Region %s is disabled, skipping it!", region)
					return
				} else {
					log.Fatalf("Unknown AWS error %s", err)
				}
			}
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
	} else if ok && aerr.Code() == unauthorizedErrorCode {
		log.Printf("Unauthorized to assume '%s'\n", account)
	} else if ok && aerr.Code() == notFoundErrorOcde {
		log.Printf("Resource was not found in account %s", account)
	} else if ok {
		// Some other AWS error occured
		log.Fatalf("Got AWS error in account %s: %s", account, aerr)
	} else {
		//Some other non-AWS error occured
		log.Fatalf("Got error in account %s: %s", account, err)
	}
}

func convertAWSTags(tags []*ec2.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[*tag.Key] = *tag.Value
	}
	return result
}

func convertAWSS3Tags(tags []*s3.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[*tag.Key] = *tag.Value
	}
	return result
}

func clientForAWSResource(res Resource) *ec2.EC2 {
	sess := session.Must(session.NewSession())
	creds := stscreds.NewCredentials(sess, fmt.Sprintf(assumeRoleARNTemplate, res.Owner()))
	return ec2.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String(res.Location()),
	})
}

func addAWSTag(r Resource, key, value string, overwrite bool) error {
	_, exist := r.Tags()[key]
	if exist && !overwrite {
		return fmt.Errorf("Key %s already exist on %s", key, r.ID())
	}
	client := clientForAWSResource(r)
	input := &ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{r.ID()}),
		Tags: []*ec2.Tag{&ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}},
	}
	_, err := client.CreateTags(input)
	return err
}

func removeAWSTag(r Resource, key string) error {
	val, exist := r.Tags()[key]
	if !exist {
		return nil
	}
	client := clientForAWSResource(r)
	input := &ec2.DeleteTagsInput{
		Resources: aws.StringSlice([]string{r.ID()}),
		Tags: []*ec2.Tag{&ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		}},
	}
	_, err := client.DeleteTags(input)
	return err
}

func awsTryWithBackoff(f func() error) error {
	try := 1
	var err error
	for {
		err = f()
		if err == nil || err != errAWSRequestLimit || try > awsMaxRequestRetries {
			break
		}
		// Stupid but simple backoff (2^try seconds): 2, 4, 8, 16, 32 etc... seconds
		time.Sleep(time.Duration(math.Exp2(float64(try))) * time.Second)
		try++
	}
	return err
}
