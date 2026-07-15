package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

// Детерминированные фикстуры для Playwright (FITTRACK_E2E_SEED=1):
// owner-инвайт с фиксированным кодом + готовый пользователь для входа.
// Идемпотентно — создаёт лишь недостающее, чтобы параллельные проекты
// (iPhone + Chrome) не гонялись за единственным ресурсом.
const (
	e2eInviteCode = "E2ESEEDINVITE"
	e2eUsername   = "e2euser"
	e2ePassword   = "e2e-password-123"
)

func seedE2E(conn *sql.DB) error {
	q := gen.New(conn)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := q.GetInviteByCode(ctx, e2eInviteCode); errors.Is(err, sql.ErrNoRows) {
		if _, err := q.CreateInvite(ctx, gen.CreateInviteParams{
			Code: e2eInviteCode, Role: "owner", CreatedAt: now,
		}); err != nil {
			return err
		}
		slog.Info("e2e: создан owner-инвайт", "code", e2eInviteCode)
	} else if err != nil {
		return err
	}

	if _, err := q.GetUserByUsername(ctx, e2eUsername); errors.Is(err, sql.ErrNoRows) {
		hash, err := auth.HashPassword(e2ePassword)
		if err != nil {
			return err
		}
		if _, err := q.CreateUser(ctx, gen.CreateUserParams{
			Username: e2eUsername, PasswordHash: hash, Role: "owner", CreatedAt: now,
		}); err != nil {
			return err
		}
		slog.Info("e2e: создан пользователь", "username", e2eUsername)
	} else if err != nil {
		return err
	}
	return nil
}
