CREATE TABLE IF NOT EXISTS pending_logins (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    email TEXT NOT NULL,
    login_id TEXT NOT NULL,
    verify_token TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    completed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);