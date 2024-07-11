package sortedset

type SortedSet struct {
	dict map[string]*Pair
	skl  *skiplist
}

func NewSortedSet() *SortedSet {
	s := &SortedSet{}
	s.dict = make(map[string]*Pair)
	s.skl = newSkipList()
	return s
}

func (s *SortedSet) Add(member string, score float64) bool {
	pair, ok := s.dict[member]
	if ok {
		if pair.Score == score {
			return false
		}
		s.skl.Remove(member, score)
		s.skl.Insert(member, score)
	}
	s.skl.Insert(member, score)
	s.dict[member] = &Pair{member, score}
	return true
}

func (s *SortedSet) Remove(member string) bool {
	pair, ok := s.dict[member]
	if !ok {
		return false
	}
	s.skl.Remove(member, pair.Score)
	delete(s.dict, member)
	return true
}

func (s *SortedSet) Get(member string) (*Pair, bool) {
	pair, ok := s.dict[member]
	return pair, ok
}

func (s *SortedSet) Len() int64 {
	return int64(len(s.dict))
}

// 获取在跳表中的索引序号
func (s *SortedSet) GetRank(member string, desc bool) int64 {
	pair, ok := s.dict[member]
	if !ok {
		return -1
	}
	rank := s.skl.getRank(member, pair.Score)
	if desc {
		rank = s.skl.length - rank
	} else {
		rank--
	}
	return rank
}
