package signal

import (
	"fmt"
	"sync"
	"time"

	"fibo-monitor/indicator"

	"go.uber.org/zap"
)

type Filter struct {
	dedupWindow time.Duration
	// cache: key -> timestamp
	lastSignalTime map[string]time.Time
	mu             sync.Mutex
	logger         *zap.Logger
}

func NewFilter(dedupWindow time.Duration, logger *zap.Logger) *Filter {
	return &Filter{
		dedupWindow:    dedupWindow,
		lastSignalTime: make(map[string]time.Time),
		logger:         logger,
	}
}

func (f *Filter) Run(inChan <-chan Signal) <-chan Signal {
	outChan := make(chan Signal, 100)

	go func() {
		defer close(outChan)
		for sig := range inChan {
			if f.shouldProcess(sig) {
				outChan <- sig
			}
		}
	}()

	return outChan
}

func (f *Filter) shouldProcess(sig Signal) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := fmt.Sprintf("%s-%s-%d", sig.Symbol, sig.Interval, sig.Type)
	lastTime, ok := f.lastSignalTime[key]

	if ok && time.Since(lastTime) < f.dedupWindow {
		// Duplicate signal
		return false
	}

	// Update last signal time
	f.lastSignalTime[key] = time.Now()
	
	// Also, if we want to ensure we don't spam, maybe we should check if the *opposite* signal happened recently? 
	// But the simple deduplication window per signal type is what's requested.
	
	return true
}

// String representation for Signal Type for the key
func (s Signal) String() string {
	if s.Type == indicator.GoldenCross {
		return "GOLDEN"
	}
	return "DEATH"
}
