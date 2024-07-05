package dict

import (
	"gedis/util"
	"sync"
	"sync/atomic"
)

type ConcurrentDict struct {
	shds  []*shard
	mask  uint32
	count *atomic.Int32
}

// 分片 减小锁的粒度
type shard struct {
	m  map[string]any
	mu sync.RWMutex
}

func NewConcurrentDict(shardCount int) *ConcurrentDict {
	// 计算得到大于shardCount的2的整数幂的值作为容量
	shardCount = util.ComputeCapacity(shardCount)
	dict := &ConcurrentDict{}
	shds := make([]*shard, shardCount)
	for i := 0; i < shardCount; i++ {
		shds[i] = &shard{m: make(map[string]any)}
	}
	dict.shds = shds
	dict.mask = uint32(shardCount - 1)
	dict.count = &atomic.Int32{}
	return dict
}

func (d *ConcurrentDict) index(code uint32) uint32 {
	return code & d.mask
}

// 获取分片
func (d *ConcurrentDict) getShard(key string) *shard {
	return d.shds[d.index(util.Fnv32(key))]
}

// 从分片获取key对应的值
func (d *ConcurrentDict) Get(key string) (any, bool) {
	shard := d.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	v, exist := shard.m[key]
	return v, exist
}

// 获取元素个数
func (d *ConcurrentDict) Count() int {
	return int(d.count.Load())
}

// 元素个数+1
func (d *ConcurrentDict) Inc() {
	d.count.Add(1)
}

// 元素个数-1
func (d *ConcurrentDict) Dec() {
	d.count.Add(-1)
}

// 从分片中删除key
func (d *ConcurrentDict) Delete(key string) (any, int) {
	shard := d.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	v, exist := shard.m[key]
	if exist {
		delete(shard.m, key)
		d.Dec()
		return v, 1
	}
	return nil, 0
}

// 往分片中添加元素
func (d *ConcurrentDict) Put(key string, value any) int {
	shard := d.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	_, exist := shard.m[key]
	if exist {
		shard.m[key] = value
		return 0
	}
	d.Inc()
	shard.m[key] = value
	return 1
}
