package billing

import (
	"context"
	"time"

	"encore.dev/storage/sqldb"
	"github.com/shopspring/decimal"
)

type DbBill struct {
	ID          int64           `db:"id,pk,auto"`
	Status      string          `db:"status"` // index
	Currency    string          `db:"currency"`
	AccountId   string          `db:"account_id"` // index
	TotalAmount decimal.Decimal `db:"total_amount"`
	periodStart time.Time       `db:"end_at"`
	periodEnd   time.Time       `db:"end_at"`
	CreatedAt   time.Time       `db:"created_at"`
}

type DbBillItem struct {
	ID           int64           `db:"id,pk,auto"`
	BillID       int64           `db:"bill_id"` // index
	Reference    string          `db:"reference"`
	Description  string          `db:"description"`
	Amount       decimal.Decimal `db:"amount"`
	Currency     string          `db:"currency"`
	ExchangeRate float64         `db:"exchange_rate"`
	CreatedAt    time.Time       `db:"created_at"`
}

var billsDb = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

func InsertBill(ctx context.Context, status string, accountId string, currency string, periodStart, periodEnd time.Time) (int64, error) {
	const query = `
		INSERT INTO bill (status, total_amount, account_id, currency, period_start, period_end, created_at)
		VALUES ($1, 0, $2, $3, $4, $5, now())
		RETURNING id
	`
	var id int64
	err := billsDb.QueryRow(ctx, query, status, currency).Scan(&id)
	return id, err
}

func InsertBillItem(ctx context.Context, billID int64, reference, description string, amount decimal.Decimal, currency string, exchangeRate float64) (int64, error) {
	const query = `
		INSERT INTO bill_item (bill_id, reference, description, amount, currency, exchange_rate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING id
	`
	var id int64
	err := billsDb.QueryRow(ctx, query, billID, reference, description, amount, currency, exchangeRate).Scan(&id)
	return id, err
}

func GetBillByID(ctx context.Context, billID int64) (*DbBill, error) {
	const query = `
		SELECT id, status, total_amount, currency, created_at
		FROM bill
		WHERE id = $1
	`
	var bill DbBill
	err := billsDb.QueryRow(ctx, query, billID).Scan(&bill.ID, &bill.Status, &bill.TotalAmount, &bill.Currency, &bill.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func GetBillItems(ctx context.Context, billID int64) ([]DbBillItem, error) {
	const query = `
		SELECT id, bill_id, reference, description, amount, currency, exchange_rate, created_at
		FROM bill_item
		WHERE bill_id = $1
	`
	rows, err := billsDb.Query(ctx, query, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DbBillItem
	for rows.Next() {
		var item DbBillItem
		err := rows.Scan(&item.ID, &item.BillID, &item.Reference, &item.Description, &item.Amount, &item.Currency, &item.ExchangeRate, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func UpdateBillTotal(ctx context.Context, billID int64, amount decimal.Decimal) error {
	const query = `
		UPDATE bill
		SET total_amount = total_amount + $1
		WHERE id = $2
	`
	_, err := billsDb.Exec(ctx, query, amount, billID)
	return err
}

func UpdateBillStatus(ctx context.Context, billID int64, status string) error {
	const query = `
		UPDATE bill
		SET status = $1
		WHERE id = $2
	`
	_, err := billsDb.Exec(ctx, query, status, billID)
	return err
}
