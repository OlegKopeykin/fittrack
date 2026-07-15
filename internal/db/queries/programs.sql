-- name: CreateProgram :one
INSERT INTO programs (user_id, name, description, created_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListProgramsForUser :many
SELECT * FROM programs
WHERE user_id = ? AND archived_at IS NULL
ORDER BY id;

-- name: GetProgram :one
SELECT * FROM programs WHERE id = ?;

-- name: DeleteProgram :execrows
DELETE FROM programs WHERE id = ? AND user_id = ?;

-- name: CreateProgramDay :one
INSERT INTO program_days (program_id, position, name, notes)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListProgramDays :many
SELECT * FROM program_days WHERE program_id = ? ORDER BY position;

-- name: CreatePrescription :one
INSERT INTO prescriptions (program_day_id, exercise_id, position, sets, rep_min, rep_max, weight_min_g, weight_max_g, rest_sec, tempo, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListPrescriptionsForDay :many
SELECT * FROM prescriptions WHERE program_day_id = ? ORDER BY position;
