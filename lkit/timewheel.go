package lkit

import (
	"container/list"
	"context"
	"sync"
	"time"

	"p2p_remote_desk/llog"
)

// TimeWheel 时间轮结构
type TimeWheel struct {
	interval       time.Duration // 时间轮槽间隔
	ticker         *time.Ticker  // 时间轮定时器
	slots          []*list.List  // 时间轮槽
	currentPos     int           // 当前时间轮指针位置
	slotNum        int           // 槽数量
	addTaskChan    chan *Task    // 添加任务通道
	removeTaskChan chan *Task    // 移除任务通道
	stopChan       chan struct{} // 停止时间轮通道
	sync.RWMutex                 // 读写锁
}

// Task 定时任务
type Task struct {
	delay   time.Duration // 延迟时间
	circle  int           // 时间轮需要转动的圈数
	key     interface{}   // 任务的唯一标识
	job     func()        // 任务函数
	removed bool          // 任务是否已被移除
}

// NewTimeWheel 创建时间轮
// interval: 时间轮槽间隔
// slotNum: 槽数量
func NewTimeWheel(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}

	tw := &TimeWheel{
		interval:       interval,
		slots:          make([]*list.List, slotNum),
		currentPos:     0,
		slotNum:        slotNum,
		addTaskChan:    make(chan *Task),
		removeTaskChan: make(chan *Task),
		stopChan:       make(chan struct{}),
	}

	// 初始化每个槽
	for i := 0; i < slotNum; i++ {
		tw.slots[i] = list.New()
	}

	return tw
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)

	SafeGo(func() {
		tw.run()
	})
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.stopChan <- struct{}{}
}

// AddTask 添加定时任务
func (tw *TimeWheel) AddTask(delay time.Duration, key interface{}, job func()) *Task {
	if delay < 0 {
		return nil
	}

	task := &Task{
		delay:   delay,
		key:     key,
		job:     job,
		removed: false,
	}

	tw.addTaskChan <- task
	return task
}

// RemoveTask 移除定时任务
func (tw *TimeWheel) RemoveTask(task *Task) {
	if task == nil {
		return
	}

	tw.removeTaskChan <- task
}

// run 运行时间轮
func (tw *TimeWheel) run() {
	defer func() {
		if tw.ticker != nil {
			tw.ticker.Stop()
		}
	}()

	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChan:
			tw.addTask(task)
		case task := <-tw.removeTaskChan:
			tw.removeTask(task)
		case <-tw.stopChan:
			return
		}
	}
}

// tickHandler 处理时间轮的一次转动
func (tw *TimeWheel) tickHandler() {
	tw.Lock()
	defer tw.Unlock()

	// 获取当前槽的任务列表
	currentSlot := tw.slots[tw.currentPos]

	// 遍历当前槽中的所有任务
	for e := currentSlot.Front(); e != nil; {
		task := e.Value.(*Task)

		// 获取下一个元素（因为当前元素可能会被删除）
		next := e.Next()

		// 如果任务已被标记为移除，则直接删除
		if task.removed {
			currentSlot.Remove(e)
			e = next
			continue
		}

		// 如果任务还需要等待更多圈数
		if task.circle > 0 {
			task.circle--
			e = next
			continue
		}

		// 执行任务
		SafeGo(func() {
			defer func() {
				if r := recover(); r != nil {
					llog.Error("时间轮任务执行panic:", r)
				}
			}()
			task.job()
		})

		// 从当前槽中移除任务
		currentSlot.Remove(e)

		e = next
	}

	// 时间轮指针前进一步
	tw.currentPos = (tw.currentPos + 1) % tw.slotNum
}

// addTask 添加任务到时间轮
func (tw *TimeWheel) addTask(task *Task) {
	tw.Lock()
	defer tw.Unlock()

	// 计算需要转动的圈数和对应的槽位置
	pos, circle := tw.getPositionAndCircle(task.delay)

	// 设置任务的圈数
	task.circle = circle

	// 将任务添加到对应的槽中
	tw.slots[pos].PushBack(task)
}

// removeTask 从时间轮中移除任务
func (tw *TimeWheel) removeTask(task *Task) {
	tw.Lock()
	defer tw.Unlock()

	// 标记任务为已移除
	task.removed = true
}

// getPositionAndCircle 计算任务在时间轮中的位置和需要转动的圈数
func (tw *TimeWheel) getPositionAndCircle(delay time.Duration) (int, int) {
	// 计算需要的槽数
	delaySeconds := int(delay / tw.interval)

	// 计算位置和圈数
	pos := (tw.currentPos + delaySeconds) % tw.slotNum
	circle := delaySeconds / tw.slotNum

	return pos, circle
}

// ScheduleTask 创建并启动一个定时任务
// interval: 任务执行间隔
// key: 任务的唯一标识
// job: 任务函数
func (tw *TimeWheel) ScheduleTask(interval time.Duration, key interface{}, job func()) {
	// 添加首次执行的任务
	task := tw.AddTask(interval, key, func() {
		// 执行任务
		job()

		// 重新添加任务，实现循环执行
		tw.AddTask(interval, key, job)
	})

	// 如果添加失败，记录错误
	if task == nil {
		llog.Error("添加定时任务失败:", key)
	}
}

// ScheduleTaskWithContext 创建并启动一个带上下文的定时任务
// ctx: 上下文，用于控制任务的生命周期
// interval: 任务执行间隔
// key: 任务的唯一标识
// job: 任务函数
func (tw *TimeWheel) ScheduleTaskWithContext(ctx context.Context, interval time.Duration, key interface{}, job func()) {
	// 创建一个任务映射，用于存储任务引用
	var currentTask *Task

	// 添加首次执行的任务
	currentTask = tw.AddTask(interval, key, func() {
		select {
		case <-ctx.Done():
			// 上下文已取消，不再重新添加任务
			return
		default:
			// 执行任务
			job()

			// 重新添加任务，实现循环执行
			currentTask = tw.AddTask(interval, key, currentTask.job)
		}
	})

	// 如果添加失败，记录错误
	if currentTask == nil {
		llog.Error("添加定时任务失败:", key)
	}

	// 监听上下文取消信号
	SafeGoWithContext(ctx, func(ctx context.Context) {
		<-ctx.Done()
		if currentTask != nil {
			tw.RemoveTask(currentTask)
		}
	})
}

func testTimeWheel() {
	tw := NewTimeWheel(1*time.Second, 60)
	tw.Start()

	// 添加任务
	tw.ScheduleTask(5*time.Second, "task1", func() {
		llog.Info("Task 1 executed")
	})

	// 添加带上下文的任务
	ctx, cancel := context.WithCancel(context.Background())
	tw.ScheduleTaskWithContext(ctx, 10*time.Second, "task2", func() {
		llog.Info("Task 2 executed")
	})

	// 模拟取消上下文
	time.Sleep(15 * time.Second)
	cancel()

	//// 创建时间轮，槽间隔为100毫秒，共60个槽（覆盖6秒）
	//timeWheel := lkit.NewTimeWheel(100*time.Millisecond, 60)
	//
	//// 启动时间轮
	//timeWheel.Start()
	//
	//// 添加一次性任务，2秒后执行
	//timeWheel.AddTask(2*time.Second, "one-time-task", func() {
	//	llog.Info("一次性任务执行")
	//})
	//
	//// 添加周期性任务，每5秒执行一次
	//timeWheel.ScheduleTask(5*time.Second, "periodic-task", func() {
	//	llog.Info("周期性任务执行")
	//})
	//
	//// 添加带上下文控制的周期性任务
	//ctx, cancel := context.WithCancel(context.Background())
	//timeWheel.ScheduleTaskWithContext(ctx, 3*time.Second, "context-task", func() {
	//	llog.Info("带上下文的周期性任务执行")
	//})
	//
	//// 停止特定任务
	//cancel()
	//
	//// 程序结束前停止时间轮
	//defer timeWheel.Stop()
}
