CREATE TABLE bill (
    id VARCHAR(255) PRIMARY KEY, 
    status VARCHAR(255) NOT NULL,
    currency VARCHAR(255) NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_bills_status ON bill(status);
CREATE INDEX idx_bills_account_id ON bill(account_id);

CREATE TABLE bill_item (
    id BIGSERIAL PRIMARY KEY,
    bill_id VARCHAR(255) REFERENCES bill(id) ON DELETE CASCADE, 
    reference VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(38, 18) NOT NULL,
    currency VARCHAR(255) NOT NULL,
    exchange_rate DECIMAL(30, 10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_db_bill_items_bill_id ON bill_item(bill_id);
