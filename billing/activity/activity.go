package activity

import (
	"context"
	"time"

	db "encore.app/billing/db"
	"github.com/shopspring/decimal"
)

type AddLineItemSignalInput struct {
	Reference    string
	Description  string
	Amount       decimal.Decimal
	Currency     string
	ExchangeRate float64
}

func CreateBillActivity(ctx context.Context, accountId string, currency string, periodStart, periodEnd time.Time) (int64, error) {
	return db.InsertBill(ctx, db.StatusOpen, accountId, currency, periodStart, periodEnd)
}

func AddLineItemActivity(
	ctx context.Context,
	billID int64,
	lineItemInput AddLineItemSignalInput,
) error {
	_, err := db.InsertBillItem(ctx, billID, lineItemInput.Reference, lineItemInput.Description, lineItemInput.Amount, lineItemInput.Currency, lineItemInput.ExchangeRate)
	if err != nil {
		return err
	}

	return db.UpdateBillTotal(ctx, billID, lineItemInput.Amount)
}

func FinalizeBillActivity(ctx context.Context, billID int64) error {
	err := db.UpdateBillStatus(ctx, billID, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}

func TimerFinalizeBillActivity(ctx context.Context, billID int64) error {
	err := db.UpdateBillStatus(ctx, billID, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}
