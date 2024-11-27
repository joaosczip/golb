package algorithms

import (
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/joaosczip/go-lb/internal/proxy"
	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
)

type timedResponseWriter struct {
	http.ResponseWriter
	startTime time.Time
	endTime time.Time
	statusCode int
}

func newTimedResponseWriter(w http.ResponseWriter) *timedResponseWriter {
	return &timedResponseWriter{
		ResponseWriter: w,
		startTime: time.Now(),
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
	avgResponseTime float64
	requestCount int
	mux sync.RWMutex
}

func newLeastResponseTimeTarget(target *lb.Target) *leastResponseTimeTarget {
	return &leastResponseTimeTarget{
		Target: target,
		mux: sync.RWMutex{},
	}
}

func (l *leastResponseTimeTarget) getAvgResponseTime() float64 {
	l.mux.RLock()
	defer l.mux.RUnlock()

	return l.avgResponseTime
}


type leastResponseTime struct {
	proxyFactory proxy.ProxyFactory
	targets []*leastResponseTimeTarget
	mux sync.RWMutex
}

func NewLeastResponseTime(targets []*lb.Target, proxyFactory proxy.ProxyFactory) *leastResponseTime {
	lrtTargets := make([]*leastResponseTimeTarget, len(targets))

	for i, target := range targets {
		lrtTargets[i] = newLeastResponseTimeTarget(target)
	}
	
	return &leastResponseTime{
		targets: lrtTargets,
		proxyFactory: proxyFactory,
		mux: sync.RWMutex{},
	}
}

func (l *leastResponseTime) targetsSortedByAvgResponseTime() []*leastResponseTimeTarget {
	l.mux.RLock()
	defer l.mux.RUnlock()
	
	targetsCopy := make([]*leastResponseTimeTarget, len(l.targets))
	copy(targetsCopy, l.targets)

	sort.Slice(targetsCopy, func(i, j int) bool {
		return targetsCopy[i].getAvgResponseTime() < targetsCopy[j].getAvgResponseTime()
	})

	return targetsCopy
}

func (l *leastResponseTime) Handle(w http.ResponseWriter, req *http.Request) error {
	sortedTargets := l.targetsSortedByAvgResponseTime()
	currentTarget := sortedTargets[0]

	nextIdx := 1
	for !currentTarget.IsHealthy() {
		if nextIdx == len(sortedTargets) {
			return errors.New("no healthy targets available")
		}

		currentTarget = sortedTargets[nextIdx]
		nextIdx++
	}

	timedRW := newTimedResponseWriter(w)

	proxy := l.proxyFactory.Create(currentTarget.Host, currentTarget.Port)
	proxy.ServeHTTP(timedRW, req)

	responseTime := timedRW.endTime.Sub(timedRW.startTime)

	currentTarget.mux.Lock()
	currentTarget.requestCount++
	currentTarget.avgResponseTime = (currentTarget.avgResponseTime + float64(responseTime.Nanoseconds())) / float64(currentTarget.requestCount)
	currentTarget.mux.Unlock()

	return nil
}