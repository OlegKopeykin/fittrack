// Package backup строит переносимый JSON-снимок лога тренировок пользователя
// (упражнения по имени, веса в кг) и восстанавливает его обратно.
package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

type ExportSet struct {
	Exercise    string   `json:"exercise"`
	Role        string   `json:"role"`
	WeightKg    *float64 `json:"weight_kg,omitempty"`
	Reps        *int     `json:"reps,omitempty"`
	DistanceKm  *float64 `json:"distance_km,omitempty"`
	DurationSec *int     `json:"duration_sec,omitempty"`
	Note        string   `json:"note,omitempty"`
}

type ExportWorkout struct {
	Date         string      `json:"date"`
	Title        string      `json:"title,omitempty"`
	StartedAt    string      `json:"started_at,omitempty"`
	FinishedAt   string      `json:"finished_at,omitempty"`
	BodyweightKg *float64    `json:"bodyweight_kg,omitempty"`
	Feeling      string      `json:"feeling,omitempty"`
	Notes        string      `json:"notes,omitempty"`
	Sets         []ExportSet `json:"sets"`
}

type Export struct {
	ExportedAt string          `json:"exported_at"`
	AppVersion string          `json:"app_version"`
	User       string          `json:"user"`
	Workouts   []ExportWorkout `json:"workouts"`
}

func gToKg(v sql.NullInt64) *float64 {
	if !v.Valid {
		return nil
	}
	kg := float64(v.Int64) / 1000
	return &kg
}

func mToKm(v sql.NullInt64) *float64 {
	if !v.Valid {
		return nil
	}
	km := float64(v.Int64) / 1000
	return &km
}

func nToInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

// Build собирает полный экспорт лога пользователя (в хронологическом порядке).
func Build(ctx context.Context, q *gen.Queries, userID int64, username, version, now string) (Export, error) {
	exp := Export{ExportedAt: now, AppVersion: version, User: username, Workouts: []ExportWorkout{}}

	exercises, err := q.ListExercisesForUser(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return exp, err
	}
	nameByID := make(map[int64]string, len(exercises))
	for _, e := range exercises {
		nameByID[e.ID] = e.Name
	}

	workouts, err := q.ListWorkoutsForUser(ctx, gen.ListWorkoutsForUserParams{
		UserID: userID, CursorDate: "", CursorID: 0, Lim: 1_000_000,
	})
	if err != nil {
		return exp, err
	}

	// Разворачиваем в хронологический порядок (запрос отдаёт свежие первыми).
	for i := len(workouts) - 1; i >= 0; i-- {
		wk := workouts[i]
		ew := ExportWorkout{
			Date: wk.Date, Title: wk.Title, StartedAt: wk.StartedAt.String, FinishedAt: wk.FinishedAt.String,
			BodyweightKg: gToKg(wk.BodyweightG), Feeling: wk.Feeling, Notes: wk.Notes, Sets: []ExportSet{},
		}
		sets, err := q.ListSetsForWorkout(ctx, wk.ID)
		if err != nil {
			return exp, err
		}
		for _, s := range sets {
			ew.Sets = append(ew.Sets, ExportSet{
				Exercise: nameByID[s.ExerciseID], Role: s.Role,
				WeightKg: gToKg(s.WeightG), Reps: nToInt(s.Reps),
				DistanceKm: mToKm(s.DistanceM), DurationSec: nToInt(s.DurationSec), Note: s.Note,
			})
		}
		exp.Workouts = append(exp.Workouts, ew)
	}
	return exp, nil
}

// Marshal возвращает форматированный JSON экспорта.
func Marshal(exp Export) ([]byte, error) {
	return json.MarshalIndent(exp, "", "  ")
}

// IsDue решает, пора ли слать бэкап при заданной частоте. Небольшой люфт (2ч)
// не даёт пропустить запуск таймера в 00:00 из-за секундных сдвигов.
func IsDue(frequency, lastSentAt string, now time.Time) bool {
	if lastSentAt == "" {
		return true
	}
	last, err := time.Parse(time.RFC3339, lastSentAt)
	if err != nil {
		return true
	}
	var interval time.Duration
	switch frequency {
	case "weekly":
		interval = 7 * 24 * time.Hour
	case "monthly":
		interval = 30 * 24 * time.Hour
	default: // daily
		interval = 24 * time.Hour
	}
	return now.Sub(last) >= interval-2*time.Hour
}
