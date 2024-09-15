package activity_test

import (
	"context"
	"testing"
	"time"

	"encore.app/billing/activity"
	db "encore.app/billing/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UnitTestSuite struct {
	suite.Suite
	ctx     context.Context
	mockDao *db.MockDao
}

func (s *UnitTestSuite) SetupTest() {
	// Create the mock DAO and inject it into the context
	s.mockDao = new(db.MockDao)
	s.ctx = db.SetDaoToContext(context.Background(), s.mockDao)
}

// TestCreateBillActivity tests the CreateBillActivity
func (s *UnitTestSuite) TestCreateBillActivity() {
	// Prepare
	accountId := "account123"
	currency := "USD"
	periodStart := time.Now()
	periodEnd := periodStart.Add(24 * time.Hour)

	s.mockDao.On("InsertBill", mock.Anything, db.StatusOpen, accountId, currency, periodStart, periodEnd).
		Return(int64(123), nil)

	// Execute
	billID, err := activity.CreateBillActivity(s.ctx, activity.CreateBillInput{
		AccountId:   accountId,
		Currency:    currency,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})

	// Assert
	require.NoError(s.T(), err, "activity should succeed")
	require.Equal(s.T(), int64(123), billID, "bill ID should be 123")
	s.mockDao.AssertCalled(s.T(), "InsertBill", mock.Anything, db.StatusOpen, accountId, currency, periodStart, periodEnd)
}

// TestAddLineItemActivity tests the AddLineItemActivity
func (s *UnitTestSuite) TestAddLineItemActivity() {
	// Prepare
	billID := int64(123)
	lineItemInput := activity.AddLineItemSignalInput{
		Reference:    "REF001",
		Description:  "Service Fee",
		Amount:       decimal.NewFromFloat(100.00),
		Currency:     "USD",
		ExchangeRate: 1.0,
	}

	s.mockDao.On("InsertBillItem", mock.Anything, billID, lineItemInput.Reference, lineItemInput.Description, lineItemInput.Amount, lineItemInput.Currency, lineItemInput.ExchangeRate).
		Return(int64(456), nil)

	s.mockDao.On("UpdateBillTotal", mock.Anything, billID, lineItemInput.Amount).
		Return(nil)

	// Execute
	err := activity.AddLineItemActivity(s.ctx, billID, lineItemInput)

	// Assert
	require.NoError(s.T(), err, "activity should succeed")
	s.mockDao.AssertCalled(s.T(), "InsertBillItem", mock.Anything, billID, lineItemInput.Reference, lineItemInput.Description, lineItemInput.Amount, lineItemInput.Currency, lineItemInput.ExchangeRate)
	s.mockDao.AssertCalled(s.T(), "UpdateBillTotal", mock.Anything, billID, lineItemInput.Amount)
}

// TestFinalizeBillActivity tests the FinalizeBillActivity
func (s *UnitTestSuite) TestFinalizeBillActivity() {
	// Prepare
	billID := int64(123)

	s.mockDao.On("UpdateBillStatus", mock.Anything, billID, db.StatusClosed).
		Return(nil)

	// Execute
	err := activity.FinalizeBillActivity(s.ctx, billID)

	// Assert
	require.NoError(s.T(), err, "activity should succeed")
	s.mockDao.AssertCalled(s.T(), "UpdateBillStatus", mock.Anything, billID, db.StatusClosed)
}

// TestTimerFinalizeBillActivity tests the TimerFinalizeBillActivity
func (s *UnitTestSuite) TestTimerFinalizeBillActivity() {
	// Prepare
	billID := int64(123)

	s.mockDao.On("UpdateBillStatus", mock.Anything, billID, db.StatusClosed).
		Return(nil)

	// Execute
	err := activity.TimerFinalizeBillActivity(s.ctx, billID)

	// Assert
	require.NoError(s.T(), err, "activity should succeed")
	s.mockDao.AssertCalled(s.T(), "UpdateBillStatus", mock.Anything, billID, db.StatusClosed)
}

// Run the test suite
func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
