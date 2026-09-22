package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Uptime struct {
	AliveTime       time.Duration
	AlivePercentage float64
}

type HealthEvent struct {
	Alive     bool
	Timestamp int64
}

type Metrics struct {
	TotalRequests  atomic.Int64
	HealthHistory  []HealthEvent
	StartTimestamp int64
	mu             sync.RWMutex
}

func (m *Metrics) AddHealthEvent(alive bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HealthHistory = append(m.HealthHistory, HealthEvent{
		Alive:     alive,
		Timestamp: time.Now().Unix(),
	})
}

func (m *Metrics) HealthEvents() []HealthEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]HealthEvent, len(m.HealthHistory))
	copy(result, m.HealthHistory)

	return result
}

func (m *Metrics) Uptime() *Uptime {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var uptime Uptime

	if len(m.HealthHistory) == 0 {
		uptime.AliveTime = time.Since(time.Unix(m.StartTimestamp, 0))
		uptime.AlivePercentage = 1.0
		return &uptime
	}

	for i := 0; i < len(m.HealthHistory); i += 2 {
		if i == 0 {
			deadTime := time.Unix(m.HealthHistory[i].Timestamp, 0)
			startTime := time.Unix(m.StartTimestamp, 0)

			uptime.AliveTime += deadTime.Sub(startTime)
		} else {
			aliveTime := time.Unix(m.HealthHistory[i-1].Timestamp, 0)
			deadTime := time.Unix(m.HealthHistory[i].Timestamp, 0)

			uptime.AliveTime += deadTime.Sub(aliveTime)
		}
	}

	if m.HealthHistory[len(m.HealthHistory)-1].Alive {
		lastAlive := time.Unix(
			m.HealthHistory[len(m.HealthHistory)-1].Timestamp,
			0,
		)

		uptime.AliveTime += time.Since(lastAlive)
	}

	totalTime := time.Since(time.Unix(m.StartTimestamp, 0))

	if totalTime > 0 {
		uptime.AlivePercentage = float64(uptime.AliveTime) / float64(totalTime)
	}

	return &uptime
}
