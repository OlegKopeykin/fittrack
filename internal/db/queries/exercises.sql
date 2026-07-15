-- name: InsertMuscleGroup :one
INSERT INTO muscle_groups (slug, name_ru, weekly_mev, weekly_mav, sort_order)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListMuscleGroups :many
SELECT * FROM muscle_groups ORDER BY sort_order, name_ru;

-- name: GetMuscleGroupBySlug :one
SELECT * FROM muscle_groups WHERE slug = ?;

-- name: CreateExercise :one
INSERT INTO exercises (owner_id, name, muscle_group_id, kind, per_arm, technique_notes, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetExercise :one
SELECT * FROM exercises WHERE id = ?;

-- name: GetGlobalExerciseByName :one
SELECT * FROM exercises WHERE owner_id IS NULL AND name = ?;

-- name: UpdateExercise :one
UPDATE exercises
SET name = ?, muscle_group_id = ?, kind = ?, per_arm = ?, technique_notes = ?
WHERE id = ?
RETURNING *;

-- name: ArchiveExercise :execrows
UPDATE exercises SET archived_at = ? WHERE id = ?;

-- name: ListExercisesForUser :many
SELECT * FROM exercises
WHERE (owner_id IS NULL OR owner_id = ?)
ORDER BY name;

-- name: SearchAliasExerciseIDs :many
SELECT DISTINCT exercise_id FROM exercise_aliases WHERE alias_norm LIKE ?;

-- name: AddAlias :one
INSERT INTO exercise_aliases (exercise_id, alias, alias_norm)
VALUES (?, ?, ?)
RETURNING *;

-- name: CountAlias :one
SELECT COUNT(*) FROM exercise_aliases WHERE exercise_id = ? AND alias_norm = ?;

-- name: ListAliasesForExercise :many
SELECT * FROM exercise_aliases WHERE exercise_id = ? ORDER BY alias;

-- name: DeleteAlias :execrows
DELETE FROM exercise_aliases WHERE id = ? AND exercise_id = ?;

-- name: CreateApiToken :one
INSERT INTO api_tokens (user_id, name, token_hash, prefix, created_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetApiTokenByHash :one
SELECT * FROM api_tokens WHERE token_hash = ?;

-- name: TouchApiToken :exec
UPDATE api_tokens SET last_used_at = ? WHERE id = ?;

-- name: ListApiTokens :many
SELECT * FROM api_tokens WHERE user_id = ? ORDER BY id DESC;

-- name: RevokeApiToken :execrows
UPDATE api_tokens SET revoked_at = ? WHERE id = ? AND user_id = ? AND revoked_at IS NULL;
