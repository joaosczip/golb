package config

import (
	"fmt"
	"net/http"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/internal/proxy"
	"github.com/joaosczip/go-lb/pkg/lb"
	alg "github.com/joaosczip/go-lb/pkg/lb/algorithms"
	"github.com/joaosczip/go-lb/pkg/lb/targetgroup"
	"gopkg.in/yaml.v3"
)

type LBConfig struct {
	Port         int           `yaml:"port"`
	TargetGroups []TargetGroup `yaml:"target-groups"`
}

type Algorithm struct {
	Type    string         `yaml:"type"`
	Options map[string]any `yaml:"options,omitempty"`
}

type TargetGroup struct {
	Name        string      `yaml:"name"`
	Algorithm   Algorithm   `yaml:"algorithm"`
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

type ConfigLoader struct {
	path         string
	httpClient   *http.Client
	proxyFactory proxy.ProxyFactory
	fileReader   FileReader
}

func NewConfigLoader(configFilePath string, httpClient *http.Client, proxyFactory proxy.ProxyFactory, fileReader FileReader) *ConfigLoader {
	return &ConfigLoader{
		path:         configFilePath,
		httpClient:   httpClient,
		proxyFactory: proxyFactory,
		fileReader:   fileReader,
	}
}

func (c *ConfigLoader) getAlgorithm(targets []*targetgroup.Target, algConfig Algorithm) alg.Algorithm {
	if algConfig.Type == "round-robin" {
		return algorithms.NewRoundRobin(targets, c.proxyFactory)
	}
	
	return algorithms.NewLeastResponseTime(targets, c.proxyFactory, algorithms.NewLeastResponseTimeOptions{
		MaxConsecutiveRequests: int64(algConfig.Options["max-consecutive-requests"].(int)),
	})
}

func (c *ConfigLoader) Load() (*lb.LoadBalancer, error) {
	configFileData, err := c.fileReader.Read(c.path)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config LBConfig
	err = yaml.Unmarshal(configFileData, &config)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
	}

	var targetGroups []*targetgroup.TargetGroup

	for _, tg := range config.TargetGroups {
		var targets []*targetgroup.Target

		for _, target := range tg.Targets {
			targets = append(targets, targetgroup.NewTarget(target.Host, target.Port))
		}

		healthCheckConfig := targetgroup.NewHealthCheckConfig(
			targetgroup.HealthCheckConfigParams{
				IntervalInSec:    tg.HealthCheck.Interval,
				TimeoutInSec:     tg.HealthCheck.Timeout,
				FailureThreshold: tg.HealthCheck.FailureThreshold,
				HealthyThreshold: tg.HealthCheck.HealthyThreshold,
				Path:             tg.HealthCheck.Path,
				HttpClient:       c.httpClient,
			},
		)

		targetGroups = append(targetGroups, targetgroup.NewTargetGroup(targetgroup.NewTargetGroupParams{
			Targets:           targets,
			HealthCheckConfig: healthCheckConfig,
			Algorithm:         c.getAlgorithm(targets, tg.Algorithm),
		}))
	}

	return lb.NewLoadBalancer(targetGroups, config.Port), nil
}
