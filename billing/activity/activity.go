package activity

import (
	"context"
	"time"

	db "encore.app/billing/db"
	"github.com/shopspring/decimal"
)

type AddLineItemSignalInput struct {
	BillId       string
	Reference    string
	Description  string
	Amount       decimal.Decimal
	Currency     string
	ExchangeRate float64
}

type CreateBillInput struct {
	BillId      string
	AccountId   string
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type CloseBillInput struct {
	BillId string
}

func CreateBillActivity(ctx context.Context, input CreateBillInput) (string, error) {
	return db.InsertBill(ctx, input.BillId, db.StatusOpen, input.AccountId, input.Currency, input.PeriodStart, input.PeriodEnd)
}

func AddLineItemActivity(
	ctx context.Context,
	input AddLineItemSignalInput,
) error {

	_, err := db.InsertBillItem(ctx, input.BillId, input.Reference, input.Description, input.Amount, input.Currency, input.ExchangeRate)
	if err != nil {
		return err
	}
	return nil
}

func UpdateBillTotalActivity(
	ctx context.Context,
	input AddLineItemSignalInput,
) error {

	err := db.UpdateBillTotal(ctx, input.BillId, input.Amount)
	if err != nil {
		return err
	}
	return nil
}

func CloseBillActivity(ctx context.Context, input CloseBillInput) error {
	err := db.UpdateBillStatus(ctx, input.BillId, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}

func TimerCloseBillActivity(ctx context.Context, input CloseBillInput) error {
	err := db.UpdateBillStatus(ctx, input.BillId, db.StatusClosed)
	if err != nil {
		return err
	}

	return nil
}
