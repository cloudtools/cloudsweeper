package cloud

import (
	"time"
)

type baseBucket struct {
	baseResource
	lastModified time.Time
	objectCount  int64
	totalSizeGB  float64
}

func (b *baseBucket) LastModified() time.Time {
	return b.lastModified
}

func (b *baseBucket) ObjectCount() int64 {
	return b.objectCount
}

func (b *baseBucket) TotalSizeGB() float64 {
	return b.totalSizeGB
}
