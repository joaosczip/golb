package main

import (
	"log"
	"net/http"
	"os"

	alg "github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb"
	"github.com/joaosczip/go-lb/pkg/lb/targetgroup"
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
	HealthyThreshold int    `yaml:"healthy-threshold"`
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

	var targets []*targetgroup.Target
	var healthCheckConfig *targetgroup.HealthCheckConfig

	for _, tg := range config.TargetGroups {
		for _, target := range tg.Targets {
			targets = append(targets, targetgroup.NewTarget(target.Host, target.Port))
		}

		healthCheckConfig = targetgroup.NewHealthCheckConfig(
			targetgroup.HealthCheckConfigParams{
				IntervalInSec:    tg.HealthCheck.Interval,
				TimeoutInSec:     tg.HealthCheck.Timeout,
				FailureThreshold: tg.HealthCheck.FailureThreshold,
				HealthyThreshold: tg.HealthCheck.HealthyThreshold,
				Path:             tg.HealthCheck.Path,
				HttpClient:       httpClient,
			},
		)
	}

	targetGroup := targetgroup.NewTargetGroup(targetgroup.NewTargetGroupParams{
		Targets:           targets,
		HealthCheckConfig: healthCheckConfig,
		Algorithm:         alg.NewRoundRobin(targets),
	})

	lb := lb.NewLoadBalancer([]*targetgroup.TargetGroup{targetGroup})

	err = lb.ListenAndServe(":9000")

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
