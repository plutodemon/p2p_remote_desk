package window

import (
	"fmt"
	"time"

	"p2p_remote_desk/client/internal/monitor"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// PerformanceStatsView 性能监控视图
type PerformanceStatsView struct {
	container *fyne.Container

	// 性能指标显示
	// 运行时间
	uptimeLabel *widget.Label
	// 帧率
	fpsLabel *widget.Label
	// CPU使用率
	cpuUsageLabel *widget.Label
	// 内存使用量
	memoryUsageLabel *widget.Label
	// 网络延迟
	latencyLabel *widget.Label
	// Goroutine数量
	goroutineLabel *widget.Label

	// 性能监控器
	monitor *monitor.PerformanceMonitor

	// 更新定时器
	updateTimer *time.Timer
	// 是否可见
	isVisible bool
}

func newPerformanceStatsView() *PerformanceStatsView {
	pv := &PerformanceStatsView{
		monitor: monitor.NewPerformanceMonitor(), // 设置性能监控的更新间隔为1秒
	}

	pv.initComponents()
	pv.startUpdateTimer()

	return pv
}

// initComponents 初始化组件
func (pv *PerformanceStatsView) initComponents() {
	// 创建性能指标标签
	pv.uptimeLabel = widget.NewLabel("运行时间: 00:00:00")
	pv.fpsLabel = widget.NewLabel("FPS: 0.00")
	pv.cpuUsageLabel = widget.NewLabel("CPU: 0.00%")
	pv.memoryUsageLabel = widget.NewLabel("内存: 0.00 MB")
	pv.latencyLabel = widget.NewLabel("延迟: 0ms")
	pv.goroutineLabel = widget.NewLabel("Goroutines: 0")

	// 创建容器
	pv.container = container.NewVBox(
		pv.uptimeLabel,
		pv.fpsLabel,
		pv.cpuUsageLabel,
		pv.memoryUsageLabel,
		pv.latencyLabel,
		pv.goroutineLabel,
	)
}

// startUpdateTimer 启动更新定时器
func (pv *PerformanceStatsView) startUpdateTimer() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !pv.isVisible {
				return
			}
			pv.updateStats()
		}
	}()
}

// updateStats 更新性能指标
func (pv *PerformanceStatsView) updateStats() {
	if !pv.isVisible {
		return
	}

	// 更新运行时间
	uptime := pv.monitor.GetUptime()
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	pv.uptimeLabel.SetText(fmt.Sprintf("运行时间: %02d:%02d:%02d", hours, minutes, seconds))

	// 更新CPU使用率
	cpuUsage := pv.monitor.GetCPUUsage()
	pv.cpuUsageLabel.SetText(fmt.Sprintf("CPU: %.2f%%", cpuUsage))

	// 更新内存使用量
	memoryUsage := float64(pv.monitor.GetMemoryUsage()) / 1024 / 1024 // 转换为MB
	pv.memoryUsageLabel.SetText(fmt.Sprintf("内存: %.2f MB", memoryUsage))

	// 更新网络延迟
	latency := pv.monitor.GetNetworkLatency()
	pv.latencyLabel.SetText(fmt.Sprintf("延迟: %dms", latency.Milliseconds()))

	// 更新goroutine数量
	goroutines := pv.monitor.GetGoroutineCount()
	pv.goroutineLabel.SetText(fmt.Sprintf("Goroutines: %d", goroutines))
}

// GetContainer 获取容器
func (pv *PerformanceStatsView) GetContainer() *fyne.Container {
	return pv.container
}

// Show 显示性能监控
func (pv *PerformanceStatsView) Show() {
	pv.isVisible = true
	pv.startUpdateTimer()
	pv.container.Show()
}

// Hide 隐藏性能监控
func (pv *PerformanceStatsView) Hide() {
	pv.isVisible = false
	pv.container.Hide()
}

// Cleanup 清理资源
func (pv *PerformanceStatsView) Cleanup() {
	pv.isVisible = false
	if pv.updateTimer != nil {
		pv.updateTimer.Stop()
	}
}
