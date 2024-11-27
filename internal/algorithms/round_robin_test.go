package algorithms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joaosczip/go-lb/internal/proxy"
	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedProxyFactory struct {
	mock.Mock
}

func (m *MockedProxyFactory) Create(host string, port int) proxy.Proxy {
	args := m.Called(host, port)
	return args.Get(0).(proxy.Proxy)
}

type MockedProxy struct {
	mock.Mock
}

func (m *MockedProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func getTargets() []*lb.Target {
	return []*lb.Target{
		{
			Host: "localhost",
			Port: 8080,
			Healthy: true,
		},
		{
			Host: "localhost",
			Port: 8081,
			Healthy: true,
		},
	}
}

func TestRoundRobin_Handle(t *testing.T) {
	proxyFactory := &MockedProxyFactory{}
	proxy := &MockedProxy{}

	t.Run("Should call the current target when it's healthy", func(t *testing.T) {
		targets := getTargets()
		rr := NewRoundRobin(targets, proxyFactory)
		
		proxyFactory.On("Create", "localhost", 8080).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)

		proxy.On("ServeHTTP", w, r).Return()

		err := rr.Handle(w, r)

		assert.Nil(t, err)
		assert.Equal(t, rr.current.Load(), int64(1))

		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)
	})

	t.Run("Should call the next target when the first one is unhealthy", func(t *testing.T) {
		targets := getTargets()
		targets[0].Healthy = false

		rr := NewRoundRobin(targets, proxyFactory)
		
		proxyFactory.On("Create", "localhost", 8081).Return(proxy)

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8081", nil)

		proxy.On("ServeHTTP", w, r).Return()

		err := rr.Handle(w, r)

		assert.Nil(t, err)
		assert.Equal(t, rr.current.Load(), int64(0))

		proxyFactory.AssertExpectations(t)
		proxy.AssertExpectations(t)
	})

	t.Run("Should return an error when all targets are unhealthy", func(t *testing.T) {
		targets := getTargets()
		targets[0].Healthy = false
		targets[1].Healthy = false

		rr := NewRoundRobin(targets, proxyFactory)
		
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)

		err := rr.Handle(w, r)

		assert.NotNil(t, err)
		assert.Error(t, err, "no healthy targets available")
		
		proxyFactory.AssertNotCalled(t, "Create")
		proxy.AssertNotCalled(t, "ServeHTTP")
	})
}