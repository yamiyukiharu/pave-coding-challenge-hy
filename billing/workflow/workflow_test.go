package workflow

import (
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
		BillId:      "1234",
		AccountId:   "account123",
		Currency:    "USD",
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
}

// Test to ensure bill creation
func (s *UnitTestSuite) TestCreateBill() {
	// Prepare
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything).Return(s.workflowInput.BillId, nil)
	s.env.OnActivity(activity.CloseBillActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityCalled(s.T(), "CreateBillActivity", mock.Anything, mock.Anything)
}

// Test to verify workflow execution when finalizing a bill via signal
func (s *UnitTestSuite) TestSignalFinalizeBill() {
	// Prepare
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything).Return(s.workflowInput.BillId, nil)
	s.env.OnActivity(activity.CloseBillActivity, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(activity.CloseBillSignal, activity.CloseBillInput{BillId: s.workflowInput.BillId})
	}, time.Hour)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityCalled(s.T(), "CloseBillActivity", mock.Anything, mock.MatchedBy(func(input activity.CloseBillInput) bool {
		return input.BillId == s.workflowInput.BillId
	}))
}

// Test to verify workflow execution with a timer to finalize a bill
func (s *UnitTestSuite) TestTimerFinalizeBill() {
	// Prepare
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything).Return(s.workflowInput.BillId, nil)
	s.env.OnActivity(activity.CloseBillActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityCalled(s.T(), "CloseBillActivity", mock.Anything, mock.Anything)
}

// Test to verify adding a line item via signal
func (s *UnitTestSuite) TestAddLineItem() {
	// Prepare
	lineItem := activity.AddLineItemSignalInput{
		Reference:    "REF001",
		Description:  "Service Fee",
		Amount:       decimal.NewFromFloat(100.00),
		Currency:     "USD",
		ExchangeRate: 1.0,
	}
	s.env.OnActivity(activity.CreateBillActivity, mock.Anything, mock.Anything).Return(s.workflowInput.BillId, nil)
	s.env.OnActivity(activity.AddLineItemActivity, mock.Anything, mock.Anything).Return(nil)
	s.env.OnActivity(activity.CloseBillActivity, mock.Anything, mock.Anything).Return(nil)

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(activity.AddLineItemSignal, lineItem)
	}, time.Second)

	// Execute
	s.env.ExecuteWorkflow(CreateBillWorkflow, s.workflowInput)

	// Assert
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.env.AssertActivityNumberOfCalls(s.T(), "AddLineItemActivity", 1)
	s.env.AssertActivityCalled(s.T(), "AddLineItemActivity", mock.Anything, mock.MatchedBy(func(input activity.AddLineItemSignalInput) bool {
		return cmp.Equal(input, lineItem)
	}))
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
