-- +goose Up
CREATE TABLE user_exercise_notes (
    user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    note        TEXT NOT NULL DEFAULT '',
    updated_at  TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (user_id, exercise_id)
);

-- +goose Down
DROP TABLE user_exercise_notes;
