package timewheel

import (
	"container/list"
	"gedis/tool/logger"
	"runtime/debug"
	"time"
)

type taskPos struct {
	pos int
	ele *list.Element
}

type task struct {
	delay    time.Duration
	key      string
	circle   int
	callback func()
}

type TimeWheel struct {
	// 间隔
	interval time.Duration
	// 定时器
	ticker *time.Ticker

	// 游标
	curSlotPos int
	// 队列大小
	slotNum int
	// 圈的点的个数
	slots []*list.List
	m     map[string]*taskPos

	addChan    chan *task
	cancelChan chan string
	stopChan   chan struct{}
}

func NewTimeWheel(interval time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		interval:   interval,
		slotNum:    slotNum,
		slots:      make([]*list.List, slotNum),
		m:          make(map[string]*taskPos),
		addChan:    make(chan *task),
		cancelChan: make(chan string),
		stopChan:   make(chan struct{}),
	}
	for i := 0; i < slotNum; i++ {
		tw.slots[i] = list.New()
	}
	return tw
}

func (tw *TimeWheel) doTask() {
	for {
		select {
		case <-tw.ticker.C:
			tw.execTask()
		case t := <-tw.addChan:
			tw.addTask(t)
		case key := <-tw.cancelChan:
			tw.cancelTask(key)
		case <-tw.stopChan:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) execTask() {
	// 获取要执行的链表
	l := tw.slots[tw.curSlotPos]
	//开启下一次循环
	if tw.curSlotPos == tw.slotNum-1 {
		tw.curSlotPos = 0
	} else {
		tw.curSlotPos++
	}
	go tw.scanList(l)
}

func (tw *TimeWheel) scanList(l *list.List) {
	e := l.Front()
	for e != nil {
		t := e.Value.(*task)
		// 不属于该圈任务
		if t.circle > 0 {
			t.circle--
			continue
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err, string(debug.Stack()))
				}
			}()
			if t.callback != nil {
				t.callback()
			}

		}()
		next := e.Next()
		l.Remove(e)
		if t.key != "" {
			delete(tw.m, t.key)
		}
		e = next
	}
}

func (tw *TimeWheel) addTask(t *task) {
	if t.key != "" {
		if _, exist := tw.m[t.key]; exist {
			return
		}
	}
	pos, circle := tw.posAndCircle(t.delay)
	t.circle = circle

	ele := tw.slots[pos].PushBack(t)
	tw.m[t.key] = &taskPos{pos: pos, ele: ele}
}

func (tw *TimeWheel) posAndCircle(delay time.Duration) (int, int) {
	// 延迟
	delaySecond := int(delay.Seconds())
	// 间隔
	intervalSecond := int(tw.interval.Seconds())
	pos := (tw.curSlotPos + delaySecond/intervalSecond) % tw.slotNum
	circle := (delaySecond / intervalSecond) / tw.slotNum
	return pos, circle
}

func (tw *TimeWheel) cancelTask(key string) {
	taskPos, exist := tw.m[key]
	if !exist {
		return
	}
	tw.slots[taskPos.pos].Remove(taskPos.ele)
	delete(tw.m, key)
}

/************外部调用的方法*****************/

func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.doTask()
}

func (tw *TimeWheel) Stop() {
	tw.stopChan <- struct{}{}
}

func (tw *TimeWheel) AddTask(delay time.Duration, key string, callback func()) {
	if delay < 0 {
		return
	}
	tw.addChan <- &task{delay: delay, key: key, callback: callback}
}

func (tw *TimeWheel) CancelTask(key string) {
	tw.cancelChan <- key
}
