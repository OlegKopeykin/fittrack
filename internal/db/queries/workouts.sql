-- name: CreateWorkout :one
INSERT INTO workouts (user_id, date, title, program_day_id, started_at, bodyweight_g, feeling, notes, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetWorkout :one
SELECT * FROM workouts WHERE id = ?;

-- name: ListWorkoutsForUser :many
SELECT * FROM workouts
WHERE user_id = @user_id
  AND (@cursor_date = '' OR date < @cursor_date OR (date = @cursor_date AND id < @cursor_id))
ORDER BY date DESC, id DESC
LIMIT @lim;

-- name: UpdateWorkout :one
UPDATE workouts
SET date = ?, title = ?, started_at = ?, finished_at = ?, bodyweight_g = ?, feeling = ?, notes = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteWorkout :execrows
DELETE FROM workouts WHERE id = ? AND user_id = ?;

-- name: CreateSet :one
INSERT INTO sets (workout_id, exercise_id, position, role, weight_g, reps, distance_m, duration_sec, note, client_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSet :one
SELECT * FROM sets WHERE id = ?;

-- name: GetSetByClientID :one
SELECT * FROM sets WHERE client_id = ?;

-- name: NextSetPosition :one
SELECT COALESCE(MAX(position), 0) + 1 FROM sets WHERE workout_id = ?;

-- name: ListSetsForWorkout :many
SELECT * FROM sets WHERE workout_id = ? ORDER BY position;

-- name: GetUnfinishedWorkoutForDay :one
SELECT * FROM workouts
WHERE user_id = ? AND program_day_id = ? AND date = ?
  AND (finished_at IS NULL OR finished_at = '')
ORDER BY id LIMIT 1;

-- name: UpdateSet :one
UPDATE sets
SET role = ?, weight_g = ?, reps = ?, distance_m = ?, duration_sec = ?, note = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSet :execrows
DELETE FROM sets WHERE id = ?;

-- name: ExerciseRecentSets :many
SELECT s.id, s.workout_id, s.position, s.role, s.weight_g, s.reps, s.distance_m, s.duration_sec, w.date AS workout_date
FROM sets s
JOIN workouts w ON w.id = s.workout_id
WHERE w.user_id = ? AND s.exercise_id = ? AND w.finished_at IS NOT NULL
ORDER BY w.date DESC, w.id DESC, s.position ASC
LIMIT ?;
