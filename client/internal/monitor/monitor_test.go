package monitor

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func Test_Monitor(t *testing.T) {
	// 1. 当前进程 CPU 使用率
	p, err := process.NewProcess(int32(getCurrentPID()))
	if err != nil {
		log.Fatal("NewProcess error:", err)
	}

	// CPU% 需要通过一段时间的采样来计算
	cpuPercentBefore, err := p.CPUPercent()
	if err != nil {
		log.Fatal("CPUPercent error:", err)
	}
	// 等待一段时间再次采样
	time.Sleep(1 * time.Second)
	cpuPercentAfter, err := p.CPUPercent()
	if err != nil {
		log.Fatal("CPUPercent error:", err)
	}

	fmt.Printf("Process CPU usage: %.2f%%\n", cpuPercentAfter-cpuPercentBefore)

	// 2. 系统总 CPU 使用率
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatal("cpu.Percent error:", err)
	}
	if len(cpuPercents) > 0 {
		fmt.Printf("Total CPU usage: %.2f%%\n", cpuPercents[0])
	}

	// 3. 当前进程内存占用
	memInfo, err := p.MemoryInfo()
	if err != nil {
		log.Fatal("MemoryInfo error:", err)
	}
	// RSS: 进程实际使用的物理内存
	// VMS: 进程分配的虚拟内存
	fmt.Printf("Process Memory (RSS): %v bytes\n", memInfo.RSS)
	fmt.Printf("Process Memory (VMS): %v bytes\n", memInfo.VMS)

	// 4. 系统总内存占用 & 5. 系统实际物理内存
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("mem.VirtualMemory error:", err)
	}
	// Total: 系统物理内存总量
	// Used: 已使用内存量
	// Free: 空闲内存量
	// UsedPercent: 使用率（百分比）
	fmt.Printf("System Memory Total: %v bytes\n", vmStat.Total)
	fmt.Printf("System Memory Used: %v bytes\n", vmStat.Used)
	fmt.Printf("System Memory Free: %v bytes\n", vmStat.Free)
	fmt.Printf("System Memory Usage: %.2f%%\n", vmStat.UsedPercent)
}

// getCurrentPID 获取当前进程 PID
func getCurrentPID() int {
	return int(os.Getpid()) // 如果你需要手动获取，也可以使用 os.Getpid()
}
