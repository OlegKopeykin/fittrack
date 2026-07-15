-- +goose Up
CREATE TABLE programs (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    archived_at TEXT,
    created_at  TEXT NOT NULL
);

CREATE TABLE program_days (
    id         INTEGER PRIMARY KEY,
    program_id INTEGER NOT NULL REFERENCES programs (id) ON DELETE CASCADE,
    position   INTEGER NOT NULL,
    name       TEXT NOT NULL,
    notes      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX program_days_program_idx ON program_days (program_id, position);

CREATE TABLE prescriptions (
    id             INTEGER PRIMARY KEY,
    program_day_id INTEGER NOT NULL REFERENCES program_days (id) ON DELETE CASCADE,
    exercise_id    INTEGER NOT NULL REFERENCES exercises (id),
    position       INTEGER NOT NULL,
    sets           INTEGER NOT NULL DEFAULT 0,
    rep_min        INTEGER,
    rep_max        INTEGER,
    weight_min_g   INTEGER,
    weight_max_g   INTEGER,
    rest_sec       INTEGER,
    tempo          TEXT NOT NULL DEFAULT '',
    notes          TEXT NOT NULL DEFAULT ''
);
CREATE INDEX prescriptions_day_idx ON prescriptions (program_day_id, position);

ALTER TABLE workouts ADD COLUMN program_day_id INTEGER REFERENCES program_days (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE workouts DROP COLUMN program_day_id;
DROP TABLE prescriptions;
DROP TABLE program_days;
DROP TABLE programs;
