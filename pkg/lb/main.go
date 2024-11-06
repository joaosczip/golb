package lb

import "net/http"

type Algorithm interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct{
	Algorithm Algorithm
}

type Target struct {
	Host string
	Port int
}

func NewTarget(host string, port int) Target {
	return Target{
		Host: host,
		Port: port,
	}
}

func NewLoadBalancer(algorithm Algorithm) *LoadBalancer {
	return &LoadBalancer{
		Algorithm: algorithm,
	}
}

func (lb *LoadBalancer) ListenAndServe(addr string) error {
	http.HandleFunc("/", lb.Algorithm.Handle)
	return http.ListenAndServe(addr, nil)
}
