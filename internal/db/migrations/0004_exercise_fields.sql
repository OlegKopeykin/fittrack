-- +goose Up
ALTER TABLE exercises ADD COLUMN equipment TEXT NOT NULL DEFAULT '';
ALTER TABLE exercises ADD COLUMN instructions TEXT NOT NULL DEFAULT '';
ALTER TABLE exercises ADD COLUMN video_url TEXT NOT NULL DEFAULT '';

CREATE TABLE exercise_secondary_muscles (
    exercise_id     INTEGER NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    muscle_group_id INTEGER NOT NULL REFERENCES muscle_groups (id),
    PRIMARY KEY (exercise_id, muscle_group_id)
);

CREATE TABLE exercise_images (
    exercise_id  INTEGER PRIMARY KEY REFERENCES exercises (id) ON DELETE CASCADE,
    content_type TEXT NOT NULL,
    bytes        BLOB NOT NULL,
    updated_at   TEXT NOT NULL
);

-- +goose Down
DROP TABLE exercise_images;
DROP TABLE exercise_secondary_muscles;
ALTER TABLE exercises DROP COLUMN video_url;
ALTER TABLE exercises DROP COLUMN instructions;
ALTER TABLE exercises DROP COLUMN equipment;
