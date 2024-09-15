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

type CreateBillInput struct {
	AccountId   string
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func CreateBillActivity(ctx context.Context, input CreateBillInput) (int64, error) {
	dao := db.GetDaoFromContext(ctx)
	return dao.InsertBill(ctx, db.StatusOpen, input.AccountId, input.Currency, input.PeriodStart, input.PeriodEnd)
}

func AddLineItemActivity(
	ctx context.Context,
	billID int64,
	lineItemInput AddLineItemSignalInput,
) error {
	dao := db.GetDaoFromContext(ctx)

	_, err := dao.InsertBillItem(ctx, billID, lineItemInput.Reference, lineItemInput.Description, lineItemInput.Amount, lineItemInput.Currency, lineItemInput.ExchangeRate)
	if err != nil {
		return err
	}

	return dao.UpdateBillTotal(ctx, billID, lineItemInput.Amount)
}

func FinalizeBillActivity(ctx context.Context, billID int64) error {
	dao := db.GetDaoFromContext(ctx)
	err := dao.UpdateBillStatus(ctx, billID, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}

func TimerFinalizeBillActivity(ctx context.Context, billID int64) error {
	dao := db.GetDaoFromContext(ctx)
	err := dao.UpdateBillStatus(ctx, billID, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}
