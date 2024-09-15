package workflow

import (
	// Use standard context package

	"testing"
	"time"

	"encore.app/billing/activity"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env           *testsuite.TestWorkflowEnvironment
	workflowInput CreateBillWorkflowInput
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	periodStart := time.Now()
	periodEnd := periodStart.Add(24 * time.Hour)
	s.workflowInput = CreateBillWorkflowInput{
		AccountId:   "account123",
		Currency:    "USD",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
}

func (s *UnitTestSuite) TestSignalFinalizeBill() {
	// Prepare
	billId := int64(123)
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(billId, nil)
	s.env.OnActivity(activity.FinalizeBillActivity, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(activity.FinalizeBillSignal, nil)
	}, time.Hour)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityCalled(s.T(), "FinalizeBillActivity", mock.Anything, billId)
	s.env.AssertActivityNotCalled(s.T(), "TimerFinalizeBillActivity", mock.Anything, billId)
}

func (s *UnitTestSuite) TestTimerFinalizeBill() {
	// Prepare
	billId := int64(123)
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(billId, nil)
	s.env.OnActivity(activity.TimerFinalizeBillActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityCalled(s.T(), "TimerFinalizeBillActivity", mock.Anything, billId)
	s.env.AssertActivityNotCalled(s.T(), "FinalizeBillActivity", mock.Anything, billId)
}

func (s *UnitTestSuite) TestAddLineItem() {
	// Prepare
	billId := int64(123)
	lineItem := activity.AddLineItemSignalInput{
		Reference:    "REF001",
		Description:  "Service Fee",
		Amount:       decimal.NewFromFloat(100.00),
		Currency:     "USD",
		ExchangeRate: 1.0,
	}
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(billId, nil)
	s.env.OnActivity(activity.AddLineItemActivity, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activity.TimerFinalizeBillActivity, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(activity.AddLineItemSignal, lineItem)
	}, time.Hour)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityNumberOfCalls(s.T(), "AddLineItemActivity", 1)
	s.env.AssertActivityCalled(
		s.T(),
		"AddLineItemActivity",
		mock.Anything,
		billId,
		mock.MatchedBy(func(args activity.AddLineItemSignalInput) bool { return cmp.Equal(args, lineItem) }),
	)
	s.env.AssertActivityCalled(s.T(), "TimerFinalizeBillActivity", mock.Anything, billId)
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
