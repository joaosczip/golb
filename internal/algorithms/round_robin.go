package algorithms

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/joaosczip/go-lb/internal/proxy"
	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"

	errs "github.com/joaosczip/go-lb/internal/errors"
)

type roundRobin struct {
	current      atomic.Int64
	targets      []*lb.Target
	proxyFactory proxy.ProxyFactory
}

func NewRoundRobin(targets []*lb.Target, proxyFactory proxy.ProxyFactory) *roundRobin {
	return &roundRobin{
		current:      atomic.Int64{},
		targets:      targets,
		proxyFactory: proxyFactory,
	}
}

func (r *roundRobin) Handle(w http.ResponseWriter, req *http.Request) error {
	numTargets := int64(len(r.targets))
	currentIndex := r.current.Load()
	currentTarget := r.targets[currentIndex]

	unhealthyTargets := 0

	for !currentTarget.IsHealthy() {
		fmt.Printf("Target %s:%d is unhealthy", currentTarget.Host, currentTarget.Port)

		currentIndex = (currentIndex + 1) % numTargets
		currentTarget = r.targets[currentIndex]

		unhealthyTargets++

		if int64(unhealthyTargets) == numTargets {
			return errs.ErrNoHealthyTargets
		}
	}

	proxy := r.proxyFactory.Create(currentTarget.Host, currentTarget.Port)
	proxy.ServeHTTP(w, req)

	r.current.Store((currentIndex + 1) % numTargets)

	return nil
}
