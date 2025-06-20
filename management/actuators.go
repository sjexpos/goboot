package management

import (
	"context"
	"fmt"
	"net/http"

	healthHttp "github.com/hellofresh/health-go/v5/checks/http"
	healthMySql "github.com/hellofresh/health-go/v5/checks/mysql"
	healthPostgres "github.com/hellofresh/health-go/v5/checks/postgres"
	healthRedis "github.com/hellofresh/health-go/v5/checks/redis"
	"gitlab.com/mikeyGlitz/gohealth/pkg/actuator"
	gohealth "gitlab.com/mikeyGlitz/gohealth/pkg/health"
)

type HealthGoWrapper struct {
	Name    string
	Checker func(ctx context.Context) error
}

func (wrapper *HealthGoWrapper) CheckHealth() (result gohealth.HealthCheckResult) {
	result.Service = wrapper.Name
	result.Status = gohealth.UP
	ctx := context.Background()
	err := wrapper.Checker(ctx)
	if err != nil {
		result.Status = gohealth.DOWN
		details := make(map[string]string)
		details["error"] = err.Error()
		result.Details = details
	}
	return
}

func NewActuators() http.HandlerFunc {

	config := &actuator.Config{
		Endpoints: []actuator.Endpoint{
			actuator.INFO,
			actuator.THREADDUMP,
			actuator.HEALTH,
			actuator.METRICS,
		},
		HealthCheckers: []gohealth.HealthChecker{
			&gohealth.PingChecker{},
			&gohealth.DiskStatusChecker{},
			&HealthGoWrapper{
				Name: "rabbit-aliveness-check",
				Checker: healthHttp.New(healthHttp.Config{
					URL: `http://guest:guest@0.0.0.0:32780/api/aliveness-test/%2f`,
				}),
			},
			&HealthGoWrapper{
				Name: "http-check",
				Checker: healthHttp.New(healthHttp.Config{
					URL: `http://example.com`,
				}),
			},
			&HealthGoWrapper{
				Name:    "some-custom-check-success",
				Checker: func(ctx context.Context) error { return nil },
			},
			&HealthGoWrapper{
				Name: "mysql",
				Checker: healthMySql.New(healthMySql.Config{
					DSN: `test:test@tcp(0.0.0.0:32778)/test?charset=utf8`,
				}),
			},
			&HealthGoWrapper{
				Name: "postgres",
				Checker: healthPostgres.New(healthPostgres.Config{
					DSN: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", "users_service", "1234", "localhost", "5432", "ecomm_users"),
				}),
			},
			&HealthGoWrapper{
				Name: "redis",
				Checker: healthRedis.New(healthRedis.Config{
					DSN: `redis://localhost:6379/`,
				}),
			},
		},
	}
	return actuator.GetHandler(config)
}
