package schedule

import (
	"container/heap"
	"sync/atomic"
	"time"
)

type OnEvent func(t time.Time)

type Timer struct {
	InitTid   uint32
	timerId   uint32
	Index     int
	expired   time.Time
	interval  time.Duration
	onTimeout OnEvent
	onCancel  OnEvent
	repeated  bool
}

type TimerHeap []*Timer

func (h TimerHeap) Len() int { return len(h) }

func (h TimerHeap) Less(i, j int) bool {
	if h[i].expired.Before(h[j].expired) {
		return true
	}

	if h[i].expired.After(h[j].expired) {
		return false
	}
	return h[i].timerId < h[j].timerId
}

func (h TimerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *TimerHeap) Push(x any) {
	n := len(*h)
	item := x.(*Timer)
	item.Index = n
	*h = append(*h, item)
}

func (h *TimerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

var timerIds uint32

const (
	MinInterval = 100 * time.Millisecond
)

func timerId() uint32 {
	return atomic.AddUint32(&timerIds, 1)
}

type TimerWheel struct {
	timerHeap   TimerHeap
	tick        *time.Ticker
	hashTimer   map[uint32]*Timer
	interval    time.Duration
	cancelTimer chan uint32
	addTimer    chan *Timer
	updateTimer chan Timer
	workLimit   chan *any
}

func NewTimerWheel(interval time.Duration, workSize int) *TimerWheel {

	if int64(interval)-int64(MinInterval) < 0 {
		interval = MinInterval
	}

	tw := &TimerWheel{
		timerHeap:   make(TimerHeap, 0),
		tick:        time.NewTicker(interval),
		hashTimer:   make(map[uint32]*Timer, 10),
		interval:    interval,
		updateTimer: make(chan Timer, 2000),
		cancelTimer: make(chan uint32, 2000),
		addTimer:    make(chan *Timer, 2000),
		workLimit:   make(chan *any, workSize*2),
	}
	heap.Init(&tw.timerHeap)
	tw.start()
	return tw
}

func (t *TimerWheel) Monitor() (add, update, cancel, worker int) {
	return len(t.addTimer), len(t.updateTimer), len(t.cancelTimer), len(t.workLimit)
}

func (t *TimerWheel) After(timeout time.Duration) (uint32, chan time.Time) {
	if timeout < t.interval {
		timeout = t.interval
	}
	ch := make(chan time.Time, 1)
	tid := timerId()
	timer := &Timer{
		timerId:   tid,
		InitTid:   tid,
		expired:   time.Now().Add(timeout),
		onTimeout: func(t time.Time) { ch <- t },
		onCancel:  nil}

	t.addTimer <- timer
	return timer.timerId, ch
}

func (t *TimerWheel) RepeatedTimer(interval time.Duration, onTimout OnEvent, onCancel OnEvent) uint32 {
	if interval < t.interval {
		interval = t.interval
	}
	tid := timerId()
	timer := &Timer{
		repeated: true,
		interval: interval,
		timerId:  tid,
		InitTid:  tid,
		expired:  time.Now().Add(interval),
		onTimeout: func(t time.Time) {
			if onTimout != nil {
				onTimout(t)
			}
		},
		onCancel: onCancel,
	}
	t.addTimer <- timer
	return tid
}

func (t *TimerWheel) AddTimer(timeout time.Duration, onTimout OnEvent, onCancel OnEvent) (uint32, chan time.Time) {
	ch := make(chan time.Time, 1)
	tid := timerId()
	timer := &Timer{
		timerId:  tid,
		InitTid:  tid,
		interval: timeout,
		expired:  time.Now().Add(timeout),
		onTimeout: func(t time.Time) {
			defer func() { ch <- t }()
			if onTimout != nil {
				onTimout(t)
			}
		},
		onCancel: onCancel}

	t.addTimer <- timer
	return timer.timerId, ch
}

func (t *TimerWheel) UpdateTimer(timerid uint32, expired time.Time) {
	timer := Timer{
		timerId: timerid,
		expired: expired}
	t.updateTimer <- timer
}

func (t *TimerWheel) CancelTimer(timerid uint32) {
	t.cancelTimer <- timerid
}

func (t *TimerWheel) checkExpired(now time.Time) {
	for t.timerHeap.Len() > 0 {
		expired := t.timerHeap[0].expired
		if expired.After(now) {
			break
		}
		timer := heap.Pop(&t.timerHeap).(*Timer)
		if timer.onTimeout == nil {
			delete(t.hashTimer, timer.timerId)
			delete(t.hashTimer, timer.InitTid)
		} else {
			t.workLimit <- nil
			go func() {
				defer func() { <-t.workLimit }()
				timer.onTimeout(now)
			}()
			if timer.repeated {
				timer.expired = timer.expired.Add(timer.interval)
				if !timer.expired.After(now) {
					timer.expired = now.Add(timer.interval)
				}
				timer.timerId = timerId()
				t.onAddTimer(timer)
			} else {
				delete(t.hashTimer, timer.timerId)
				delete(t.hashTimer, timer.InitTid)
			}
		}
	}
}

func (t *TimerWheel) start() {

	go func() {
		for {
			select {
			case now := <-t.tick.C:
				t.checkExpired(now)
			case updateT := <-t.updateTimer:
				if timer, ok := t.hashTimer[updateT.timerId]; ok {
					timer.expired = updateT.expired
					heap.Fix(&t.timerHeap, timer.Index)
				}
			case timer := <-t.addTimer:
				t.onAddTimer(timer)
			case timerid := <-t.cancelTimer:
				if timer, ok := t.hashTimer[timerid]; ok {
					delete(t.hashTimer, timer.InitTid)
					delete(t.hashTimer, timer.timerId)
					heap.Remove(&t.timerHeap, timer.Index)
					if nil != timer.onCancel {
						t.workLimit <- nil
						go func() {
							<-t.workLimit
							timer.onCancel(time.Now())
						}()
					}
				}
			}
		}
	}()
}

func (t *TimerWheel) onAddTimer(timer *Timer) {
	heap.Push(&t.timerHeap, timer)
	if !timer.repeated {
		t.hashTimer[timer.timerId] = timer
	} else {
		t.hashTimer[timer.InitTid] = timer
	}
}
