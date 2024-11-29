package algorithms

import (
	"net/http/httptest"
	"testing"

	"sync/atomic"

	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	errs "github.com/joaosczip/go-lb/internal/errors"
)

func buildLRTTargets(targets []*lb.Target) []*leastResponseTimeTarget {
	lrtTargets := make([]*leastResponseTimeTarget, len(targets))

	lrtTargets[0] = &leastResponseTimeTarget{
		Target:          targets[0],
		avgResponseTime: atomic.Int64{},
		requestCount:    atomic.Int64{},
	}
	lrtTargets[0].avgResponseTime.Store(112)
	lrtTargets[0].requestCount.Store(10)

	lrtTargets[1] = &leastResponseTimeTarget{
		Target:          targets[1],
		avgResponseTime: atomic.Int64{},
		requestCount:    atomic.Int64{},
	}
	lrtTargets[1].avgResponseTime.Store(100)
	lrtTargets[1].requestCount.Store(3)

	return lrtTargets
}

func TestLeastResponseTime_Handle(t *testing.T) {
	proxyFactory := &MockedProxyFactory{}
	proxy := &MockedProxy{}
	lrtOptions := NewLeastResponseTimeOptions{
		maxConsecutiveRequests: int64(10),
	}

	t.Run("Should call the target with the least avg response time", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)

		lrt := NewLeastResponseTime(targets, proxyFactory, lrtOptions)
		lrt.requestsCount.Store(10)
		lrt.targets = lrtTargets

		proxyFactory.On("Create", "localhost", 8081).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8081", nil)

		proxy.On("ServeHTTP", mock.Anything, r).Return()

		err := lrt.Handle(w, r)

		assert.Nil(t, err)
		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)

		assert.Equal(t, lrtTargets[1].consecutiveRequests.Load(), int64(1))
	})

	t.Run("Should call the next healthy target if the target with the least avg response time is not healthy", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)

		lrt := NewLeastResponseTime(targets, proxyFactory, lrtOptions)
		lrt.targets = lrtTargets
		lrt.targets[1].Healthy = false

		proxyFactory.On("Create", "localhost", 8080).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)

		proxy.On("ServeHTTP", mock.Anything, r).Return()

		err := lrt.Handle(w, r)

		assert.Nil(t, err)
		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)
	})

	t.Run("Should return error if there are no healthy targets", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)

		lrt := NewLeastResponseTime(targets, proxyFactory, lrtOptions)
		lrt.targets = lrtTargets
		lrt.targets[0].Healthy = false
		lrt.targets[1].Healthy = false

		proxyFactory.On("Create", "localhost", 8081).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8081", nil)

		err := lrt.Handle(w, r)

		assert.ErrorIs(t, err, errs.ErrNoHealthyTargets)
		proxyFactory.AssertNotCalled(t, "Create")
		proxy.AssertNotCalled(t, "ServeHTTP")
	})

	t.Run("Should call the next healthy target when the current target has received more consecutive requests than the allowed consecutive calls", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)
		lrtTargets[1].consecutiveRequests.Store(11)

		lrt := NewLeastResponseTime(targets, proxyFactory, lrtOptions)
		lrt.targets = lrtTargets

		proxyFactory.On("Create", "localhost", 8080).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)

		proxy.On("ServeHTTP", mock.Anything, r).Return()

		err := lrt.Handle(w, r)

		assert.Nil(t, err)
		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)
	})

	t.Run("Should call the first target in the list when the overral request count is 0", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)

		for _, target := range lrtTargets {
			target.requestCount.Store(0)
			target.avgResponseTime.Store(0)
		}

		lrt := NewLeastResponseTime(targets, proxyFactory, lrtOptions)
		lrt.requestsCount.Store(0)
		lrt.targets = lrtTargets

		proxyFactory.On("Create", "localhost", 8080).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)

		proxy.On("ServeHTTP", mock.Anything, r).Return()

		err := lrt.Handle(w, r)

		assert.Nil(t, err)
		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)

		assert.Equal(t, lrt.requestsCount.Load(), int64(1))
		assert.Equal(t, lrt.targets[0].requestCount.Load(), int64(1))
		assert.Equal(t, lrt.targets[1].requestCount.Load(), int64(0))
	})
}
