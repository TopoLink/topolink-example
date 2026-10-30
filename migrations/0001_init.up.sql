-- Orders, payments, accounts, ledger, settlements.
-- Applied by docker-compose via /docker-entrypoint-initdb.d.

CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    status      VARCHAR(32) NOT NULL,
    currency    VARCHAR(8) NOT NULL,
    total       BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

CREATE TABLE IF NOT EXISTS order_items (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sku         VARCHAR(64) NOT NULL,
    quantity    INTEGER NOT NULL,
    unit_price  BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);

CREATE TABLE IF NOT EXISTS payments (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id),
    provider    VARCHAR(32) NOT NULL,
    amount      BIGINT NOT NULL,
    status      VARCHAR(32) NOT NULL,
    external_id VARCHAR(128),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments (order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status);

CREATE TABLE IF NOT EXISTS accounts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    balance     BIGINT NOT NULL DEFAULT 0,
    currency    VARCHAR(8) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_accounts_user ON accounts (user_id);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id          BIGSERIAL PRIMARY KEY,
    account_id  BIGINT NOT NULL,
    order_id    BIGINT NOT NULL,
    amount      BIGINT NOT NULL,
    direction   VARCHAR(8) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger_entries (account_id, created_at DESC);

CREATE TABLE IF NOT EXISTS settlements (
    id          BIGSERIAL PRIMARY KEY,
    batch_id    VARCHAR(64) NOT NULL UNIQUE,
    amount      BIGINT NOT NULL DEFAULT 0,
    status      VARCHAR(32) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
