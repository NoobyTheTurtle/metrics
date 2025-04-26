CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    value_double DOUBLE PRECISION NULL,
    value_int BIGINT NULL,
    UNIQUE(name, type)
);