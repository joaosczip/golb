package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb"
	"gopkg.in/yaml.v3"
)

type LBConfig struct {
	Port         int           `yaml:"port"`
	TargetGroups []TargetGroup `yaml:"target-groups"`
}

type TargetGroup struct {
	Name        string      `yaml:"name"`
	Algorithm   string      `yaml:"algorithm"`
	HealthCheck HealthCheck `yaml:"health-check"`
	Targets     []Target    `yaml:"targets"`
}

type HealthCheck struct {
	Interval         int    `yaml:"interval"`
	Timeout          int    `yaml:"timeout"`
	FailureThreshold int    `yaml:"failure-threshold"`
	Path             string `yaml:"path"`
}

type Target struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func main() {
	httpClient := http.DefaultClient

	configFileData, err := os.ReadFile("lb-config.yml")

	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	var config LBConfig
	err = yaml.Unmarshal(configFileData, &config)

	if err != nil {
		log.Fatalf("could not parse config file: %v", err)
	}

	var targets []*lb.Target
	var healthCheckConfig *lb.HealthCheckConfig

	for _, tg := range config.TargetGroups {
		for _, target := range tg.Targets {
			targets = append(targets, &lb.Target{
				Host: target.Host,
				Port: target.Port,
			})
		}

		healthCheckConfig = lb.NewHealthCheckConfig(lb.HealthCheckConfigParams{
			IntervalInSec:    tg.HealthCheck.Interval,
			TimeoutInSec:     tg.HealthCheck.Timeout,
			FailureThreshold: tg.HealthCheck.FailureThreshold,
			Path:             tg.HealthCheck.Path,
			HttpClient:       httpClient,
		})
	}

	targetGroup := lb.NewTargetGroup(targets, healthCheckConfig)

	roundRobin := algorithms.NewRoundRobin(targetGroup)
	lb := lb.NewLoadBalancer(roundRobin)

	err = lb.ListenAndServe(":9000")

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
