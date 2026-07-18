-- +goose Up
CREATE TABLE telegram_settings (
    user_id      INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    bot_token    TEXT NOT NULL DEFAULT '',
    bot_username TEXT NOT NULL DEFAULT '',
    chat_id      TEXT NOT NULL DEFAULT '',
    enabled      INTEGER NOT NULL DEFAULT 0,
    frequency    TEXT NOT NULL DEFAULT 'daily', -- daily | weekly | monthly
    last_sent_at TEXT NOT NULL DEFAULT '',
    updated_at   TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE telegram_settings;
