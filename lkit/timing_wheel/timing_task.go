package timing_wheel

import "container/list"

type Runnable interface {
	Run(expirationTime int64)
}

type TimingTask struct {
	Runnable
	ExpirationTime int64         // timestamp in millisecond
	element        *list.Element // 自己本身的指针
	list           *list.List    // 自己所属的 list ==> TimingTaskList.TaskList
}

func NewTimingTask(runnable Runnable, expirationTime int64) *TimingTask {
	return &TimingTask{
		Runnable:       runnable,
		ExpirationTime: expirationTime,
	}
}

func (m *TimingTask) Remove() {
	if m.list == nil {
		return
	}
	m.list.Remove(m.element)
}

type TimingTaskList struct {
	ExpirationTime int64      // -1 表示没有过期时间（里面没有任务）
	Duration       int64      // 任务时间间隔小于 这个 Duration 的，都放在这个 list 里面
	TaskList       *list.List // 存放 TimingTask
}

func (t *TimingTaskList) Add(task *TimingTask) {
	task.element = t.TaskList.PushBack(task)
	task.list = t.TaskList
}

func (t *TimingTaskList) Remove(task *TimingTask) {
	t.TaskList.Remove(task.element)
}

func (t *TimingTaskList) Flush(reinsert func(task *TimingTask)) {
	for t.TaskList.Len() > 0 {
		elem := t.TaskList.Front()
		t.TaskList.Remove(elem)
		reinsert(elem.Value.(*TimingTask))
	}
	t.TaskList.Init()
	t.SetExpirationTime(-1)
}

func (t *TimingTaskList) SetExpirationTime(expirationTime int64) bool {
	if t.ExpirationTime != expirationTime {
		t.ExpirationTime = expirationTime
		return true
	}
	return false
}

func (t *TimingTaskList) ContainsExpirationTime(expirationTime int64) bool {
	if t.ExpirationTime == -1 {
		return false
	}
	if t.ExpirationTime <= expirationTime && expirationTime < t.ExpirationTime+t.Duration {
		return true
	}
	return false
}

func NewTimingTaskList(duration int64) *TimingTaskList {
	return &TimingTaskList{
		ExpirationTime: -1,
		Duration:       duration,
		TaskList:       list.New(),
	}
}
