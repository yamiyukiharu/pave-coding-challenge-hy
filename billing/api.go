package billing

import (
	"context"
	"time"

	"encore.app/billing/activity"
	"encore.app/billing/db"
	billing "encore.app/billing/workflow"
	"github.com/shopspring/decimal"

	"go.temporal.io/sdk/client"
)

// ==================================================================

// Response defines the standard API response for billing services.
type Response struct {
	Message string
	Error   string `json:",omitempty"`
}

// Bill represents a bill object in the system.
type Bill struct {
	ID          string
	Status      string
	Currency    string
	TotalAmount float64
	LineItems   []LineItem
}

// LineItem represents an individual fee in a bill.
type LineItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

// ListBillsResponse defines the response for the ListBills API.
type ListBillsResponse struct {
	Bills []*Bill `json:"bills"`
}

// ==================================================================

type CreateBillRequest struct {
	AccountId   string    `json:"account_id"`
	Currency    string    `json:"currency"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

//encore:api public method=POST path=/bills
func (s *Service) CreateBill(ctx context.Context, req *CreateBillRequest) (*Response, error) {
	ctx = db.SetDaoToContext(ctx, s.dao)
	options := client.StartWorkflowOptions{
		ID:        "create-bill-workflow",
		TaskQueue: BillingTaskQueue,
	}
	we, err := s.client.ExecuteWorkflow(ctx, options, billing.CreateBillWorkflow, billing.CreateBillWorkflowInput{
		AccountId:   req.AccountId,
		Currency:    req.Currency,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
	})
	if err != nil {
		return nil, err
	}

	msg := "Workflow started with ID: " + we.GetID()
	return &Response{Message: msg}, nil
}

// ==================================================================

type AddLineItemRequest struct {
	Reference    string          `json:"reference"`
	Description  string          `json:"description"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	ExchangeRate float64         `json:"exchange_rate"`
}

//encore:api public method=POST path=/bills/item
func (s *Service) AddLineItemSignal(ctx context.Context, req *AddLineItemRequest) (*Response, error) {
	ctx = db.SetDaoToContext(ctx, s.dao)
	err := s.client.SignalWorkflow(ctx, "haha", "", activity.AddLineItemSignal, activity.AddLineItemSignalInput{
		Reference:    req.Reference,
		Description:  req.Description,
		Amount:       req.Amount,
		Currency:     req.Currency,
		ExchangeRate: req.ExchangeRate,
	})
	if err != nil {
		return nil, err
	}

	return &Response{Message: "Line item added to workflow"}, nil
}
