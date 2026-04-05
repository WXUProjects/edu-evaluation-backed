package cache

import (
	"sync"
	"time"
)

// State 熔断器状态
type State int

const (
	Closed   State = iota // 正常
	Open                  // 熔断开启
	HalfOpen              // 半开（试探恢复）
)

const (
	maxFailures    = 5          // 连续失败阈值
	cooldown       = 30 * time.Second // 冷却时间
)

// HealthChecker Redis 健康检查器 + 熔断器
type HealthChecker struct {
	mu         sync.Mutex
	state      State
	failures   int
	lastFailAt time.Time
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{state: Closed}
}

// IsAvailable 检查 Redis 是否可用
func (h *HealthChecker) IsAvailable() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch h.state {
	case Closed:
		return true
	case Open:
		// 冷却后进入半开
		if time.Since(h.lastFailAt) >= cooldown {
			h.state = HalfOpen
			return true
		}
		return false
	case HalfOpen:
		return true
	}
	return true
}

// RecordFailure 记录一次 Redis 失败
func (h *HealthChecker) RecordFailure() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.failures++
	h.lastFailAt = time.Now()

	if h.state == HalfOpen {
		// 半开状态下失败，重新熔断
		h.state = Open
		h.failures = 0
		return
	}

	if h.failures >= maxFailures {
		h.state = Open
		h.failures = 0
	}
}

// RecordSuccess 记录一次成功（用于半开状态恢复）
func (h *HealthChecker) RecordSuccess() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.failures = 0
	h.state = Closed
}
