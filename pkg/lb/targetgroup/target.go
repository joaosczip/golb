package targetgroup

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Target struct {
	Host    string
	Port    int
	Healthy bool
	mux     sync.RWMutex
}

func NewTarget(host string, port int) *Target {
	return &Target{
		Host: host,
		Port: port,
	}
}

func (t *Target) setHealthy(healthy bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.Healthy = healthy
}

func (t *Target) IsHealthy() bool {
	t.mux.RLock()
	defer t.mux.RUnlock()
	return t.Healthy
}

func (t *Target) healthCheck(hc HealthCheckConfig) {
	ticker := time.NewTicker(time.Duration(hc.Interval) * time.Second)
	defer ticker.Stop()

	healthCheckUrl := fmt.Sprintf("http://%s:%d%s", t.Host, t.Port, hc.Path)
	failures := 0
	succeeded := 0

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
				succeeded++
				fmt.Printf("health check passed for target %s:%d, %d, %d\n", t.Host, t.Port, succeeded, hc.HealthyThreshold)

				if !t.IsHealthy() && succeeded >= hc.HealthyThreshold {
					fmt.Printf("target %s:%d is healthy\n", t.Host, t.Port)
					t.setHealthy(true)
				}

				res.Body.Close()
				continue
			}
		}

		fmt.Printf("health check failed for target %s:%d: %v\n", t.Host, t.Port, err)
		failures++
		succeeded = 0

		if failures > hc.FailureThreshold {
			fmt.Printf("target %s:%d is unhealthy\n", t.Host, t.Port)
			t.setHealthy(false)
		}
	}
}
