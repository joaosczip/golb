package config

import (
	"fmt"
	"net/http"

	"github.com/joaosczip/go-lb/internal/algorithms"
	"github.com/joaosczip/go-lb/internal/proxy"
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

type ConfigLoader struct {
	path string
	httpClient *http.Client
	proxyFactory proxy.ProxyFactory
	fileReader FileReader
}

func NewConfigLoader(configFilePath string, httpClient *http.Client, proxyFactory proxy.ProxyFactory, fileReader FileReader) *ConfigLoader {
	return &ConfigLoader{
		path: configFilePath,
		httpClient: httpClient,
		proxyFactory: proxyFactory,
		fileReader: fileReader,
	}
}

func (c *ConfigLoader) Load() ([]*targetgroup.TargetGroup, error) {
	configFileData, err := c.fileReader.Read(c.path)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config LBConfig
	err = yaml.Unmarshal(configFileData, &config)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
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
				HttpClient:       c.httpClient,
			},
		)
	}

	targetGroup := targetgroup.NewTargetGroup(targetgroup.NewTargetGroupParams{
		Targets:           targets,
		HealthCheckConfig: healthCheckConfig,
		Algorithm:         algorithms.NewRoundRobin(targets, c.proxyFactory),
	})

	return []*targetgroup.TargetGroup{targetGroup}, nil
}