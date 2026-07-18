-- name: GetUserExerciseNote :one
SELECT note FROM user_exercise_notes WHERE user_id = ? AND exercise_id = ?;

-- name: UpsertUserExerciseNote :exec
INSERT INTO user_exercise_notes (user_id, exercise_id, note, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, exercise_id) DO UPDATE SET
    note = excluded.note,
    updated_at = excluded.updated_at;
