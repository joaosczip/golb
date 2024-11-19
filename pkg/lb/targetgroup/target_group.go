package targetgroup

import alg "github.com/joaosczip/go-lb/pkg/lb/algorithms"

type TargetGroup struct {
	Targets           []*Target
	HealthCheckConfig *HealthCheckConfig
	Algorithm         alg.Algorithm
}

type NewTargetGroupParams struct {
	Targets           []*Target
	HealthCheckConfig *HealthCheckConfig
	Algorithm         alg.Algorithm
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
