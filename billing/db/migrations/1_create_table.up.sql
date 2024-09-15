CREATE TABLE bill (
    id BIGSERIAL PRIMARY KEY,
    status VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    total_amount DECIMAL(20, 8) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT status_idx UNIQUE (status),
    CONSTRAINT account_id_idx UNIQUE (account_id)
);

CREATE INDEX idx_bills_status ON bill(status);
CREATE INDEX idx_bills_account_id ON bill(account_id);

CREATE TABLE bill_item (
    id BIGSERIAL PRIMARY KEY,
    bill_id BIGINT REFERENCES bill(id) ON DELETE CASCADE,
    reference VARCHAR(255) NOT NULL,
    description TEXT,
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    exchange_rate FLOAT8 NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_db_bill_items_bill_id ON bill_item(bill_id);
