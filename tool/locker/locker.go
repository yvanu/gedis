package locker

import (
	"gedis/util"
	"sort"
	"sync"
)

type Locker struct {
	muList []*sync.RWMutex

	mask uint32
}

func NewLocker(count int) *Locker {

	l := &Locker{}

	count = util.ComputeCapacity(count)
	l.mask = uint32(count) - 1 // 全1
	l.muList = make([]*sync.RWMutex, count)

	for i := 0; i < count; i++ {
		l.muList[i] = &sync.RWMutex{}
	}

	return l
}

func (l *Locker) Locks(keys ...string) {
	// 获取这些key对应的index
	indexes := l.toIndex(keys...)

	// 依次锁住
	for _, index := range indexes {
		l.muList[index].Lock()
	}
}

func (l *Locker) Unlocks(keys ...string) {
	// 获取这些key对应的index
	indexes := l.toIndex(keys...)

	// 依次解锁
	for _, index := range indexes {
		l.muList[index].Unlock()
	}
}

func (l *Locker) RLocks(keys ...string) {
	// 获取这些key对应的index
	indexes := l.toIndex(keys...)

	// 依次锁住
	for _, index := range indexes {
		l.muList[index].RLock()
	}
}

func (l *Locker) RUnlocks(keys ...string) {
	// 获取这些key对应的index
	indexes := l.toIndex(keys...)

	// 依次解锁
	for _, index := range indexes {
		l.muList[index].RUnlock()
	}
}

func (l *Locker) RWLocks(wkeys []string, rkeys []string) {
	// wkey加lock，rkey加rlock

	allIndexList := l.toIndex(append(wkeys, rkeys...)...)

	wIndexMap := make(map[uint32]struct{})
	for _, key := range wkeys {
		wIndexMap[util.Fnv32(key)&l.mask] = struct{}{}
	}

	for _, idx := range allIndexList {
		mu := l.muList[idx]
		if _, ok := wIndexMap[idx]; ok {
			mu.Lock()
		} else {
			mu.RLock()
		}
	}
}

func (l *Locker) RWUnlocks(wkeys []string, rkeys []string) {
	allIndexList := l.toIndex(append(wkeys, rkeys...)...)

	wIndexMap := make(map[uint32]struct{})
	for _, key := range wkeys {
		wIndexMap[util.Fnv32(key)&l.mask] = struct{}{}
	}

	for _, idx := range allIndexList {
		mu := l.muList[idx]
		if _, ok := wIndexMap[idx]; ok {
			mu.Unlock()
		} else {
			mu.RUnlock()
		}
	}
}

func (l *Locker) toIndex(keys ...string) []uint32 {
	var indexList = make([]uint32, 0, len(keys))
	for _, key := range keys {
		indexList = append(indexList, util.Fnv32(key)&l.mask)
	}
	sort.Slice(indexList, func(i, j int) bool {
		return indexList[i] < indexList[j]
	})
	return indexList
}
