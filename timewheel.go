package timewheel

import (
	"container/list"
	"time"
)

// Job 任务回调函数
type Job func(interface{})

// TimeWheel 时间轮
type TimeWheel struct {
	interval          time.Duration       // 时间轮多久往前移动一格
	ticker            *time.Ticker        // 定时器
	slots             []*list.List        // 时间轮槽
	slotNum           int                 // 时间轮槽数量
	currentPos        int                 // 当前在槽中的位置
	timer             map[interface{}]int // 保存对应key在时间轮槽中的位置，用于删除操作
	job               Job                 // 执行的任务
	addTaskChannel    chan *Task          // 增加任务Channel
	removeTaskChannel chan interface{}    // 移除任务Channel
	stopChanel        chan bool           // 时间轮停止Channel
}

type Task struct {
	delay  time.Duration // 延迟时间
	circle int           // 圈数
	key    interface{}   // key用来执行删除操作
	data   interface{}   // 回调参数
}

// New 创建时间轮
func New(interval time.Duration, slotNum int, job Job) *TimeWheel {
	if interval <= 0 || slotNum <= 0 || job == nil {
		return nil
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		slotNum:           slotNum,
		currentPos:        0,
		timer:             make(map[interface{}]int),
		job:               job,
		addTaskChannel:    make(chan *Task),
		removeTaskChannel: make(chan interface{}),
		stopChanel:        make(chan bool),
	}
	tw.initSlot()
	return tw
}

// 初始化槽，每个槽都指向一个双向链表
func (tw *TimeWheel) initSlot() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// AddTask 添加定时任务
func (tw *TimeWheel) AddTask(delay time.Duration, key interface{}, data interface{}) {
	if delay <= 0 {
		return
	}
	tw.addTaskChannel <- &Task{delay: delay, key: key, data: data}
}

// RemoveTask 移除定时任务
func (tw *TimeWheel) RemoveTask(key interface{}) {
	tw.removeTaskChannel <- key
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.stopChanel <- true
}

// 启动时间轮
func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChanel:
			tw.ticker.Stop()
			return
		}
	}
}

// ticker处理函数
func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
}

// 扫描并运行任务
func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for i := l.Front(); i != nil; {
		task := i.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			i = i.Next()
			continue
		}

		// 执行回调
		go tw.job(task.data)
		next := i.Next()
		l.Remove(i)
		if task.key != nil {
			delete(tw.timer, task.key)
		}
		i = next
	}
}

// 将任务添加到任务链表中
func (tw *TimeWheel) addTask(task *Task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle
	tw.slots[pos].PushBack(task)
	if task.key != nil {
		tw.timer[task.key] = pos
	}
}

// 获取在时间轮中的位置和需要转动的圈数
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (int, int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	pos := (tw.currentPos + (delaySeconds-1)/intervalSeconds) % tw.slotNum
	circle := delaySeconds / intervalSeconds / tw.slotNum
	return pos, circle
}

// 根据Key移除定时任务
func (tw *TimeWheel) removeTask(key interface{}) {
	pos, ok := tw.timer[key]
	if !ok {
		return
	}
	l := tw.slots[pos]
	for i := l.Front(); i != nil; i = i.Next() {
		task := i.Value.(*Task)
		if task.key == key {
			delete(tw.timer, key)
			l.Remove(i)
		}
	}
}
