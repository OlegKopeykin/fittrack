-- name: InsertMuscleGroup :one
INSERT INTO muscle_groups (slug, name_ru, weekly_mev, weekly_mav, sort_order)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListMuscleGroups :many
SELECT * FROM muscle_groups ORDER BY sort_order, name_ru;

-- name: GetMuscleGroupBySlug :one
SELECT * FROM muscle_groups WHERE slug = ?;

-- name: CreateExercise :one
INSERT INTO exercises (owner_id, name, muscle_group_id, kind, per_arm, technique_notes, equipment, instructions, video_url, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetExercise :one
SELECT * FROM exercises WHERE id = ?;

-- name: GetGlobalExerciseByName :one
SELECT * FROM exercises WHERE owner_id IS NULL AND name = ?;

-- name: UpdateExercise :one
UPDATE exercises
SET name = ?, muscle_group_id = ?, kind = ?, per_arm = ?, technique_notes = ?, equipment = ?, instructions = ?, video_url = ?
WHERE id = ?
RETURNING *;

-- name: SetSecondaryMuscle :exec
INSERT INTO exercise_secondary_muscles (exercise_id, muscle_group_id)
VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: ClearSecondaryMuscles :exec
DELETE FROM exercise_secondary_muscles WHERE exercise_id = ?;

-- name: ListSecondaryMuscleIDs :many
SELECT muscle_group_id FROM exercise_secondary_muscles WHERE exercise_id = ?;

-- name: SetExerciseImage :exec
INSERT INTO exercise_images (exercise_id, content_type, bytes, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT (exercise_id) DO UPDATE SET content_type = excluded.content_type, bytes = excluded.bytes, updated_at = excluded.updated_at;

-- name: GetExerciseImage :one
SELECT content_type, bytes FROM exercise_images WHERE exercise_id = ?;

-- name: DeleteExerciseImage :execrows
DELETE FROM exercise_images WHERE exercise_id = ?;

-- name: HasExerciseImage :one
SELECT EXISTS(SELECT 1 FROM exercise_images WHERE exercise_id = ?);

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
