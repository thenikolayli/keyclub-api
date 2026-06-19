CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    nickname TEXT,
    middle_name TEXT,
    last_name TEXT NOT NULL,
    all_hours REAL NOT NULL,
    term_hours REAL NOT NULL,
    grad_year INTEGER NOT NULL,
    class TEXT NOT NULL,
    strikes INTEGER NOT NULL,
    personal_email TEXT NOT NULL,
    school_email TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    shirt_size TEXT NOT NULL,
    paid_dues BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);