package main

import (
	"log"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb"
)

func main() {
	targets := []*lb.Target{
		lb.NewTarget("localhost", 8080),
		lb.NewTarget("localhost", 8081),
		lb.NewTarget("localhost", 8082),
	}
	targetGroup := lb.NewTargetGroup(targets, lb.NewHealthCheckConfig(
		5, 2, 2,
	))
	
	roundRobin := algorithms.NewRoundRobin(targetGroup)
	lb := lb.NewLoadBalancer(roundRobin)

	err := lb.ListenAndServe(":9000")

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}