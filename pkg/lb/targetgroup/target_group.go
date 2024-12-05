package targetgroup

import lb "github.com/joaosczip/go-lb/pkg/lb/algorithms"

type TargetGroup struct {
	Targets           []*Target
	HealthCheckConfig *HealthCheckConfig
	Algorithm         lb.Algorithm
}

type NewTargetGroupParams struct {
	Targets           []*Target
	HealthCheckConfig *HealthCheckConfig
	Algorithm         lb.Algorithm
}

func NewTargetGroup(params NewTargetGroupParams) *TargetGroup {
	tg := &TargetGroup{
		Targets:           params.Targets,
		HealthCheckConfig: params.HealthCheckConfig,
		Algorithm:         params.Algorithm,
	}

	for _, target := range tg.Targets {
		go target.healthCheck(*tg.HealthCheckConfig)
	}

	return tg
}
