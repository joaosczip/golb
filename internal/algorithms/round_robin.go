package algorithms

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/joaosczip/go-lb/pkg/lb"
)

type roundRobin struct {
	current atomic.Int64
	targetGroup *lb.TargetGroup
}

func NewRoundRobin(targetGroup *lb.TargetGroup) *roundRobin {
	return &roundRobin{
		current: atomic.Int64{},
		targetGroup: targetGroup,
	}
}

func (r *roundRobin) Handle(w http.ResponseWriter, req *http.Request) error {
	numTargets := int64(len(r.targetGroup.Targets))
	currentIndex := r.current.Load()
	currentTarget := r.targetGroup.Targets[currentIndex]

	unhealthyTargets := 0

	for !currentTarget.IsHealthy() {
		fmt.Printf("Target %s:%d is not healthy", currentTarget.Host, currentTarget.Port)

		currentIndex = (currentIndex + 1) % numTargets
		currentTarget = r.targetGroup.Targets[currentIndex]

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