package db

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	"github.com/shopspring/decimal"
)

type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

type DbBill struct {
	Id          string    `db:"id,pk,auto"`
	Status      Status    `db:"status"` // index
	Currency    string    `db:"currency"`
	AccountId   string    `db:"account_id"` // index
	PeriodStart time.Time `db:"end_at"`
	PeriodEnd   time.Time `db:"end_at"`
	CreatedAt   time.Time `db:"created_at"`
}

type DbBillItem struct {
	Id           int64           `db:"id,pk,auto"`
	BillId       string          `db:"bill_id"` // index
	Reference    string          `db:"reference"`
	Description  string          `db:"description"`
	Amount       decimal.Decimal `db:"amount"`
	Currency     string          `db:"currency"`
	ExchangeRate float64         `db:"exchange_rate"`
	CreatedAt    time.Time       `db:"created_at"`
}

type BillingDaoInterface interface {
	InsertBill(ctx context.Context, status Status, accountId, currency string, periodStart, periodEnd time.Time) (int64, error)
	InsertBillItem(ctx context.Context, billId int64, reference, description string, amount decimal.Decimal, currency string, exchangeRate float64) (int64, error)
	GetBillByID(ctx context.Context, billId int64) (*DbBill, error)
	GetBillItems(ctx context.Context, billId int64) ([]DbBillItem, error)
	UpdateBillTotal(ctx context.Context, billId int64, amount decimal.Decimal) error
	UpdateBillStatus(ctx context.Context, billId int64, status Status) error
}

type BillingDao struct {
	Db *sqldb.Database
}

var db = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func InsertBill(ctx context.Context, id string, status Status, accountId string, currency string, periodStart, periodEnd time.Time) (string, error) {
	const query = `
		INSERT INTO bill (id, status, total_amount, account_id, currency, period_start, period_end, created_at)
		VALUES ($1, $2, 0, $3, $4, $5, $6, now())
		RETURNING id
	`
	err := db.QueryRow(ctx, query, id, status, accountId, currency, periodStart, periodEnd).Scan(&id)
	return id, err
}

func InsertBillItem(ctx context.Context, billId string, reference, description string, amount decimal.Decimal, currency string, exchangeRate float64) (int64, error) {
	const query = `
		INSERT INTO bill_item (bill_id, reference, description, amount, currency, exchange_rate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING id
	`
	var id int64
	err := db.QueryRow(ctx, query, billId, reference, description, amount, currency, exchangeRate).Scan(&id)
	return id, err
}

func GetBillByID(ctx context.Context, billId string) (*DbBill, error) {
	const query = `
		SELECT id, status, currency, account_id, created_at
		FROM bill
		WHERE id = $1
	`
	var bill DbBill
	err := db.QueryRow(ctx, query, billId).Scan(
		&bill.Id,
		&bill.Status,
		&bill.Currency,
		&bill.AccountId,
		&bill.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func GetBillItems(ctx context.Context, billId string) ([]DbBillItem, error) {
	const query = `
		SELECT id, bill_id, reference, description, amount, currency, exchange_rate, created_at
		FROM bill_item
		WHERE bill_id = $1
	`
	rows, err := db.Query(ctx, query, billId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DbBillItem
	for rows.Next() {
		var item DbBillItem
		err := rows.Scan(&item.Id, &item.BillId, &item.Reference, &item.Description, &item.Amount, &item.Currency, &item.ExchangeRate, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func UpdateBillTotal(ctx context.Context, billId string, amount decimal.Decimal) error {
	const query = `
		UPDATE bill
		SET total_amount = total_amount + $1
		WHERE id = $2
	`
	_, err := db.Exec(ctx, query, amount, billId)
	return err
}

func UpdateBillStatus(ctx context.Context, billId string, status Status) error {
	const query = `
		UPDATE bill
		SET status = $1
		WHERE id = $2
	`
	_, err := db.Exec(ctx, query, status, billId)
	return err
}

func GetBillDetailsWithTotal(ctx context.Context, billId string) (*DbBill, []DbBillItem, decimal.Decimal, error) {
	bill, err := GetBillByID(ctx, billId)
	if err != nil {
		return nil, nil, decimal.Zero, err
	}

	lineItems, err := GetBillItems(ctx, billId)
	if err != nil {
		return nil, nil, decimal.Zero, err
	}

	totalAmount := decimal.Zero
	for _, item := range lineItems {
		totalAmount = totalAmount.Add(item.Amount)
	}

	return bill, lineItems, totalAmount, nil
}

func GetBillsByAccountAndStatus(ctx context.Context, accountId string, status Status) ([]DbBill, error) {
	const query = `
		SELECT id, status, currency, account_id, period_start, period_end, created_at
		FROM bill
		WHERE status = $1 AND account_id= $2
	`
	rows, err := db.Query(ctx, query, status, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []DbBill
	for rows.Next() {
		var bill DbBill
		err := rows.Scan(
			&bill.Id,
			&bill.Status,
			&bill.Currency,
			&bill.AccountId,
			&bill.PeriodStart,
			&bill.PeriodEnd,
			&bill.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}

	return bills, nil
}
