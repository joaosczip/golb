package lb

import (
	"fmt"
	"net/http"

	tg "github.com/joaosczip/go-lb/pkg/lb/targetgroup"
)

type LoadBalancer struct {
	TargetGroups []*tg.TargetGroup
	Port int
}

func NewLoadBalancer(targetGroups []*tg.TargetGroup, port int) *LoadBalancer {
	return &LoadBalancer{
		TargetGroups: targetGroups,
		Port: port,
	}
}

func (lb *LoadBalancer) ListenAndServe() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, targetGroup := range lb.TargetGroups {
			err := targetGroup.Algorithm.Handle(w, r)

			if err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", lb.Port), nil)
}
