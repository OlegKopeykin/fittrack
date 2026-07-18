package backup

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

func kgToG(kg *float64) sql.NullInt64 {
	if kg == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(math.Round(*kg * 1000)), Valid: true}
}

func kmToM(km *float64) sql.NullInt64 {
	if km == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(math.Round(*km * 1000)), Valid: true}
}

func intToN(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// Restore воссоздаёт лог из экспорта в одной транзакции. Идемпотентно по
// (дата, название, старт): уже существующие тренировки пропускаются, поэтому
// повторный импорт того же файла ничего не дублирует. Недостающие упражнения
// создаются как пользовательские с группой мышц по умолчанию. Возвращает
// число добавленных и пропущенных тренировок.
func Restore(ctx context.Context, database *sql.DB, q *gen.Queries, userID int64, now string, exp Export) (int, int, error) {
	existing, err := q.ListWorkoutsForUser(ctx, gen.ListWorkoutsForUserParams{
		UserID: userID, CursorDate: "", CursorID: 0, Lim: 1_000_000,
	})
	if err != nil {
		return 0, 0, err
	}
	seen := make(map[string]bool, len(existing))
	for _, w := range existing {
		seen[w.Date+"|"+w.Title+"|"+w.StartedAt.String] = true
	}

	exs, err := q.ListExercisesForUser(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return 0, 0, err
	}
	idByName := make(map[string]int64, len(exs))
	for _, e := range exs {
		idByName[e.Name] = e.ID
	}

	groups, err := q.ListMuscleGroups(ctx)
	if err != nil {
		return 0, 0, err
	}
	if len(groups) == 0 {
		return 0, 0, fmt.Errorf("нет групп мышц")
	}
	defaultGroup := groups[0].ID

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()
	qtx := q.WithTx(tx)

	imported, skipped := 0, 0
	for _, w := range exp.Workouts {
		key := w.Date + "|" + w.Title + "|" + w.StartedAt
		if w.Date == "" || seen[key] {
			skipped++
			continue
		}
		seen[key] = true

		wk, err := qtx.CreateWorkout(ctx, gen.CreateWorkoutParams{
			UserID: userID, Date: w.Date, Title: w.Title, StartedAt: nullStr(w.StartedAt),
			BodyweightG: kgToG(w.BodyweightKg), Feeling: w.Feeling, Notes: w.Notes,
			CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			return imported, skipped, err
		}
		for i, s := range w.Sets {
			name := strings.TrimSpace(s.Exercise)
			if name == "" {
				continue
			}
			exID, ok := idByName[name]
			if !ok {
				created, err := qtx.CreateExercise(ctx, gen.CreateExerciseParams{
					OwnerID: sql.NullInt64{Int64: userID, Valid: true}, Name: name,
					MuscleGroupID: defaultGroup, Kind: "compound", CreatedAt: now,
				})
				if err != nil {
					return imported, skipped, err
				}
				exID = created.ID
				idByName[name] = exID
			}
			role := s.Role
			if role == "" {
				role = "working"
			}
			if _, err := qtx.CreateSet(ctx, gen.CreateSetParams{
				WorkoutID: wk.ID, ExerciseID: exID, Position: int64(i), Role: role,
				WeightG: kgToG(s.WeightKg), Reps: intToN(s.Reps),
				DistanceM: kmToM(s.DistanceKm), DurationSec: intToN(s.DurationSec),
				Note: s.Note, CreatedAt: now,
			}); err != nil {
				return imported, skipped, err
			}
		}
		if w.FinishedAt != "" {
			if _, err := qtx.UpdateWorkout(ctx, gen.UpdateWorkoutParams{
				Date: w.Date, Title: w.Title, StartedAt: nullStr(w.StartedAt),
				FinishedAt: nullStr(w.FinishedAt), BodyweightG: kgToG(w.BodyweightKg),
				Feeling: w.Feeling, Notes: w.Notes, UpdatedAt: now, ID: wk.ID,
			}); err != nil {
				return imported, skipped, err
			}
		}
		imported++
	}
	if err := tx.Commit(); err != nil {
		return imported, skipped, err
	}
	return imported, skipped, nil
}
