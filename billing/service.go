package billing

import (
	"context"
	"fmt"

	"encore.app/billing/activity"
	"encore.app/billing/db"
	"encore.app/billing/workflow"
	"encore.dev"
	"encore.dev/config"
	"encore.dev/storage/sqldb"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type Config struct {
	TemporalHost config.String
}

var (
	envName  = encore.Meta().Environment.Name
	cfg      = config.Load[Config]()
	dbClient = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
		Migrations: "./db/migrations",
	})
)

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
	dao    db.BillingDaoInterface
}

func initService() (*Service, error) {
	c, err := client.Dial(client.Options{HostPort: cfg.TemporalHost()})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

	var _ db.BillingDaoInterface = (*db.BillingDao)(nil)
	dao := &db.BillingDao{Db: dbClient}
	w := worker.New(c, BillingTaskQueue, worker.Options{})

	w.RegisterWorkflow(workflow.CreateBillWorkflow)
	w.RegisterActivity(activity.CreateBillActivity)
	w.RegisterActivity(activity.AddLineItemActivity)
	w.RegisterActivity(activity.FinalizeBillActivity)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %v", err)
	}
	return &Service{client: c, worker: w, dao: dao}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
