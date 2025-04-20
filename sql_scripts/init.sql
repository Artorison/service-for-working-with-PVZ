CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    token TEXT
);

CREATE TABLE IF NOT EXISTS pvz (
    id TEXT PRIMARY KEY,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL,
    city TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
    id TEXT PRIMARY KEY,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL,
    pvz_id TEXT NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (pvz_id) REFERENCES pvz(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL,
    type TEXT NOT NULL,
    reception_id TEXT NOT NULL,
    FOREIGN KEY (reception_id) REFERENCES receptions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_receptions_pvz_date
    ON receptions(pvz_id, create_date);