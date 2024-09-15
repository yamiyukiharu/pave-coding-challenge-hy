package workflow

import (
	"time"

	activity "encore.app/billing/activity"
	"go.temporal.io/sdk/workflow"
)

type CreateBillWorkflowInput struct {
	BillId      string
	AccountId   string
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type WorkflowResult struct {
	BillId string
}

func CreateBillWorkflow(ctx workflow.Context, workflowInput CreateBillWorkflowInput) (*WorkflowResult, error) {

	durationUntilEnd := workflowInput.PeriodEnd.Sub(workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: durationUntilEnd,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, activity.CreateBillActivity, workflowInput).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	isDone := false
	isEnded := false
	createBillSignalCh := workflow.GetSignalChannel(ctx, activity.CreateBillSignal)
	lineItemSignalCh := workflow.GetSignalChannel(ctx, activity.AddLineItemSignal)
	finalizeBillSignalCh := workflow.GetSignalChannel(ctx, activity.CloseBillSignal)
	timerFuture := workflow.NewTimer(ctx, durationUntilEnd)
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(createBillSignalCh, func(c workflow.ReceiveChannel, more bool) {
		var input activity.CreateBillInput
		c.Receive(ctx, &input)

		err := workflow.ExecuteActivity(ctx, activity.CreateBillActivity, input).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to create Bull", "Error", err)
			return
		}
		workflow.GetLogger(ctx).Info("Created Bull", "BillId", input.BillId)
	})

	selector.AddReceive(lineItemSignalCh, func(c workflow.ReceiveChannel, more bool) {
		var input activity.AddLineItemSignalInput
		c.Receive(ctx, &input)
		workflow.GetLogger(ctx).Info("Received signal,adding item to bill", "BillId", input.BillId)

		err := workflow.ExecuteActivity(ctx, activity.AddLineItemActivity, input).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to add line item", "Error", err)
			return
		}
		workflow.GetLogger(ctx).Info("Added line item", "BillId", input.BillId, "Description", input.Description)
	})

	selector.AddReceive(finalizeBillSignalCh, func(c workflow.ReceiveChannel, more bool) {
		var input activity.CloseBillInput
		c.Receive(ctx, &input)
		workflow.GetLogger(ctx).Info("Received signal, closing the bill", "BillId", input.BillId)

		err := workflow.ExecuteActivity(ctx, activity.CloseBillActivity, input).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("Failed to finalize the bill", "Error", err)
			return
		}

		workflow.GetLogger(ctx).Info("Successfully finalized the bill", "BillId", input.BillId)
		isDone = true
	})

	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if isEnded {
			return
		}
		workflow.GetLogger(ctx).Info("Billing period ended, closing the bill", "BillId", workflowInput.BillId)

		workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
		runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

		workflow.SignalExternalWorkflow(ctx, workflowID, runID, activity.CloseBillSignal, activity.CloseBillInput{BillId: workflowInput.BillId})
		isEnded = true
	})

	for {
		selector.Select(ctx)
		if isDone {
			return nil, nil
		}
		workflow.Sleep(ctx, time.Second*1)
	}
}
