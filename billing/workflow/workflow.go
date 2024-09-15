package workflow

import (
	"time"

	activity "encore.app/billing/activity"
	"go.temporal.io/sdk/workflow"
)

type CreateBillWorkflowInput struct {
	AccountId   string
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type WorkflowResult struct {
	BillID int64
}

func CreateBillWorkflow(ctx workflow.Context, input CreateBillWorkflowInput) (*WorkflowResult, error) {

	durationUntilEnd := input.PeriodEnd.Sub(workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: durationUntilEnd,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var billID int64
	err := workflow.ExecuteActivity(ctx, activity.CreateBillActivity, input.AccountId, input.Currency, input.PeriodStart, input.PeriodEnd).Get(ctx, &billID)
	if err != nil {
		return nil, err
	}

	isDone := false
	lineItemSignalCh := workflow.GetSignalChannel(ctx, activity.AddLineItemSignal)
	finalizeBillSignalCh := workflow.GetSignalChannel(ctx, activity.FinalizeBillSignal)
	timerFuture := workflow.NewTimer(ctx, durationUntilEnd)
	selector := workflow.NewSelector(ctx)

	for {
		selector.AddReceive(lineItemSignalCh, func(c workflow.ReceiveChannel, more bool) {
			var lineItemInput activity.AddLineItemSignalInput
			c.Receive(ctx, &lineItemInput)

			err := workflow.ExecuteActivity(ctx, activity.AddLineItemActivity, billID, lineItemInput).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to add line item", "Error", err)
				return
			}
			workflow.GetLogger(ctx).Info("Added line item", "BillID", billID, "Description", lineItemInput.Description)
		})

		selector.AddReceive(finalizeBillSignalCh, func(c workflow.ReceiveChannel, more bool) {
			workflow.GetLogger(ctx).Info("Received signal, closing the bill", "BillID", billID)

			err := workflow.ExecuteActivity(ctx, activity.FinalizeBillActivity, billID).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to finalize the bill", "Error", err)
				return
			}

			workflow.GetLogger(ctx).Info("Successfully finalized the bill", "BillID", billID)
			isDone = true
		})

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			workflow.GetLogger(ctx).Info("Billing period ended, closing the bill", "BillID", billID)

			err := workflow.ExecuteActivity(ctx, activity.TimerFinalizeBillActivity, billID).Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("Failed to finalize the bill", "Error", err)
				return
			}

			workflow.GetLogger(ctx).Info("Successfully finalized the bill", "BillID", billID)
			isDone = true
		})

		selector.Select(ctx)
		workflow.Sleep(ctx, time.Second*1)
		if isDone {
			return nil, nil
		}
	}
}
