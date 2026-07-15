-- +goose Up
ALTER TABLE workouts ADD COLUMN title TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE workouts DROP COLUMN title;
