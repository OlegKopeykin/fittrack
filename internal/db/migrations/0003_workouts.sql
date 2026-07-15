-- +goose Up
CREATE TABLE workouts (
    id           INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    date         TEXT NOT NULL,          -- YYYY-MM-DD, wall-clock
    started_at   TEXT,
    finished_at  TEXT,
    bodyweight_g INTEGER,
    feeling      TEXT NOT NULL DEFAULT '',
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);
CREATE INDEX workouts_user_date_idx ON workouts (user_id, date DESC, id DESC);

CREATE TABLE sets (
    id           INTEGER PRIMARY KEY,
    workout_id   INTEGER NOT NULL REFERENCES workouts (id) ON DELETE CASCADE,
    exercise_id  INTEGER NOT NULL REFERENCES exercises (id),
    position     INTEGER NOT NULL,
    role         TEXT NOT NULL DEFAULT 'working' CHECK (role IN ('warmup','ramp','working')),
    weight_g     INTEGER,   -- strength; NULL for bodyweight
    reps         INTEGER,
    distance_m   INTEGER,   -- cardio
    duration_sec INTEGER,   -- cardio / isometric holds
    note         TEXT NOT NULL DEFAULT '',
    client_id    TEXT UNIQUE,  -- client-generated UUID for idempotent replay
    created_at   TEXT NOT NULL
);
CREATE INDEX sets_workout_idx ON sets (workout_id, position);
CREATE INDEX sets_exercise_idx ON sets (exercise_id, workout_id);

-- +goose Down
DROP TABLE sets;
DROP TABLE workouts;
