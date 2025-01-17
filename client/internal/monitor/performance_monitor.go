package monitor

import (
	"runtime"
	"time"
)

type PerformanceMonitor struct {
	startTime      time.Time
	cpuStats       runtime.MemStats
	lastStatTime   time.Time
	networkLatency time.Duration
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime:    time.Now(),
		lastStatTime: time.Now(),
	}
}

func (pm *PerformanceMonitor) GetCPUUsage() float64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	now := time.Now()
	duration := now.Sub(pm.lastStatTime).Seconds()

	// 计算CPU使用率（这里使用GC时间作为近似值）
	cpuUsage := float64(stats.PauseTotalNs-pm.cpuStats.PauseTotalNs) / (duration * 1e9) * 100

	pm.cpuStats = stats
	pm.lastStatTime = now

	return cpuUsage
}

func (pm *PerformanceMonitor) GetMemoryUsage() uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

func (pm *PerformanceMonitor) GetNetworkLatency() time.Duration {
	return pm.networkLatency
}

func (pm *PerformanceMonitor) SetNetworkLatency(latency time.Duration) {
	pm.networkLatency = latency
}

func (pm *PerformanceMonitor) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

func (pm *PerformanceMonitor) GetUptime() time.Duration {
	return time.Since(pm.startTime)
}
