package dict

import (
	"sync"
	"sync/atomic"
)

type ConcurrentDict struct {
	shds  []*shard
	mask  uint32
	count *atomic.Int32
}

type shard struct {
	m  map[string]any
	mu sync.RWMutex
}
