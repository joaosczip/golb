package lb

import (
	"net/http"

	tg "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
)

type LoadBalancer struct {
	targetGroups []*tg.TargetGroup
}

func NewLoadBalancer(targetGroups []*tg.TargetGroup) *LoadBalancer {
	return &LoadBalancer{
		targetGroups: targetGroups,
	}
}

func (lb *LoadBalancer) ListenAndServe(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, targetGroup := range lb.targetGroups {
			err := targetGroup.Algorithm.Handle(w, r)

			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
		}
	})
	return http.ListenAndServe(addr, nil)
}
