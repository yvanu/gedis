package timewheel

import "time"

type Delay struct {
	tw *TimeWheel
}

func NewDelay() *Delay {
	delay := &Delay{}
	delay.tw = NewTimeWheel(time.Second, 3600)
	delay.tw.Start()
	return delay
}

// 添加绝对时间延迟任务
func (d *Delay) AddAt(expireTime time.Time, key string, callback func()) {
	d.Add(time.Until(expireTime), key, callback)
}

// 添加相对时间延迟任务
func (d *Delay) Add(delay time.Duration, key string, callback func()) {
	d.tw.AddTask(delay, key, callback)
}

func (d *Delay) Cancel(key string) {
	d.tw.CancelTask(key)
}
