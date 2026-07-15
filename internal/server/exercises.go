package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/exercises"
)

var validKinds = map[string]bool{
	"compound": true, "isolation": true, "isometric": true, "bodyweight": true, "cardio": true,
}

type muscleGroupDTO struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	NameRu    string `json:"name_ru"`
	WeeklyMEV *int64 `json:"weekly_mev,omitempty"`
	WeeklyMAV *int64 `json:"weekly_mav,omitempty"`
}

func toMuscleGroupDTO(g gen.MuscleGroup) muscleGroupDTO {
	d := muscleGroupDTO{ID: g.ID, Slug: g.Slug, NameRu: g.NameRu}
	if g.WeeklyMev.Valid {
		d.WeeklyMEV = &g.WeeklyMev.Int64
	}
	if g.WeeklyMav.Valid {
		d.WeeklyMAV = &g.WeeklyMav.Int64
	}
	return d
}

type exerciseDTO struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	MuscleGroupID  int64    `json:"muscle_group_id"`
	Kind           string   `json:"kind"`
	PerArm         bool     `json:"per_arm"`
	TechniqueNotes string   `json:"technique_notes"`
	Global         bool     `json:"global"`
	Archived       bool     `json:"archived"`
	Aliases        []string `json:"aliases,omitempty"`
}

func toExerciseDTO(e gen.Exercise) exerciseDTO {
	return exerciseDTO{
		ID:             e.ID,
		Name:           e.Name,
		MuscleGroupID:  e.MuscleGroupID,
		Kind:           e.Kind,
		PerArm:         e.PerArm == 1,
		TechniqueNotes: e.TechniqueNotes,
		Global:         !e.OwnerID.Valid,
		Archived:       e.ArchivedAt.Valid,
	}
}

func (s *server) handleListMuscleGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.q.ListMuscleGroups(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	out := make([]muscleGroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, toMuscleGroupDTO(g))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) handleListExercises(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	all, err := s.q.ListExercisesForUser(r.Context(), nullInt(uid))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}

	includeArchived := r.URL.Query().Get("include_archived") == "1"
	groupFilter := r.URL.Query().Get("muscle_group")
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	// Поиск: по алиасам через SQL (alias_norm уже нормализован), по имени —
	// в Go (SQLite lower() не понижает регистр кириллицы).
	normQ := exercises.Normalize(q)
	aliasMatch := map[int64]bool{}
	if q != "" {
		ids, err := s.q.SearchAliasExerciseIDs(r.Context(), "%"+normQ+"%")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", nil)
			return
		}
		for _, id := range ids {
			aliasMatch[id] = true
		}
	}

	// Слаги групп → id (для фильтра).
	var groupID int64
	if groupFilter != "" {
		g, err := s.q.GetMuscleGroupBySlug(r.Context(), groupFilter)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"muscle_group": "неизвестная группа"})
			return
		}
		groupID = g.ID
	}

	out := make([]exerciseDTO, 0, len(all))
	for _, e := range all {
		if !includeArchived && e.ArchivedAt.Valid {
			continue
		}
		if groupID != 0 && e.MuscleGroupID != groupID {
			continue
		}
		if q != "" && !aliasMatch[e.ID] && !strings.Contains(exercises.Normalize(e.Name), normQ) {
			continue
		}
		out = append(out, toExerciseDTO(e))
	}
	writeJSON(w, http.StatusOK, out)
}

type exerciseInput struct {
	Name           string `json:"name"`
	MuscleGroup    string `json:"muscle_group"` // slug
	Kind           string `json:"kind"`
	PerArm         bool   `json:"per_arm"`
	TechniqueNotes string `json:"technique_notes"`
}

func (in exerciseInput) validate() map[string]string {
	f := map[string]string{}
	if strings.TrimSpace(in.Name) == "" {
		f["name"] = "обязательно"
	}
	if !validKinds[in.Kind] {
		f["kind"] = "compound|isolation|isometric|bodyweight|cardio"
	}
	if in.MuscleGroup == "" {
		f["muscle_group"] = "обязательно"
	}
	return f
}

func (s *server) handleCreateExercise(w http.ResponseWriter, r *http.Request) {
	var in exerciseInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if f := in.validate(); len(f) > 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", f)
		return
	}
	g, err := s.q.GetMuscleGroupBySlug(r.Context(), in.MuscleGroup)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"muscle_group": "неизвестная группа"})
		return
	}
	ex, err := s.q.CreateExercise(r.Context(), gen.CreateExerciseParams{
		OwnerID:        nullInt(s.currentUserID(r)),
		Name:           strings.TrimSpace(in.Name),
		MuscleGroupID:  g.ID,
		Kind:           in.Kind,
		PerArm:         boolInt(in.PerArm),
		TechniqueNotes: in.TechniqueNotes,
		CreatedAt:      s.opts.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "conflict", map[string]string{"name": "уже есть"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, toExerciseDTO(ex))
}

// handleBulkExercises — массовая заливка упражнений (машинный API).
// Идемпотентна по имени в пределах пользователя: существующие пропускаются.
func (s *server) handleBulkExercises(w http.ResponseWriter, r *http.Request) {
	var items []exerciseInput
	if !decodeJSON(w, r, &items) {
		return
	}
	uid := s.currentUserID(r)
	now := s.opts.Now().UTC().Format(time.RFC3339)

	existing, err := s.q.ListExercisesForUser(r.Context(), nullInt(uid))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	have := map[string]bool{}
	for _, e := range existing {
		have[e.Name] = true
	}

	created, skipped := 0, 0
	for i, in := range items {
		if f := in.validate(); len(f) > 0 {
			writeError(w, http.StatusBadRequest, "invalid_input",
				map[string]string{"item": strconv.Itoa(i)})
			return
		}
		if have[strings.TrimSpace(in.Name)] {
			skipped++
			continue
		}
		g, err := s.q.GetMuscleGroupBySlug(r.Context(), in.MuscleGroup)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input",
				map[string]string{"item": strconv.Itoa(i), "muscle_group": in.MuscleGroup})
			return
		}
		if _, err := s.q.CreateExercise(r.Context(), gen.CreateExerciseParams{
			OwnerID:        nullInt(uid),
			Name:           strings.TrimSpace(in.Name),
			MuscleGroupID:  g.ID,
			Kind:           in.Kind,
			PerArm:         boolInt(in.PerArm),
			TechniqueNotes: in.TechniqueNotes,
			CreatedAt:      now,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal", nil)
			return
		}
		have[strings.TrimSpace(in.Name)] = true
		created++
	}
	writeJSON(w, http.StatusOK, map[string]int{"created": created, "skipped": skipped})
}

// exerciseOwnedOrGlobal загружает упражнение, доступное пользователю
// (глобальное или своё). Возвращает false + пишет ошибку, если недоступно.
func (s *server) exerciseForUser(w http.ResponseWriter, r *http.Request) (gen.Exercise, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Exercise{}, false
	}
	ex, err := s.q.GetExercise(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Exercise{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return gen.Exercise{}, false
	}
	if ex.OwnerID.Valid && ex.OwnerID.Int64 != s.currentUserID(r) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Exercise{}, false
	}
	return ex, true
}

func (s *server) handleGetExercise(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	dto := toExerciseDTO(ex)
	aliases, err := s.q.ListAliasesForExercise(r.Context(), ex.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	for _, a := range aliases {
		dto.Aliases = append(dto.Aliases, a.Alias)
	}
	writeJSON(w, http.StatusOK, dto)
}

func (s *server) handleUpdateExercise(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	if !ex.OwnerID.Valid {
		writeError(w, http.StatusForbidden, "forbidden", map[string]string{"exercise": "глобальное нельзя менять"})
		return
	}
	var in exerciseInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if f := in.validate(); len(f) > 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", f)
		return
	}
	g, err := s.q.GetMuscleGroupBySlug(r.Context(), in.MuscleGroup)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"muscle_group": "неизвестная группа"})
		return
	}
	upd, err := s.q.UpdateExercise(r.Context(), gen.UpdateExerciseParams{
		Name:           strings.TrimSpace(in.Name),
		MuscleGroupID:  g.ID,
		Kind:           in.Kind,
		PerArm:         boolInt(in.PerArm),
		TechniqueNotes: in.TechniqueNotes,
		ID:             ex.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toExerciseDTO(upd))
}

func (s *server) handleArchiveExercise(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	if !ex.OwnerID.Valid {
		writeError(w, http.StatusForbidden, "forbidden", map[string]string{"exercise": "глобальное нельзя удалить"})
		return
	}
	if _, err := s.q.ArchiveExercise(r.Context(), gen.ArchiveExerciseParams{
		ArchivedAt: sql.NullString{String: s.opts.Now().UTC().Format(time.RFC3339), Valid: true},
		ID:         ex.ID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleAddAlias(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	if !ex.OwnerID.Valid {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	var in struct {
		Alias string `json:"alias"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	norm := exercises.Normalize(in.Alias)
	if norm == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"alias": "обязательно"})
		return
	}
	n, _ := s.q.CountAlias(r.Context(), gen.CountAliasParams{ExerciseID: ex.ID, AliasNorm: norm})
	if n > 0 {
		writeError(w, http.StatusConflict, "conflict", nil)
		return
	}
	a, err := s.q.AddAlias(r.Context(), gen.AddAliasParams{
		ExerciseID: ex.ID, Alias: strings.TrimSpace(in.Alias), AliasNorm: norm,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": a.ID, "alias": a.Alias})
}
