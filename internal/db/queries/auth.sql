-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name, role, created_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UpdateUserPassword :execrows
UPDATE users SET password_hash = ? WHERE username = ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CreateInvite :one
INSERT INTO invites (code, role, created_by, created_at, expires_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetInviteByCode :one
SELECT * FROM invites WHERE code = ?;

-- name: ConsumeInvite :execrows
UPDATE invites
SET used_by = ?, used_at = ?
WHERE code = ?
  AND used_by IS NULL
  AND (expires_at IS NULL OR expires_at > ?);

-- name: ListInvites :many
SELECT * FROM invites ORDER BY id DESC;

-- name: DeleteUnusedInvite :execrows
DELETE FROM invites WHERE id = ? AND used_by IS NULL;
