-- name: GetTelegramSettings :one
SELECT * FROM telegram_settings WHERE user_id = ?;

-- name: UpsertTelegramSettings :exec
INSERT INTO telegram_settings (user_id, bot_token, bot_username, chat_id, enabled, frequency, last_sent_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id) DO UPDATE SET
    bot_token = excluded.bot_token,
    bot_username = excluded.bot_username,
    chat_id = excluded.chat_id,
    enabled = excluded.enabled,
    frequency = excluded.frequency,
    last_sent_at = excluded.last_sent_at,
    updated_at = excluded.updated_at;

-- name: DeleteTelegramSettings :exec
DELETE FROM telegram_settings WHERE user_id = ?;

-- name: ListEnabledTelegram :many
SELECT * FROM telegram_settings WHERE enabled = 1 AND bot_token <> '' AND chat_id <> '';

-- name: TouchTelegramSent :exec
UPDATE telegram_settings SET last_sent_at = ? WHERE user_id = ?;
