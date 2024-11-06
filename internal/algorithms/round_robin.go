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
	current int64
	targetGroup *lb.TargetGroup
}

func NewRoundRobin(targetGroup *lb.TargetGroup) *roundRobin {
	return &roundRobin{
		current: 0,
		targetGroup: targetGroup,
	}
}

func (r *roundRobin) Handle(w http.ResponseWriter, req *http.Request) error {
	currentIndex := atomic.LoadInt64(&r.current)
	numTargets := int64(len(r.targetGroup.Targets))
	currentTarget := r.targetGroup.Targets[currentIndex]

	if !currentTarget.Healthy {
		fmt.Printf("Target %s:%d is not healthy", currentTarget.Host, currentTarget.Port)

		for i := 0; i < int(numTargets); i++ {
			currentIndex = (currentIndex + 1) % numTargets
			currentTarget = r.targetGroup.Targets[currentIndex]

			if currentTarget.Healthy {
				atomic.StoreInt64(&r.current, currentIndex)
				break
			}
		}
		
		return errors.New("no healthy targets available")
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", currentTarget.Host, currentTarget.Port),
	})

	if currentIndex == numTargets-1 {
		atomic.StoreInt64(&r.current, 0)
	} else {
		atomic.AddInt64(&r.current, 1)
	}

	proxy.ServeHTTP(w, req)

	return nil
}