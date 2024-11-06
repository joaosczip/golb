package lb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Target struct {
	Host    string
	Port    int
	Healthy bool
}

func NewTarget(host string, port int) *Target {
	return &Target{
		Host: host,
		Port: port,
	}
}

func (t *Target) healthCheck(hc HealthCheckConfig) {
	ticker := time.NewTicker(time.Duration(hc.Interval) * time.Second)
	defer ticker.Stop()

	healthCheckUrl := fmt.Sprintf("http://%s:%d%s", t.Host, t.Port, hc.Path)
	failures := 0

	for {
		<-ticker.C

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(hc.Timeout)*time.Second)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthCheckUrl, nil)

		if err != nil {
			log.Fatalf("could not create request: %v", err)
		}
		
		res, err := hc.HttpClient.Do(req)

		cancel()

		if err == nil {
			if res.StatusCode == http.StatusOK {
				failures = 0
				fmt.Printf("health check passed for target %s:%d\n", t.Host, t.Port)
				t.Healthy = true
				res.Body.Close()
				continue
			}
		}

		fmt.Printf("health check failed for target %s:%d: %v\n", t.Host, t.Port, err)
		failures++

		if failures > hc.FailureThreshold {
			fmt.Printf("target %s:%d is unhealthy\n", t.Host, t.Port)
			t.Healthy = false
		}
	}
}