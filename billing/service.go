package billing

import (
	"context"
	"fmt"

	"encore.app/billing/activity"
	"encore.app/billing/workflow"
	"encore.dev"
	"encore.dev/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type Config struct {
	TemporalHost config.String
}

var (
	envName = encore.Meta().Environment.Name
	cfg     = config.Load[Config]()
)

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
}

func initService() (*Service, error) {
	c, err := client.Dial(client.Options{HostPort: cfg.TemporalHost()})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}

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
	return &Service{client: c, worker: w}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
