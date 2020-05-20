package Server

import (
	"sync"
	"github.com/jacksonyoudi/ratelimit"
)

type BucketMap struct {
	Map sync.Map
}

func (b *BucketMap) set(key string, bucket *ratelimit.Bucket) {
	b.Map.Store(key, bucket)
}
func (b *BucketMap) get(key string) (*ratelimit.Bucket, bool) {
	load, ok := b.Map.Load(key)
	if ok {
		return load.(*ratelimit.Bucket), ok
	} else {
		return nil, ok
	}
}

// timer
