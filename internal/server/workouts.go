package server

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

// --- конвертация весов/дистанций между API (кг/км) и хранилищем (граммы/метры) ---

func gramsToKg(v sql.NullInt64) *float64 {
	if !v.Valid {
		return nil
	}
	kg := float64(v.Int64) / 1000
	return &kg
}

func kgToGrams(kg *float64) sql.NullInt64 {
	if kg == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(math.Round(*kg * 1000)), Valid: true}
}

func metersToKm(v sql.NullInt64) *float64 {
	if !v.Valid {
		return nil
	}
	km := float64(v.Int64) / 1000
	return &km
}

func kmToMeters(km *float64) sql.NullInt64 {
	if km == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(math.Round(*km * 1000)), Valid: true}
}

func intToNull(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullToInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

// --- DTO ---

type setDTO struct {
	ID          int64    `json:"id"`
	ExerciseID  int64    `json:"exercise_id"`
	Position    int64    `json:"position"`
	Role        string   `json:"role"`
	WeightKg    *float64 `json:"weight_kg,omitempty"`
	Reps        *int     `json:"reps,omitempty"`
	DistanceKm  *float64 `json:"distance_km,omitempty"`
	DurationSec *int     `json:"duration_sec,omitempty"`
	Note        string   `json:"note"`
	ClientID    string   `json:"client_id,omitempty"`
}

func toSetDTO(s gen.Set) setDTO {
	return setDTO{
		ID: s.ID, ExerciseID: s.ExerciseID, Position: s.Position, Role: s.Role,
		WeightKg: gramsToKg(s.WeightG), Reps: nullToInt(s.Reps),
		DistanceKm: metersToKm(s.DistanceM), DurationSec: nullToInt(s.DurationSec),
		Note: s.Note, ClientID: s.ClientID.String,
	}
}

type workoutDTO struct {
	ID           int64    `json:"id"`
	Date         string   `json:"date"`
	StartedAt    string   `json:"started_at,omitempty"`
	FinishedAt   string   `json:"finished_at,omitempty"`
	BodyweightKg *float64 `json:"bodyweight_kg,omitempty"`
	Feeling      string   `json:"feeling"`
	Notes        string   `json:"notes"`
	Sets         []setDTO `json:"sets,omitempty"`
}

func toWorkoutDTO(w gen.Workout) workoutDTO {
	return workoutDTO{
		ID: w.ID, Date: w.Date, StartedAt: w.StartedAt.String, FinishedAt: w.FinishedAt.String,
		BodyweightKg: gramsToKg(w.BodyweightG), Feeling: w.Feeling, Notes: w.Notes,
	}
}

// --- workouts ---

func (s *server) handleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Date      string `json:"date"`
		StartedAt string `json:"started_at"`
	}
	if r.ContentLength > 0 && !decodeJSON(w, r, &in) {
		return
	}
	now := s.opts.Now().UTC()
	if in.Date == "" {
		in.Date = now.Format("2006-01-02")
	}
	started := sql.NullString{String: now.Format(time.RFC3339), Valid: true}
	if in.StartedAt != "" {
		started = sql.NullString{String: in.StartedAt, Valid: true}
	}
	wk, err := s.q.CreateWorkout(r.Context(), gen.CreateWorkoutParams{
		UserID:    s.currentUserID(r),
		Date:      in.Date,
		StartedAt: started,
		Feeling:   "",
		Notes:     "",
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, toWorkoutDTO(wk))
}

func (s *server) workoutForUser(w http.ResponseWriter, r *http.Request) (gen.Workout, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Workout{}, false
	}
	wk, err := s.q.GetWorkout(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && wk.UserID != s.currentUserID(r)) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Workout{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return gen.Workout{}, false
	}
	return wk, true
}

func (s *server) handleGetWorkout(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.workoutForUser(w, r)
	if !ok {
		return
	}
	dto := toWorkoutDTO(wk)
	sets, err := s.q.ListSetsForWorkout(r.Context(), wk.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	dto.Sets = make([]setDTO, 0, len(sets))
	for _, st := range sets {
		dto.Sets = append(dto.Sets, toSetDTO(st))
	}
	writeJSON(w, http.StatusOK, dto)
}

func (s *server) handleListWorkouts(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 {
		limit = v
	}
	cursorDate, cursorID := "", int64(0)
	if c := r.URL.Query().Get("cursor"); c != "" {
		d, id, ok := decodeCursor(c)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"cursor": "плохой курсор"})
			return
		}
		cursorDate, cursorID = d, id
	}
	list, err := s.q.ListWorkoutsForUser(r.Context(), gen.ListWorkoutsForUserParams{
		UserID:     s.currentUserID(r),
		CursorDate: cursorDate,
		CursorID:   cursorID,
		Lim:        int64(limit),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	items := make([]workoutDTO, 0, len(list))
	for _, wk := range list {
		items = append(items, toWorkoutDTO(wk))
	}
	next := ""
	if len(list) == limit {
		last := list[len(list)-1]
		next = encodeCursor(last.Date, last.ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": next})
}

func (s *server) handleUpdateWorkout(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.workoutForUser(w, r)
	if !ok {
		return
	}
	var in struct {
		Date         *string  `json:"date"`
		FinishedAt   *string  `json:"finished_at"`
		BodyweightKg *float64 `json:"bodyweight_kg"`
		Feeling      *string  `json:"feeling"`
		Notes        *string  `json:"notes"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	date := wk.Date
	if in.Date != nil {
		date = *in.Date
	}
	finished := wk.FinishedAt
	if in.FinishedAt != nil {
		finished = sql.NullString{String: *in.FinishedAt, Valid: *in.FinishedAt != ""}
	}
	bw := wk.BodyweightG
	if in.BodyweightKg != nil {
		bw = kgToGrams(in.BodyweightKg)
	}
	feeling := wk.Feeling
	if in.Feeling != nil {
		feeling = *in.Feeling
	}
	notes := wk.Notes
	if in.Notes != nil {
		notes = *in.Notes
	}
	upd, err := s.q.UpdateWorkout(r.Context(), gen.UpdateWorkoutParams{
		Date: date, StartedAt: wk.StartedAt, FinishedAt: finished, BodyweightG: bw,
		Feeling: feeling, Notes: notes, UpdatedAt: s.opts.Now().UTC().Format(time.RFC3339), ID: wk.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toWorkoutDTO(upd))
}

func (s *server) handleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.workoutForUser(w, r)
	if !ok {
		return
	}
	if _, err := s.q.DeleteWorkout(r.Context(), gen.DeleteWorkoutParams{ID: wk.ID, UserID: wk.UserID}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- sets ---

type setInput struct {
	ExerciseID  int64    `json:"exercise_id"`
	Role        string   `json:"role"`
	WeightKg    *float64 `json:"weight_kg"`
	Reps        *int     `json:"reps"`
	DistanceKm  *float64 `json:"distance_km"`
	DurationSec *int     `json:"duration_sec"`
	Note        string   `json:"note"`
	ClientID    string   `json:"client_id"`
}

var validRoles = map[string]bool{"warmup": true, "ramp": true, "working": true}

func (s *server) handleAddSet(w http.ResponseWriter, r *http.Request) {
	wk, ok := s.workoutForUser(w, r)
	if !ok {
		return
	}
	var in setInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.ExerciseID == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"exercise_id": "обязательно"})
		return
	}
	if in.Role == "" {
		in.Role = "working"
	}
	if !validRoles[in.Role] {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"role": "warmup|ramp|working"})
		return
	}

	// Идемпотентность: повтор с тем же client_id возвращает существующий подход.
	if in.ClientID != "" {
		if existing, err := s.q.GetSetByClientID(r.Context(), sql.NullString{String: in.ClientID, Valid: true}); err == nil {
			if existing.WorkoutID == wk.ID {
				writeJSON(w, http.StatusOK, toSetDTO(existing))
				return
			}
		}
	}

	pos, err := s.q.NextSetPosition(r.Context(), wk.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	created, err := s.q.CreateSet(r.Context(), gen.CreateSetParams{
		WorkoutID:   wk.ID,
		ExerciseID:  in.ExerciseID,
		Position:    pos,
		Role:        in.Role,
		WeightG:     kgToGrams(in.WeightKg),
		Reps:        intToNull(in.Reps),
		DistanceM:   kmToMeters(in.DistanceKm),
		DurationSec: intToNull(in.DurationSec),
		Note:        in.Note,
		ClientID:    sql.NullString{String: in.ClientID, Valid: in.ClientID != ""},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, toSetDTO(created))
}

func (s *server) setForUser(w http.ResponseWriter, r *http.Request) (gen.Set, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Set{}, false
	}
	st, err := s.q.GetSet(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Set{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return gen.Set{}, false
	}
	wk, err := s.q.GetWorkout(r.Context(), st.WorkoutID)
	if err != nil || wk.UserID != s.currentUserID(r) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Set{}, false
	}
	return st, true
}

func (s *server) handleUpdateSet(w http.ResponseWriter, r *http.Request) {
	st, ok := s.setForUser(w, r)
	if !ok {
		return
	}
	var in setInput
	if !decodeJSON(w, r, &in) {
		return
	}
	role := st.Role
	if in.Role != "" {
		if !validRoles[in.Role] {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"role": "warmup|ramp|working"})
			return
		}
		role = in.Role
	}
	upd, err := s.q.UpdateSet(r.Context(), gen.UpdateSetParams{
		Role: role, WeightG: kgToGrams(in.WeightKg), Reps: intToNull(in.Reps),
		DistanceM: kmToMeters(in.DistanceKm), DurationSec: intToNull(in.DurationSec),
		Note: in.Note, ID: st.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toSetDTO(upd))
}

func (s *server) handleDeleteSet(w http.ResponseWriter, r *http.Request) {
	st, ok := s.setForUser(w, r)
	if !ok {
		return
	}
	if _, err := s.q.DeleteSet(r.Context(), st.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- история упражнения (для столбца «Прошлый») ---

func (s *server) handleExerciseHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return
	}
	sessions := 3
	if v, err := strconv.Atoi(r.URL.Query().Get("sessions")); err == nil && v > 0 && v <= 20 {
		sessions = v
	}
	rows, err := s.q.ExerciseRecentSets(r.Context(), gen.ExerciseRecentSetsParams{
		UserID: s.currentUserID(r), ExerciseID: id, Limit: int64(sessions * 15),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}

	type histSet struct {
		Role        string   `json:"role"`
		WeightKg    *float64 `json:"weight_kg,omitempty"`
		Reps        *int     `json:"reps,omitempty"`
		DistanceKm  *float64 `json:"distance_km,omitempty"`
		DurationSec *int     `json:"duration_sec,omitempty"`
	}
	type session struct {
		Date string    `json:"date"`
		Sets []histSet `json:"sets"`
	}

	out := make([]session, 0, sessions)
	byWorkout := map[int64]int{} // workout_id -> index в out
	for _, row := range rows {
		idx, seen := byWorkout[row.WorkoutID]
		if !seen {
			if len(out) >= sessions {
				continue
			}
			out = append(out, session{Date: row.WorkoutDate})
			idx = len(out) - 1
			byWorkout[row.WorkoutID] = idx
		}
		out[idx].Sets = append(out[idx].Sets, histSet{
			Role: row.Role, WeightKg: gramsToKg(row.WeightG), Reps: nullToInt(row.Reps),
			DistanceKm: metersToKm(row.DistanceM), DurationSec: nullToInt(row.DurationSec),
		})
	}
	writeJSON(w, http.StatusOK, out)
}
