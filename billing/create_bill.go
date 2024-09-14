package billing

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/workflow"
)

type CreateBillWorkflowInput struct {
	AccountId   string
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type AddLineItemSignalInput struct {
	Reference    string
	Description  string
	Amount       decimal.Decimal
	Currency     string
	ExchangeRate float64
}

type WorkflowResult struct {
	BillID int64
}

func CreateBillWorkflow(ctx workflow.Context, input CreateBillWorkflowInput) (*WorkflowResult, error) {

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: input.PeriodEnd.Sub(time.Now().Add(time.Hour * 24)),
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var billID int64
	err := workflow.ExecuteActivity(ctx, CreateBillActivity, input.AccountId, input.Currency, input.PeriodStart, input.PeriodEnd).Get(ctx, &billID)
	if err != nil {
		return nil, err
	}

	lineItemSignalCh := workflow.GetSignalChannel(ctx, "AddLineItem")

	durationUntilEnd := input.PeriodEnd.Sub(workflow.Now(ctx))

	timerFuture := workflow.NewTimer(ctx, durationUntilEnd)

	for {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(lineItemSignalCh, func(c workflow.ReceiveChannel, more bool) {
			var lineItemInput AddLineItemSignalInput
			c.Receive(ctx, &lineItemInput)

			err := workflow.ExecuteActivity(ctx, AddLineItemActivity, billID, lineItemInput).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to add line item", "Error", err)
				return
			}
			workflow.GetLogger(ctx).Info("Added line item", "BillID", billID, "Description", lineItemInput.Description)
		})

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			workflow.GetLogger(ctx).Info("Billing period ended, closing the bill", "BillID", billID)

			err := workflow.ExecuteActivity(ctx, FinalizeBillActivity, billID).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to finalize the bill", "Error", err)
				return
			}

			workflow.GetLogger(ctx).Info("Successfully finalized the bill", "BillID", billID)
			return
		})

		selector.Select(ctx)
	}
}

func CreateBillActivity(ctx context.Context, accountId string, currency string, periodStart, periodEnd time.Time) (int64, error) {
	return InsertBill(ctx, "Open", accountId, currency, periodStart, periodEnd)
}

func AddLineItemActivity(ctx context.Context, billID int64, lineItemInput AddLineItemSignalInput) error {
	_, err := InsertBillItem(ctx, billID, lineItemInput.Reference, lineItemInput.Description, lineItemInput.Amount, lineItemInput.Currency, lineItemInput.ExchangeRate)
	if err != nil {
		return err
	}

	return UpdateBillTotal(ctx, billID, lineItemInput.Amount)
}

func FinalizeBillActivity(ctx context.Context, billID int64) error {
	err := UpdateBillStatus(ctx, billID, "closed")
	if err != nil {
		return err
	}

	return nil
}
