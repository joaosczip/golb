package algorithms

import (
	"net/http/httptest"
	"sync"
	"testing"

	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


func buildLRTTargets(targets []*lb.Target) []*leastResponseTimeTarget {
	lrtTargets := make([]*leastResponseTimeTarget, len(targets))
	
	lrtTargets[0] = &leastResponseTimeTarget{
		Target: targets[0],
		avgResponseTime: 112.0,
		requestCount: 10,
		mux: sync.RWMutex{},
	}
	lrtTargets[1] = &leastResponseTimeTarget{
		Target: targets[1],
		avgResponseTime: 100.10,
		requestCount: 3,
		mux: sync.RWMutex{},
	}

	return lrtTargets
}

func TestLeastResponseTime_Handle(t *testing.T) {
	proxyFactory := &MockedProxyFactory{}
	proxy := &MockedProxy{}
	
	t.Run("Should call the target with the least avg response time", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)
		
		lrt := NewLeastResponseTime(targets, proxyFactory)
		lrt.targets = lrtTargets

		proxyFactory.On("Create", "localhost", 8081).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8081", nil)

		proxy.On("ServeHTTP", mock.Anything, r).Return()

		err := lrt.Handle(w, r)

		assert.Nil(t, err)
		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)
	})

	t.Run("Should call the next healthy target if the target with the least avg response time is not healthy", func(t *testing.T) {
		targets := getTargets()
		lrtTargets := buildLRTTargets(targets)

		lrt := NewLeastResponseTime(targets, proxyFactory)
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

		lrt := NewLeastResponseTime(targets, proxyFactory)
		lrt.targets = lrtTargets
		lrt.targets[0].Healthy = false
		lrt.targets[1].Healthy = false

		proxyFactory.On("Create", "localhost", 8081).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8081", nil)

		err := lrt.Handle(w, r)

		assert.Error(t, err, "no healthy targets available")
		proxyFactory.AssertNotCalled(t, "Create")
		proxy.AssertNotCalled(t, "ServeHTTP")
	})
}

