package db

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

// MockDao should implement BillingDaoInterface
type MockDao struct {
	mock.Mock
}

func (m *MockDao) InsertBill(ctx context.Context, status Status, accountId, currency string, periodStart, periodEnd time.Time) (int64, error) {
	args := m.Called(ctx, status, accountId, currency, periodStart, periodEnd)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDao) InsertBillItem(ctx context.Context, billID int64, reference, description string, amount decimal.Decimal, currency string, exchangeRate float64) (int64, error) {
	args := m.Called(ctx, billID, reference, description, amount, currency, exchangeRate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDao) GetBillByID(ctx context.Context, billID int64) (*DbBill, error) {
	args := m.Called(ctx, billID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DbBill), args.Error(1)
}

func (m *MockDao) GetBillItems(ctx context.Context, billID int64) ([]DbBillItem, error) {
	args := m.Called(ctx, billID)
	return args.Get(0).([]DbBillItem), args.Error(1)
}

func (m *MockDao) UpdateBillTotal(ctx context.Context, billID int64, amount decimal.Decimal) error {
	args := m.Called(ctx, billID, amount)
	return args.Error(0)
}

func (m *MockDao) UpdateBillStatus(ctx context.Context, billID int64, status Status) error {
	args := m.Called(ctx, billID, status)
	return args.Error(0)
}
