CREATE TABLE IF NOT EXISTS pending_logins (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    email TEXT NOT NULL,
    login_id TEXT NOT NULL,
    verify_token TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    completed_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);