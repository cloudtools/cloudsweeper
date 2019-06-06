// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package cloud

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	compute "google.golang.org/api/compute/v1"
	storage "google.golang.org/api/storage/v1"
)

// Google Cloud API error codes can be found here:
// https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto

var (
	// ErrPermissionDenied is returned if not enough permissions to perform action
	ErrPermissionDenied = errors.New("permission denied")
)

// gcpResourceManager uses the Go API client for Google Cloud
// https://github.com/google/google-api-go-client
type gcpResourceManager struct {
	projects []string
	compute  *compute.Service
	storage  *storage.Service
}

func (m *gcpResourceManager) Owners() []string {
	return m.projects
}

func (m *gcpResourceManager) InstancesPerAccount() map[string][]Instance {
	log.Println("Getting instances in all projects")
	result := make(map[string][]Instance)
	var resultMutex sync.Mutex // Projects are processed in parallel
	m.forEachProject(func(project string) {
		instList := []Instance{}
		var listMutex sync.Mutex // Zones are proccessed in parallel
		m.forEachZone(project, func(zone string) {
			inst, err := m.getInstances(project, zone)
			if err != nil {
				log.Printf("Could not list instances in (%s, %s): %s", project, zone, err)
				if err == ErrPermissionDenied {
					log.Println(err)
				} else {
					// If it was an unknown error, abort
					log.Fatalln(err)
				}
			} else if len(inst) > 0 {
				listMutex.Lock()
				instList = append(instList, inst...)
				listMutex.Unlock()
			}
		})
		resultMutex.Lock()
		result[project] = instList
		resultMutex.Unlock()
	})
	return result
}

func (m *gcpResourceManager) ImagesPerAccount() map[string][]Image {
	log.Println("Getting images in all projects")
	result := make(map[string][]Image)
	var resultMutex sync.Mutex // Projects are processed in parallel
	m.forEachProject(func(project string) {
		images, err := m.getImages(project)
		if err != nil {
			log.Printf("Could not list images in %s: %s", project, err)
			if err == ErrPermissionDenied {
				log.Println(err)
			} else {
				// If it was an unknown error, abort
				log.Fatalln(err)
			}
		} else if len(images) > 0 {
			resultMutex.Lock()
			result[project] = images
			resultMutex.Unlock()
		}
	})
	return result
}

func (m *gcpResourceManager) VolumesPerAccount() map[string][]Volume {
	log.Println("Getting volumes in all projects")
	result := make(map[string][]Volume)
	var resultMutex sync.Mutex // Projects are processed in parallel
	m.forEachProject(func(project string) {
		diskList := []Volume{}
		var listMutex sync.Mutex // Zones are proccessed in parallel
		m.forEachZone(project, func(zone string) {
			volumes, err := m.getVolumes(project, zone)
			if err != nil {
				log.Printf("Could not list disks in (%s, %s): %s", project, zone, err)
				if err == ErrPermissionDenied {
					log.Println(err)
				} else {
					// If it was an unknown error, abort
					log.Fatalln(err)
				}
			} else if len(volumes) > 0 {
				listMutex.Lock()
				diskList = append(diskList, volumes...)
				listMutex.Unlock()
			}
		})
		resultMutex.Lock()
		result[project] = diskList
		resultMutex.Unlock()
	})
	return result
}

func (m *gcpResourceManager) SnapshotsPerAccount() map[string][]Snapshot {
	log.Println("Getting snapshots in all projects")
	result := make(map[string][]Snapshot)
	var resultMutex sync.Mutex
	m.forEachProject(func(project string) {
		snapshots, err := m.getSnapshots(project)
		if err != nil {
			log.Printf("Could not list snapshots in %s: %s", project, err)
			if err == ErrPermissionDenied {
				log.Println(err)
			} else {
				// If it was an unknown error, abort
				log.Fatalln(err)
			}
		} else if len(snapshots) > 0 {
			resultMutex.Lock()
			result[project] = snapshots
			resultMutex.Unlock()
		}
	})
	return result
}

func (m *gcpResourceManager) BucketsPerAccount() map[string][]Bucket {
	log.Println("Getting buckets in all projects")
	result := make(map[string][]Bucket)
	var resultMutex sync.Mutex
	m.forEachProject(func(project string) {
		buckets, err := m.getBuckets(project)
		if err != nil {
			log.Printf("Could not list buckets in %s: %s", project, err)
			if err == ErrPermissionDenied {
				log.Println(err)
			} else {
				// If it was an unknown error, abort
				log.Fatalln(err)
			}
		} else if len(buckets) > 0 {
			resultMutex.Lock()
			result[project] = buckets
			resultMutex.Unlock()
		}
	})
	return result
}

func (m *gcpResourceManager) AllResourcesPerAccount() map[string]*ResourceCollection {
	log.Println("Getting all compute resources in all accounts")
	result := make(map[string]*ResourceCollection)
	var resultMutex sync.Mutex
	var wg sync.WaitGroup
	var instanceMap map[string][]Instance
	var imageMap map[string][]Image
	var volumeMap map[string][]Volume
	var snapMap map[string][]Snapshot
	wg.Add(4)
	go func() {
		instanceMap = m.InstancesPerAccount()
		wg.Done()
	}()
	go func() {
		imageMap = m.ImagesPerAccount()
		wg.Done()
	}()
	go func() {
		volumeMap = m.VolumesPerAccount()
		wg.Done()
	}()
	go func() {
		snapMap = m.SnapshotsPerAccount()
		wg.Done()
	}()
	wg.Wait()
	for _, project := range m.projects {
		collection := &ResourceCollection{
			Owner:     project,
			Instances: instanceMap[project],
			Images:    imageMap[project],
			Volumes:   volumeMap[project],
			Snapshots: snapMap[project],
		}
		resultMutex.Lock()
		result[project] = collection
		resultMutex.Unlock()
	}
	return result
}

func (m *gcpResourceManager) CleanupInstances(instances []Instance) error {
	return cleanupInstances(instances)
}

func (m *gcpResourceManager) CleanupImages(images []Image) error {
	return cleanupImages(images)
}

func (m *gcpResourceManager) CleanupVolumes(volumes []Volume) error {
	return cleanupVolumes(volumes)
}

func (m *gcpResourceManager) CleanupSnapshots(snapshots []Snapshot) error {
	return cleanupSnapshots(snapshots)
}

func (m *gcpResourceManager) CleanupBuckets(buckets []Bucket) error {
	return cleanupBuckets(buckets)
}

func (m *gcpResourceManager) forEachProject(f func(project string)) {
	var wg sync.WaitGroup
	wg.Add(len(m.projects))
	for i := range m.projects {
		go func(i int) {
			log.Printf("Accessing project %s", m.projects[i])
			f(m.projects[i])
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (m *gcpResourceManager) forEachZone(project string, f func(zone string)) {
	zones, err := m.compute.Zones.List(project).Do()
	if err != nil {
		log.Printf("Could not list zones in %s. Err: %v", project, err)
		return
	}
	var wg sync.WaitGroup
	for _, z := range zones.Items {
		wg.Add(1)
		go func(z string) {
			f(z)
			wg.Done()
		}(z.Name)
	}
	wg.Wait()
}

func (m *gcpResourceManager) getInstances(project, zone string) ([]Instance, error) {
	instances, err := m.compute.Instances.List(project, zone).Do()
	if err != nil {
		if instances != nil && isGCPAccessDeniedError(instances.HTTPStatusCode) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	res := []Instance{}
	for _, i := range instances.Items {
		creationTime, err := time.Parse(time.RFC3339, i.CreationTimestamp)
		if err != nil {
			log.Printf("Could not parse timestamp of %s (in %s): %s", i.Name, project, err)
			// Set to Now so it doesn't incorrecntly get tagged for deletion
			creationTime = time.Now()
		}
		labels := i.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		res = append(res, &gcpInstance{baseInstance{
			baseResource: baseResource{
				csp:          GCP,
				owner:        project,
				id:           i.Name,
				location:     zone,
				public:       true,
				tags:         i.Labels,
				creationTime: creationTime,
			},
			instanceType: parseGCPResourceURL(i.MachineType),
		},
			m.compute,
		})
	}
	return res, nil
}

func (m *gcpResourceManager) getImages(project string) ([]Image, error) {
	images, err := m.compute.Images.List(project).Do()
	if err != nil {
		if images != nil && isGCPAccessDeniedError(images.HTTPStatusCode) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	imgList := []Image{}
	for _, img := range images.Items {
		creationTime, err := time.Parse(time.RFC3339, img.CreationTimestamp)
		if err != nil {
			log.Printf("Could not parse timestamp of %s (in %s): %s", img.Name, project, err)
			// Set to Now so it doesn't incorrecntly get tagged for deletion
			creationTime = time.Now()
		}
		labels := img.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		imgList = append(imgList, &gcpImage{
			baseImage: baseImage{
				baseResource: baseResource{
					csp:          GCP,
					id:           img.Name,
					owner:        project,
					location:     "",
					creationTime: creationTime,
					tags:         labels,
					public:       true,
				},
				name:   img.Name,
				sizeGB: img.DiskSizeGb,
			},
			compute: m.compute,
		})
	}
	return imgList, nil
}

func (m *gcpResourceManager) getVolumes(project, zone string) ([]Volume, error) {
	volumes, err := m.compute.Disks.List(project, zone).Do()
	if err != nil {
		if volumes != nil && isGCPAccessDeniedError(volumes.HTTPStatusCode) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	diskList := []Volume{}
	for _, disk := range volumes.Items {
		creationTime, err := time.Parse(time.RFC3339, disk.CreationTimestamp)
		if err != nil {
			log.Printf("Could not parse timestamp of %s (in %s): %s", disk.Name, project, err)
			// Set to Now so it doesn't incorrecntly get tagged for deletion
			creationTime = time.Now()
		}
		labels := disk.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		diskList = append(diskList, &gcpVolume{
			baseVolume: baseVolume{
				baseResource: baseResource{
					csp:          GCP,
					owner:        project,
					id:           disk.Name,
					location:     zone,
					creationTime: creationTime,
					public:       true,
					tags:         labels,
				},
				sizeGB:     disk.SizeGb,
				encrypted:  false,
				attached:   disk.Users != nil && len(disk.Users) > 0,
				volumeType: parseGCPResourceURL(disk.Type),
			},
			compute: m.compute,
		})
	}
	return diskList, nil
}

func (m *gcpResourceManager) getSnapshots(project string) ([]Snapshot, error) {
	snapshots, err := m.compute.Snapshots.List(project).Do()
	if err != nil {
		if snapshots != nil && isGCPAccessDeniedError(snapshots.HTTPStatusCode) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	snapList := []Snapshot{}
	for _, snap := range snapshots.Items {
		creationTime, err := time.Parse(time.RFC3339, snap.CreationTimestamp)
		if err != nil {
			log.Printf("Could not parse timestamp of %s (in %s): %s", snap.Name, project, err)
			// Set to Now so it doesn't incorrecntly get tagged for deletion
			creationTime = time.Now()
		}
		labels := snap.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		snapList = append(snapList, &gcpSnapshot{
			baseSnapshot: baseSnapshot{
				baseResource: baseResource{
					csp:          GCP,
					id:           snap.Name,
					owner:        project,
					location:     "",
					public:       true,
					creationTime: creationTime,
					tags:         labels,
				},
				encrypted: false,
				inUse:     false,
				sizeGB:    snap.DiskSizeGb,
			},
			compute: m.compute,
		})
	}
	return snapList, nil
}

func (m *gcpResourceManager) getBuckets(project string) ([]Bucket, error) {
	buckets, err := m.storage.Buckets.List(project).Do()
	if err != nil {
		if buckets != nil && isGCPAccessDeniedError(buckets.HTTPStatusCode) {
			return nil, ErrPermissionDenied
		}
		return nil, err
	}
	buckList := []Bucket{}
	for _, buck := range buckets.Items {
		creationTime, err := time.Parse(time.RFC3339, buck.TimeCreated)
		if err != nil {
			// Set to Now so it doesn't incorrecntly get tagged for deletion
			creationTime = time.Now()
		}
		lastModified, err := time.Parse(time.RFC3339, buck.Updated)
		if err != nil {
			lastModified = time.Time{}
		}
		labels := buck.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		count, size, err := m.bucketDetails(buck.Name)
		if err != nil {
			log.Printf("Could not get object details for %s: %s", buck.Name, err)
		}
		buckList = append(buckList, &gcpBucket{
			baseBucket: baseBucket{
				baseResource: baseResource{
					csp:          GCP,
					owner:        project,
					id:           buck.Name,
					tags:         labels,
					creationTime: creationTime,
					public:       false,
					location:     buck.Location,
				},
				lastModified:       lastModified,
				objectCount:        count,
				totalSizeGB:        size,
				storageTypeSizesGB: make(map[string]float64),
			},
			storage: m.storage,
		})
	}
	return buckList, nil
}

// bucketDetails will determine how many objects there are in a bucket and what
// the total bucket size is.
func (m *gcpResourceManager) bucketDetails(bucketID string) (int64, float64, error) {
	var count int64
	var sizeGB float64
	var nextPageToken string
	for ok := true; ok; ok = nextPageToken != "" {
		objs, err := m.storage.Objects.List(bucketID).Do()
		if err != nil {
			if objs != nil && isGCPAccessDeniedError(objs.HTTPStatusCode) {
				return 0, 0.0, ErrPermissionDenied
			}
			return 0, 0.0, err
		}
		nextPageToken = objs.NextPageToken
		for _, obj := range objs.Items {
			sizeGB += (float64(obj.Size) / gbDivider)
			count++
		}
	}
	return count, sizeGB, nil
}

// Figure out if http response code is permission denied
func isGCPAccessDeniedError(code int) bool {
	switch code {
	case 403:
		return true
	case 401:
		return true
	default:
		return false
	}
}

func parseGCPResourceURL(in string) string {
	parts := strings.Split(in, "/")
	n := len(parts)
	if n > 0 {
		return parts[n-1]
	}
	return in
}
