CREATE TABLE IF NOT EXISTS member_tokens (
    member_id TEXT NOT NULL REFERENCES members(id),
    token TEXT NOT NULL,
    PRIMARY KEY (member_id, token)
);

CREATE INDEX IF NOT EXISTS idx_member_tokens_token ON member_tokens(token);