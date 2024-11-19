package targetgroup

import "net/http"

type HealthCheckConfig struct {
	Interval         int
	Timeout          int
	FailureThreshold int
	Path             string
	HttpClient       *http.Client
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
