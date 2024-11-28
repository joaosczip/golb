package algorithms

import (
	"errors"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joaosczip/go-lb/internal/proxy"
	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
)

type timedResponseWriter struct {
	http.ResponseWriter
	startTime  time.Time
	endTime    time.Time
	statusCode int
}

func newTimedResponseWriter(w http.ResponseWriter) *timedResponseWriter {
	return &timedResponseWriter{
		ResponseWriter: w,
		startTime:      time.Now(),
	}
}

func (t *timedResponseWriter) WriteHeader(statusCode int) {
	t.statusCode = statusCode
	t.ResponseWriter.WriteHeader(statusCode)
}

func (t *timedResponseWriter) Write(b []byte) (int, error) {
	if t.statusCode == 0 {
		t.statusCode = http.StatusOK
	}

	t.endTime = time.Now()
	return t.ResponseWriter.Write(b)
}

type leastResponseTimeTarget struct {
	*lb.Target
	avgResponseTime     atomic.Int64
	requestCount        atomic.Int64
	consecutiveRequests atomic.Int64
}

func newLeastResponseTimeTarget(target *lb.Target) *leastResponseTimeTarget {
	return &leastResponseTimeTarget{
		Target: target,
	}
}

func (l *leastResponseTimeTarget) setAvgResponseTime(responseTime time.Duration) {
	l.requestCount.Add(1)
	l.avgResponseTime.Store(
		(l.avgResponseTime.Load() + responseTime.Nanoseconds()) / l.requestCount.Load(),
	)
}

type leastResponseTime struct {
	proxyFactory           proxy.ProxyFactory
	targets                []*leastResponseTimeTarget
	maxConsecutiveRequests int64
	requestsCount          atomic.Int64
	mux                    sync.RWMutex
}

func NewLeastResponseTime(targets []*lb.Target, proxyFactory proxy.ProxyFactory, maxConsecutiveRequests int64) *leastResponseTime {
	lrtTargets := make([]*leastResponseTimeTarget, len(targets))

	for i, target := range targets {
		lrtTargets[i] = newLeastResponseTimeTarget(target)
	}

	return &leastResponseTime{
		targets:                lrtTargets,
		proxyFactory:           proxyFactory,
		maxConsecutiveRequests: maxConsecutiveRequests,
		requestsCount:          atomic.Int64{},
		mux:                    sync.RWMutex{},
	}
}

func (l *leastResponseTime) targetsSortedByAvgResponseTime() []*leastResponseTimeTarget {
	l.mux.RLock()
	defer l.mux.RUnlock()

	targetsCopy := make([]*leastResponseTimeTarget, len(l.targets))
	copy(targetsCopy, l.targets)

	sort.Slice(targetsCopy, func(i, j int) bool {
		return targetsCopy[i].avgResponseTime.Load() < targetsCopy[j].avgResponseTime.Load()
	})

	return targetsCopy
}

func (l *leastResponseTime) Handle(w http.ResponseWriter, req *http.Request) error {
	sortedTargets := l.targets
	currentTarget := sortedTargets[0]

	if l.requestsCount.Load() > 0 {
		sortedTargets := l.targetsSortedByAvgResponseTime()
		currentTarget = sortedTargets[0]
	}

	nextIdx := 1
	for !currentTarget.IsHealthy() || currentTarget.consecutiveRequests.Load() > l.maxConsecutiveRequests {
		if nextIdx == len(sortedTargets) {
			return errors.New("no healthy targets available")
		}

		currentTarget.consecutiveRequests.Store(0)
		currentTarget = sortedTargets[nextIdx]
		nextIdx++
	}

	timedRW := newTimedResponseWriter(w)

	proxy := l.proxyFactory.Create(currentTarget.Host, currentTarget.Port)
	proxy.ServeHTTP(timedRW, req)

	responseTime := timedRW.endTime.Sub(timedRW.startTime)

	currentTarget.consecutiveRequests.Add(1)
	currentTarget.setAvgResponseTime(responseTime)

	l.requestsCount.Add(1)

	return nil
}
