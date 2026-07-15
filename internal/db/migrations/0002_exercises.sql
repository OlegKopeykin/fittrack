-- +goose Up
CREATE TABLE muscle_groups (
    id         INTEGER PRIMARY KEY,
    slug       TEXT NOT NULL UNIQUE,
    name_ru    TEXT NOT NULL,
    weekly_mev INTEGER,
    weekly_mav INTEGER,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE exercises (
    id              INTEGER PRIMARY KEY,
    owner_id        INTEGER REFERENCES users (id) ON DELETE CASCADE,  -- NULL = глобальный seed
    name            TEXT NOT NULL,
    muscle_group_id INTEGER NOT NULL REFERENCES muscle_groups (id),
    kind            TEXT NOT NULL CHECK (kind IN ('compound','isolation','isometric','bodyweight','cardio')),
    per_arm         INTEGER NOT NULL DEFAULT 0,
    technique_notes TEXT NOT NULL DEFAULT '',
    archived_at     TEXT,
    created_at      TEXT NOT NULL
);
-- Имя уникально в пределах владельца (глобальные — owner_id IS NULL → 0).
CREATE UNIQUE INDEX exercises_owner_name_idx ON exercises (COALESCE(owner_id, 0), name);
CREATE INDEX exercises_group_idx ON exercises (muscle_group_id);

CREATE TABLE exercise_aliases (
    id          INTEGER PRIMARY KEY,
    exercise_id INTEGER NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    alias       TEXT NOT NULL,
    alias_norm  TEXT NOT NULL  -- нормализовано: lowercase, ё→е, схлопнуты пробелы
);
CREATE UNIQUE INDEX exercise_aliases_uniq ON exercise_aliases (exercise_id, alias_norm);
CREATE INDEX exercise_aliases_norm_idx ON exercise_aliases (alias_norm);

CREATE TABLE api_tokens (
    id           INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,  -- sha256 hex от "fit_"+random
    prefix       TEXT NOT NULL,          -- первые символы для отображения в списке
    created_at   TEXT NOT NULL,
    last_used_at TEXT,
    expires_at   TEXT,
    revoked_at   TEXT
);
CREATE INDEX api_tokens_user_idx ON api_tokens (user_id);

-- +goose Down
DROP TABLE api_tokens;
DROP TABLE exercise_aliases;
DROP TABLE exercises;
DROP TABLE muscle_groups;
