CREATE TABLE IF NOT EXISTS metrics (
    id          SERIAL PRIMARY KEY,
    key         VARCHAR(255) NOT NULL,
    value_float DOUBLE PRECISION NULL,
    value_int   BIGINT NULL,
    UNIQUE(key)
);