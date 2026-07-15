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

	user, err := q.GetUserByUsername(ctx, e2eUsername)
	if errors.Is(err, sql.ErrNoRows) {
		hash, hErr := auth.HashPassword(e2ePassword)
		if hErr != nil {
			return hErr
		}
		user, err = q.CreateUser(ctx, gen.CreateUserParams{
			Username: e2eUsername, PasswordHash: hash, Role: "owner", CreatedAt: now,
		})
		if err != nil {
			return err
		}
		slog.Info("e2e: создан пользователь", "username", e2eUsername)
		if err := seedE2ETraining(ctx, q, user.ID, now); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// seedE2ETraining — демо-программа и тренировка для показа экранов.
func seedE2ETraining(ctx context.Context, q *gen.Queries, uid int64, now string) error {
	ex, err := q.GetGlobalExerciseByName(ctx, "Присед в Смите")
	if err != nil {
		return err
	}
	prog, err := q.CreateProgram(ctx, gen.CreateProgramParams{
		UserID: uid, Name: "Демо-программа", Description: "пример", CreatedAt: now,
	})
	if err != nil {
		return err
	}
	day, err := q.CreateProgramDay(ctx, gen.CreateProgramDayParams{
		ProgramID: prog.ID, Position: 0, Name: "День 1",
	})
	if err != nil {
		return err
	}
	if _, err := q.CreatePrescription(ctx, gen.CreatePrescriptionParams{
		ProgramDayID: day.ID, ExerciseID: ex.ID, Position: 0, Sets: 3,
		RepMin: sql.NullInt64{Int64: 6, Valid: true}, RepMax: sql.NullInt64{Int64: 10, Valid: true},
		WeightMinG: sql.NullInt64{Int64: 70000, Valid: true}, WeightMaxG: sql.NullInt64{Int64: 90000, Valid: true},
	}); err != nil {
		return err
	}

	wk, err := q.CreateWorkout(ctx, gen.CreateWorkoutParams{
		UserID: uid, Date: "2026-05-10", StartedAt: sql.NullString{String: now, Valid: true},
		BodyweightG: sql.NullInt64{Int64: 86500, Valid: true}, Feeling: "бодро",
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return err
	}
	sets := []struct {
		w, r int64
		role string
	}{{40000, 12, "warmup"}, {60000, 12, "working"}, {62000, 7, "working"}}
	for i, s := range sets {
		if _, err := q.CreateSet(ctx, gen.CreateSetParams{
			WorkoutID: wk.ID, ExerciseID: ex.ID, Position: int64(i), Role: s.role,
			WeightG: sql.NullInt64{Int64: s.w, Valid: true}, Reps: sql.NullInt64{Int64: s.r, Valid: true},
		}); err != nil {
			return err
		}
	}
	_, err = q.UpdateWorkout(ctx, gen.UpdateWorkoutParams{
		Date: "2026-05-10", FinishedAt: sql.NullString{String: now, Valid: true},
		BodyweightG: sql.NullInt64{Int64: 86500, Valid: true}, Feeling: "бодро",
		UpdatedAt: now, ID: wk.ID,
	})
	return err
}
