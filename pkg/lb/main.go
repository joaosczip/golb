package lb

import (
	"net/http"
)

type Algorithm interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

type LoadBalancer struct {
	Algorithm Algorithm
}

func NewLoadBalancer(algorithm Algorithm) *LoadBalancer {
	return &LoadBalancer{
		Algorithm: algorithm,
	}
}

func (lb *LoadBalancer) ListenAndServe(addr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := lb.Algorithm.Handle(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
	})
	return http.ListenAndServe(addr, nil)
}
