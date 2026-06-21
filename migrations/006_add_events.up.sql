CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    date TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    address TEXT,
    n_of_slots INTEGER NOT NULL,
    n_of_volunteers INTEGER NOT NULL,
    total_hours REAL,
    leaders TEXT,
    made_by TEXT,
    sign_up_url TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);