// https://github.com/apache/kafka/blob/7e1c453af9533aba8c19da2d08ce6595c1441fc0/server-common/src/main/java/org/apache/kafka/server/util/timer/TimingWheel.java
package timing_wheel

import (
	"reflect"

	"github.com/sjoy/server_kit/basic/time_util"
	"github.com/sjoy/server_kit/basic/xlog"
	"github.com/sjoy/server_kit/ds/circularbuffer"
)

type TimingWheel struct {
	TickTime      int64             // 一个格子的时间（毫秒）
	WheelSize     int               // 格子数量
	Interval      int64             // 整个时间轮的时间（毫秒）
	CurrentTime   int64             // 可以被 TickTime 整除的时间
	Buckets       []*TimingTaskList // WheelSize 个格子，每个格子存放一个链表
	OverflowWheel *TimingWheel      // 下一层时间轮

	LastExecutedTasks *circularbuffer.CircularBuffer[*TimingTask]

	OfferExpirationTime func(expirationTime int64) // bucket 设置过期时间成功之后，会调用这个方法，把【时间】加到优先队列里
}

func (tw *TimingWheel) AddOverflowWheel() {
	if tw.OverflowWheel == nil {
		tw.OverflowWheel = NewTimingWheel(tw.Interval, tw.WheelSize, tw.CurrentTime, tw.OfferExpirationTime)
	}
}

func (tw *TimingWheel) AddTimingTask(task *TimingTask) bool {
	expirationTime := task.ExpirationTime

	if expirationTime < tw.CurrentTime+tw.TickTime {
		// already expired
		return false
	} else if expirationTime < tw.CurrentTime-tw.CurrentTime%tw.Interval+tw.Interval {
		// tw.CurrentTime-tw.CurrentTime%tw.Interval 当前时间轮，第一个格子的开始时间(时间戳)
		// + tw.Interval 之后，得到当前时间轮最后一个格子的结束时间(时间戳)
		// 注意这里：不是用的 距离开始时间的偏移值来算 idx。 idx 的位置只和 到期时间本身有关！！！

		// put in its own bucket
		virtualId := expirationTime / tw.TickTime
		bucket := tw.Buckets[virtualId%int64(tw.WheelSize)]
		bucket.Add(task)

		// 去重
		if bucket.SetExpirationTime(virtualId * tw.TickTime) {
			tw.OfferExpirationTime(virtualId * tw.TickTime)
		}
		return true
	} else {
		// out of the interval. put it into the parent timer
		if tw.OverflowWheel == nil {
			tw.AddOverflowWheel()
		}
		return tw.OverflowWheel.AddTimingTask(task)
	}
}

func (tw *TimingWheel) AddOrRun(task *TimingTask) (added bool) {
	if !reflect.TypeOf(task.Runnable).Comparable() {
		xlog.ErrorF("task.Runnable(%v) is not comparable", reflect.TypeOf(task.Runnable))
		return false
	}
	if !tw.AddTimingTask(task) {
		task.list = nil
		if tw.LastExecutedTasks.Count(func(timingTask *TimingTask) bool {
			return timingTask.ExpirationTime == task.ExpirationTime && timingTask.Runnable == task.Runnable
		}) < 3 {
			tw.LastExecutedTasks.Push(task)

			// if task.ExpirationTime < tw.CurrentTime {
			// 	xlog.Error("task.ExpirationTime < tw.CurrentTime", reflect.TypeOf(task.Runnable), task.ExpirationTime, tw.CurrentTime)
			// }

			time_util.StashSpecificLogicTime(task.ExpirationTime)
			defer time_util.PopLogicTime()

			task.Run(task.ExpirationTime)
		} else {
			// 可能引发死循环
			xlog.Error("!!!Infinite Loop Warning!!!:", reflect.TypeOf(task.Runnable), task.Runnable, task.ExpirationTime)
		}
		return false
	}
	return true
}

// AdvanceClock
// Try to advance the clock
// 时间轮的当前时间（推进到这个时间，也就是当前定时任务的触发的时间，毫秒）  时间戳
func (tw *TimingWheel) AdvanceClock(expirationTime int64) {
	if expirationTime >= tw.CurrentTime+tw.TickTime {
		tw.CurrentTime = expirationTime - expirationTime%tw.TickTime

		// Try to advance the clock of the overflow wheel if present
		if tw.OverflowWheel != nil {
			tw.OverflowWheel.AdvanceClock(tw.CurrentTime)
		}
	}
}

// expirationTime: 当前时间戳
func (tw *TimingWheel) GetTimingTaskList(expirationTime int64) *TimingTaskList {
	if expirationTime < tw.CurrentTime+tw.TickTime { // 表示已经过期了？
		return nil
	}

	if expirationTime < tw.CurrentTime-tw.CurrentTime%tw.Interval+tw.Interval {
		// 只要是属于这个时间轮的 过期时间，那就按存放的方式来取
		virtualId := expirationTime / tw.TickTime
		bucket := tw.Buckets[virtualId%int64(tw.WheelSize)]

		if bucket.ExpirationTime != -1 {
			return bucket
		} else {
			return nil
		}
	}

	if tw.OverflowWheel != nil {
		return tw.OverflowWheel.GetTimingTaskList(expirationTime)
	}

	return nil
}

func (tw *TimingWheel) GetTaskCountDetail() (totalCount int, taskCountDetailList [][]int) {
	taskCountDetail := make([]int, tw.WheelSize)
	for i, bucket := range tw.Buckets {
		taskCountDetail[i] = bucket.TaskList.Len()
		totalCount += taskCountDetail[i]
	}
	taskCountDetailList = append(taskCountDetailList, taskCountDetail)

	if tw.OverflowWheel != nil {
		overflowTotalCount, overflowTaskCountDetailList := tw.OverflowWheel.GetTaskCountDetail()
		totalCount += overflowTotalCount
		taskCountDetailList = append(taskCountDetailList, overflowTaskCountDetailList...)
	}

	return
}

func NewTimingWheel(tickTime int64, wheelSize int, startTime int64, offerExpirationTime func(expirationTime int64)) *TimingWheel {
	if offerExpirationTime == nil {
		xlog.Error("must implement the function offerExpirationTime")
	}
	tw := &TimingWheel{
		TickTime:            tickTime,
		WheelSize:           wheelSize,
		Interval:            tickTime * int64(wheelSize),
		CurrentTime:         startTime - startTime%tickTime,
		Buckets:             make([]*TimingTaskList, wheelSize),
		LastExecutedTasks:   circularbuffer.NewCircularBuffer[*TimingTask](15),
		OfferExpirationTime: offerExpirationTime,
	}
	for i := 0; i < wheelSize; i++ {
		tw.Buckets[i] = NewTimingTaskList(tickTime)
	}
	return tw
}
