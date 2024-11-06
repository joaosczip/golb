package lb

import "net/http"

type HealthCheckConfig struct {
	Interval         int
	Timeout          int
	FailureThreshold int
	Path             string
	HttpClient       *http.Client
}

type TargetGroup struct {
	Targets           []*Target
	HealthCheckConfig *HealthCheckConfig
}

type HealthCheckConfigParams struct {
	IntervalInSec    int
	TimeoutInSec     int
	FailureThreshold int
	Path             string `default:"/health"`
	HttpClient       *http.Client
}

func NewHealthCheckConfig(params HealthCheckConfigParams) *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval:         params.IntervalInSec,
		Timeout:          params.TimeoutInSec,
		FailureThreshold: params.FailureThreshold,
		Path:             params.Path,
		HttpClient:       params.HttpClient,
	}
}

func NewTargetGroup(targets []*Target, healthCheckConfig *HealthCheckConfig) *TargetGroup {
	tg := &TargetGroup{
		Targets:           targets,
		HealthCheckConfig: healthCheckConfig,
	}

	for _, target := range tg.Targets {
		go target.healthCheck(*healthCheckConfig)
	}

	return tg
}