package billing

import (
	"context"
	"time"

	"encore.app/billing/activity"
	"encore.app/billing/db"
	billing "encore.app/billing/workflow"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"go.temporal.io/sdk/client"
)

// ==================================================================

// Response defines the standard API response for billing services.
type Response struct {
	Message string
	Error   string `json:",omitempty"`
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
	billId := uuid.New().String()
	options := client.StartWorkflowOptions{
		ID:        billId,
		TaskQueue: BillingTaskQueue,
	}
	we, err := s.client.ExecuteWorkflow(ctx, options, billing.CreateBillWorkflow, billing.CreateBillWorkflowInput{
		BillId:      billId,
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
	BillId       string          `json:"bill_id"`
	Reference    string          `json:"reference"`
	Description  string          `json:"description"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	ExchangeRate float64         `json:"exchange_rate"`
}

//encore:api public method=POST path=/bills/item
func (s *Service) AddLineItem(ctx context.Context, req *AddLineItemRequest) (*Response, error) {
	err := s.client.SignalWorkflow(ctx, req.BillId, "", activity.AddLineItemSignal, activity.AddLineItemSignalInput{
		BillId:       req.BillId,
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

// ==================================================================

type CloseBillRequest struct {
	BillId string `json:"bill_id"`
}

type BillDetailsResponse struct {
	Bill        *db.DbBill      `json:"bill"`
	LineItems   []db.DbBillItem `json:"line_items"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}

//encore:api public method=POST path=/bills/close
func (s *Service) CloseBill(ctx context.Context, req *CloseBillRequest) (*BillDetailsResponse, error) {
	// check if closable

	// trigger close
	err := s.client.SignalWorkflow(ctx, req.BillId, "", activity.CloseBillSignal, activity.CloseBillInput{
		BillId: req.BillId,
	})
	if err != nil {
		return nil, err
	}

	// wait for workflow to complete
	we := s.client.GetWorkflow(ctx, req.BillId, "")
	var result any

	err = we.Get(ctx, &result) // Blocking until the workflow is finished
	if err != nil {
		return nil, err
	}

	bill, lineItems, totalAmount, err := db.GetBillDetailsWithTotal(ctx, req.BillId)
	if err != nil {
		return nil, err
	}

	return &BillDetailsResponse{
		Bill:        bill,
		LineItems:   lineItems,
		TotalAmount: totalAmount,
	}, nil
}
