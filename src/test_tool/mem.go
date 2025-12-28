package test_tool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemMonitor 是一个用于测试期间监控内存使用的工具
type MemMonitor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMemMonitor 创建一个新的内存监控器
func NewMemMonitor() *MemMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &MemMonitor{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动内存监控，interval 为采样间隔（如 time.Second）
func (m *MemMonitor) Start(interval time.Duration) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var maxAlloc, maxSys uint64

		for {
			select {
			case <-m.ctx.Done():
				// 测试结束，打印最终和峰值
				var final runtime.MemStats
				runtime.ReadMemStats(&final)
				fmt.Printf("✅ [MemMonitor] Final - Alloc: %s, Sys: %s\n",
					humanizeBytes(final.Alloc), humanizeBytes(final.Sys))
				fmt.Printf("📈 [MemMonitor] Peak  - Alloc: %s, Sys: %s\n",
					humanizeBytes(maxAlloc), humanizeBytes(maxSys))
				return
			case <-ticker.C:
				var mstat runtime.MemStats
				runtime.ReadMemStats(&mstat)
				if mstat.Alloc > maxAlloc {
					maxAlloc = mstat.Alloc
				}
				if mstat.Sys > maxSys {
					maxSys = mstat.Sys
				}
				fmt.Printf("📊 [MemMonitor] Alloc: %s, Sys: %s\n",
					humanizeBytes(mstat.Alloc), humanizeBytes(mstat.Sys))
			}
		}
	}()
}

// Stop 停止监控并等待 goroutine 退出
func (m *MemMonitor) Stop() {
	m.cancel()
	m.wg.Wait()
}

// ———————— 辅助函数 ————————

func humanizeBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
