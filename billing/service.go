package billing

import (
	"context"
	"fmt"
	"log"

	"encore.app/billing/activity"
	"encore.app/billing/workflow"
	"encore.dev"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var (
	envName = encore.Meta().Environment.Name
)

//encore:service
type Service struct {
	client client.Client
	worker worker.Worker
}

func initService() (*Service, error) {
	c, err := client.Dial(client.Options{HostPort: "127.0.0.1:7233"})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %v", err)
	}
	log.Print("hehhe")

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
