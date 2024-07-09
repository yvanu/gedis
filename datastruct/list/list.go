package list

import "errors"

type Consumer func(idx int, val any) bool
type Except func(val any) bool

type node struct {
	pre  *node
	next *node
	val  any
}

// LinkedList 双向链表
type LinkedList struct {
	first *node
	last  *node

	size int
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

func newNode(val any) *node {
	return &node{val: val}
}

func (l *LinkedList) Add(val any) {
	n := newNode(val)
	// 判断是不是空链表
	if l.first == nil {
		l.first = n
		l.last = n
	} else {
		l.last.next = n
		n.pre = l.last
		l.last = n
	}
	l.size++
}

func (l *LinkedList) find(idx int) *node {
	if idx < 0 || idx >= l.size {
		return nil
	}

	if idx < l.size/2 {
		current := l.first
		for i := 0; i < idx; i++ {
			current = current.next
		}
		return current
	} else {
		current := l.last
		for i := l.size - 1; i > idx; i-- {
			current = current.pre
		}
		return current
	}
}

func (l *LinkedList) Get(idx int) (any, error) {
	n := l.find(idx)
	if n == nil {
		return nil, errors.New("idx error")
	}
	return n.val, nil
}

func (l *LinkedList) Update(idx int, val any) error {
	n := l.find(idx)
	if n == nil {
		return errors.New("idx error")
	}
	n.val = val
	return nil
}

func (l *LinkedList) del(n *node) {
	pre := n.pre
	next := n.next
	if pre != nil {
		pre.next = next
	} else {
		l.first = next
	}
	if next != nil {
		next.pre = pre
	} else {
		l.last = pre
	}

	// gc用
	n.pre = nil
	n.next = nil

	l.size--
}

func (l *LinkedList) Del(idx int) (any, error) {
	n := l.find(idx)
	if n == nil {
		return nil, errors.New("idx error")
	}
	l.del(n)
	return n.val, nil
}

func (l *LinkedList) ForEach(consumer Consumer) {
	i := 0
	for n := l.first; n != nil; n = n.next {
		if !consumer(i, n.val) {
			break
		}
		i++
	}
}

func (l *LinkedList) Contain(except Except) bool {
	var result bool
	l.ForEach(func(idx int, val any) bool {
		if except(val) {
			result = true
			return false
		}
		return true
	})
	return result
}

func (l *LinkedList) DelByVal(except Except) int {
	var removed int
	l.ForEach(func(idx int, val any) bool {
		if except(val) {
			l.del(l.find(idx))
			removed++
		}
		return true
	})
	return removed
}

func (l *LinkedList) Len() int {
	return l.size
}
