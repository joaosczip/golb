package main

import (
	"log"
	"net/http"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb"
)

func main() {
	httpClient := http.DefaultClient

	targets := []*lb.Target{
		lb.NewTarget("localhost", 8080),
		lb.NewTarget("localhost", 8081),
		lb.NewTarget("localhost", 8082),
	}
	targetGroup := lb.NewTargetGroup(targets, lb.NewHealthCheckConfig(
		lb.HealthCheckConfigParams{
			IntervalInSec: 2,
			TimeoutInSec: 2,
			FailureThreshold: 3,
			Path: "/health",
			HttpClient: httpClient,
		},
	))

	roundRobin := algorithms.NewRoundRobin(targetGroup)
	lb := lb.NewLoadBalancer(roundRobin)

	err := lb.ListenAndServe(":9000")

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}