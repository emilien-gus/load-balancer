CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    capacity INT NOT NULL CHECK (capacity > 0),
    rate_per_sec INT NOT NULL CHECK (rate_per_sec > 0)
);