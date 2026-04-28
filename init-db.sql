CREATE DATABASE order_db;
CREATE DATABASE payment_db;

\connect order_db;

CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    item_name TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
    );

\connect payment_db;

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    amount BIGINT NOT NULL,
    status TEXT NOT NULL
);