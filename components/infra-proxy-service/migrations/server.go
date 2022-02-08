package migrations

import (
	"github.com/chef/automate/components/infra-proxy-service/migrations/pipeline"
	"github.com/chef/automate/components/infra-proxy-service/service"
)

type MigrationServer struct {
	Service          *service.Service
	phaseOnePipeline pipeline.PhaseOnePipleine
}

// NewMigrationServer returns an infra-proxy migration server
func NewMigrationServer(service *service.Service) *MigrationServer {
	c := pipeline.SetupPhaseOnePipeline()
	return &MigrationServer{
		Service:          service,
		phaseOnePipeline: c,
	}
}