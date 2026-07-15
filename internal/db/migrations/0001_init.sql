-- +goose Up
CREATE TABLE users (
    id            INTEGER PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('owner', 'user')),
    created_at    TEXT NOT NULL
);

CREATE TABLE invites (
    id         INTEGER PRIMARY KEY,
    code       TEXT NOT NULL UNIQUE,
    role       TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('owner', 'user')),
    created_by INTEGER REFERENCES users (id),
    created_at TEXT NOT NULL,
    expires_at TEXT,
    used_by    INTEGER REFERENCES users (id),
    used_at    TEXT
);

-- Схема хранилища сессий scs (alexedwards/scs/sqlite3store).
CREATE TABLE sessions (
    token  TEXT PRIMARY KEY,
    data   BLOB NOT NULL,
    expiry REAL NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- +goose Down
DROP TABLE sessions;
DROP TABLE invites;
DROP TABLE users;
