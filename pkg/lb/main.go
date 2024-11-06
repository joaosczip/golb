package lb

import (
	"net/http"
)

type Algorithm interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct{
	Algorithm Algorithm
}

type HealthCheckConfig struct {
	Interval int
	Timeout int
	FailureThreshold int
}

type TargetGroup struct {
	Targets []*Target
	HealthCheckConfig *HealthCheckConfig
}

type Target struct {
	Host string
	Port int
	Healthy bool
}

func NewTarget(host string, port int) *Target {
	return &Target{
		Host: host,
		Port: port,
	}
}

func NewHealthCheckConfig(intervalInSec int, timeoutInSec int, failureThreshold int) *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval: intervalInSec,
		Timeout: timeoutInSec,
		FailureThreshold: failureThreshold,
	}
}

func NewTargetGroup(targets []*Target, healthCheckConfig *HealthCheckConfig) *TargetGroup {
	return &TargetGroup{
		Targets: targets,
		HealthCheckConfig: healthCheckConfig,
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
