package lb

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	lb "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
)

type roundRobin struct {
	current atomic.Int64
	targets []*lb.Target
}

func NewRoundRobin(targets []*lb.Target) *roundRobin {
	return &roundRobin{
		current: atomic.Int64{},
		targets: targets,
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
			return errors.New("no healthy targets available")
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", currentTarget.Host, currentTarget.Port),
	})

	proxy.ServeHTTP(w, req)

	r.current.Store((currentIndex + 1) % numTargets)

	return nil
}
