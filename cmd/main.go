package main

import (
	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb"
)

func main() {
	targets := []lb.Target{
		lb.NewTarget("localhost", 8080),
		lb.NewTarget("localhost", 8081),
		lb.NewTarget("localhost", 8082),
	}
	roundRobin := algorithms.NewRoundRobin(targets)
	lb := lb.NewLoadBalancer(roundRobin)

	lb.ListenAndServe(":9000")
}