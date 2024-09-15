package db_test

import (
	"context"
	"testing"
	"time"

	"encore.app/billing/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestInsertBillAndRetrieve(t *testing.T) {
	ctx := context.Background()

	// Insert a new bill with ETH as the currency
	periodStart := time.Now()
	periodEnd := periodStart.Add(24 * time.Hour)
	billID, err := db.InsertBill(ctx, "bill1", db.StatusOpen, "accountETH", "ETH", periodStart, periodEnd)
	require.NoError(t, err, "failed to insert bill")
	require.NotZero(t, billID, "bill ID should not be zero")

	// Retrieve the bill by ID and verify its fields
	bill, err := db.GetBillByID(ctx, billID)
	require.NoError(t, err, "failed to get bill by ID")
	require.Equal(t, db.StatusOpen, bill.Status, "bill status should be 'open'")
	require.Equal(t, "ETH", bill.Currency, "bill currency should be 'ETH'")
	require.Equal(t, "accountETH", bill.AccountId, "account ID should be 'accountETH'")
}

func TestInsertBillItemAndRetrieve(t *testing.T) {
	ctx := context.Background()

	// Insert a new bill with USD as the currency
	periodStart := time.Now()
	periodEnd := periodStart.Add(24 * time.Hour)
	billID, err := db.InsertBill(ctx, "bill2", db.StatusOpen, "account456", "USD", periodStart, periodEnd)
	require.NoError(t, err, "failed to insert bill")
	require.NotZero(t, billID, "bill ID should not be zero")

	// Insert a bill item with high precision (e.g., 18 decimal places)
	amount := decimal.NewFromFloat(0.123456789012345678)
	itemID, err := db.InsertBillItem(ctx, billID, "REF001", "Crypto Payment", amount, "ETH", 3000.123456789012345678)
	require.NoError(t, err, "failed to insert bill item")
	require.NotZero(t, itemID, "bill item ID should not be zero")

	// Retrieve the bill items and verify
	items, err := db.GetBillItems(ctx, billID)
	require.NoError(t, err, "failed to get bill items")
	require.Len(t, items, 1, "there should be one bill item")

	item := items[0]
	require.Equal(t, "REF001", item.Reference, "reference should match")
	require.Equal(t, "Crypto Payment", item.Description, "description should match")
	require.Equal(t, amount.Round(18), item.Amount.Round(18), "amount should match with high precision")
	require.Equal(t, "ETH", item.Currency, "currency should match")
	require.Equal(t, decimal.NewFromFloat(3000.123456789012345678).Round(18), decimal.NewFromFloat(item.ExchangeRate).Round(18), "exchange rate should match")
}

func TestUpdateBillStatus(t *testing.T) {
	ctx := context.Background()

	// Insert a new bill
	periodStart := time.Now()
	periodEnd := periodStart.Add(24 * time.Hour)
	billID, err := db.InsertBill(ctx, "bill4", db.StatusOpen, "account999", "USD", periodStart, periodEnd)
	require.NoError(t, err, "failed to insert bill")

	// Update the bill status to 'closed'
	err = db.UpdateBillStatus(ctx, billID, db.StatusClosed)
	require.NoError(t, err, "failed to update bill status")

	// Retrieve the bill and verify the status is now 'closed'
	bill, err := db.GetBillByID(ctx, billID)
	require.NoError(t, err, "failed to get updated bill")
	require.Equal(t, db.StatusClosed, bill.Status, "bill status should be 'closed'")
}
