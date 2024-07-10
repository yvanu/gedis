package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	defaultMaxLevel = 1 << 4
)

type Pair struct {
	Member string
	Score  float64
}

type Level struct {
	forward *node
	span    int64
}

type node struct {
	Pair
	backward *node
	levels   []*Level
}

func newNode(level int16, member string, score float64) *node {
	n := &node{
		Pair:   Pair{member, score},
		levels: make([]*Level, level),
	}
	for i := range n.levels {
		n.levels[i] = &Level{}
	}
	return n
}

type skiplist struct {
	header   *node
	tailer   *node
	length   int64 // 节点个数
	maxLevel int16 // 最高层数
}

func newSkipList() *skiplist {
	s := &skiplist{
		// 虚拟节点
		header:   newNode(defaultMaxLevel, "", 0),
		maxLevel: 1,
	}
	return s
}

// 随机层高
func randomLevel() int16 {
	total := uint64(1)<<defaultMaxLevel - 1           // 0x 00 00 00 00 00 00 FF FF  defaultMaxLevel个1
	k := rand.Uint64()%total + 1                      // [1, 0xffff]
	return defaultMaxLevel - int16(bits.Len64(k)) + 1 // 最终效果是达到 [1, defaultMaxLevel]
}

func (s *skiplist) Insert(member string, score float64) *node {
	beforeNode := make([]*node, defaultMaxLevel)      // 索引表示层数，前驱节点
	beforeNodeOrder := make([]int64, defaultMaxLevel) // 前驱节点编号

	node := s.header
	for i := s.maxLevel - 1; i >= 0; i-- {

		if i == s.maxLevel-1 {
			beforeNodeOrder[i] = 0
		} else {
			beforeNodeOrder[i] = beforeNodeOrder[i+1]
		}

		for node.levels[i].forward != nil &&
			(node.levels[i].forward.Score < score || (node.levels[i].forward.Score == score && node.levels[i].forward.Member < member)) {
			beforeNodeOrder[i] += node.levels[i].span
			node = node.levels[i].forward
		}
		beforeNode[i] = node
	}

	newLevel := randomLevel()
	if newLevel > s.maxLevel {
		for i := s.maxLevel; i < newLevel; i++ {
			beforeNodeOrder[i] = 0
			beforeNode[i] = s.header

			beforeNode[i].levels[i].forward = nil
			beforeNode[i].levels[i].span = s.length
		}
		s.maxLevel = newLevel
	}

	node = newNode(newLevel, member, score)

	for i := int16(0); i < newLevel; i++ {
		node.levels[i].forward = beforeNode[i].levels[i].forward
		beforeNode[i].levels[i].forward = node

		node.levels[i].span = beforeNode[i].levels[i].span - (beforeNodeOrder[0] - beforeNodeOrder[i])
		beforeNode[i].levels[i].span = beforeNodeOrder[0] - beforeNodeOrder[i] + 1
	}

	for i := newLevel; i < s.maxLevel; i++ {
		// 前驱节点的跨度++ 相当于跳过了新的节点
		beforeNode[i].levels[i].span++
	}

	if beforeNode[0] == s.header {
		node.backward = nil
	} else {
		node.backward = beforeNode[0]
	}

	if node.levels[0].forward != nil {
		node.levels[0].forward.backward = node
	} else {
		s.tailer = node
	}

	s.length++
	return node
}

func (s *skiplist) Remove(member string, score float64) bool {
	beforeNode := make([]*node, defaultMaxLevel)
	n := s.header

	// 获取被删除节点的前驱节点
	for i := s.maxLevel - 1; i >= 0; i-- {
		for n.levels[i].forward != nil &&
			(n.levels[i].forward.Score < score || (n.levels[i].forward.Score == score && n.levels[i].forward.Member < member)) {
			n = n.levels[i].forward
		}
		beforeNode[i] = n
	}

	// 待删除的节点
	n = n.levels[0].forward

	if n != nil && n.Score == score && n.Member == member {
		s.removeNode(n, beforeNode)
		return true
	}
	return false
}

func (s *skiplist) removeNode(n *node, beforeNode []*node) {
	for i := int16(0); i < s.maxLevel; i++ {
		if beforeNode[i].levels[i].forward == n {

			beforeNode[i].levels[i].span += n.levels[i].span - 1
			beforeNode[i].levels[i].forward = n.levels[i].forward
		} else {
			beforeNode[i].levels[i].span--
		}
	}

	if n.levels[0].forward != nil {
		n.levels[0].forward.backward = n.backward
	} else {
		s.tailer = n.backward
	}

	for s.maxLevel > 1 && s.header.levels[s.maxLevel-1].forward == nil {
		s.maxLevel--
	}

	s.length--
}

func (s *skiplist) getRank(member string, score float64) int64 {
	var nodeOrder int64
	n := s.header
	for i := s.maxLevel - 1; i >= 0; i-- {
		for n.levels[i].forward != nil &&
			(n.levels[i].forward.Score < score || (n.levels[i].forward.Score == score && n.levels[i].forward.Member < member)) {
			nodeOrder += n.levels[i].span
			nodeOrder += n.levels[i].span
			n = n.levels[i].forward
		}
		if n != s.header && n.Score == score && n.Member == member {
			return nodeOrder
		}
	}
	return 0
}

func (s *skiplist) getByRank(rank int64) *node {
	var order int64
	n := s.header
	for i := s.maxLevel - 1; i >= 0; i-- {
		if n.levels[i].forward != nil && (order+n.levels[i].span <= order) {
			n = n.levels[i].forward
			order += n.levels[i].span
		}
		if order == rank {
			return n
		}
	}
	return nil
}
