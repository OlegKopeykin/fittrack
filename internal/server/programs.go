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
)

// --- input ---

type prescriptionInput struct {
	ExerciseID   int64    `json:"exercise_id"`
	ExerciseName string   `json:"exercise_name"`
	Sets         int      `json:"sets"`
	RepMin       *int     `json:"rep_min"`
	RepMax       *int     `json:"rep_max"`
	WeightMinKg  *float64 `json:"weight_min_kg"`
	WeightMaxKg  *float64 `json:"weight_max_kg"`
	RestSec      *int     `json:"rest_sec"`
	Tempo        string   `json:"tempo"`
	Notes        string   `json:"notes"`
}

type programDayInput struct {
	Name      string              `json:"name"`
	Notes     string              `json:"notes"`
	Exercises []prescriptionInput `json:"exercises"`
}

type programInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Days        []programDayInput `json:"days"`
}

// --- output ---

type prescriptionDTO struct {
	ID          int64    `json:"id"`
	ExerciseID  int64    `json:"exercise_id"`
	Position    int64    `json:"position"`
	Sets        int64    `json:"sets"`
	RepMin      *int     `json:"rep_min,omitempty"`
	RepMax      *int     `json:"rep_max,omitempty"`
	WeightMinKg *float64 `json:"weight_min_kg,omitempty"`
	WeightMaxKg *float64 `json:"weight_max_kg,omitempty"`
	RestSec     *int     `json:"rest_sec,omitempty"`
	Tempo       string   `json:"tempo,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

func toPrescriptionDTO(p gen.Prescription) prescriptionDTO {
	return prescriptionDTO{
		ID: p.ID, ExerciseID: p.ExerciseID, Position: p.Position, Sets: p.Sets,
		RepMin: nullToInt(p.RepMin), RepMax: nullToInt(p.RepMax),
		WeightMinKg: gramsToKg(p.WeightMinG), WeightMaxKg: gramsToKg(p.WeightMaxG),
		RestSec: nullToInt(p.RestSec), Tempo: p.Tempo, Notes: p.Notes,
	}
}

type programDayDTO struct {
	ID        int64             `json:"id"`
	Position  int64             `json:"position"`
	Name      string            `json:"name"`
	Notes     string            `json:"notes,omitempty"`
	Exercises []prescriptionDTO `json:"exercises"`
}

type programDTO struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Archived    bool            `json:"archived,omitempty"`
	Days        []programDayDTO `json:"days,omitempty"`
}

func (s *server) handleCreateProgram(w http.ResponseWriter, r *http.Request) {
	var in programInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"name": "обязательно"})
		return
	}

	// ВАЖНО: ответ нельзя писать при открытой транзакции — scs сохраняет
	// сессию в момент записи ответа, а единственная коннекция занята
	// транзакцией (дедлок). Вся работа с tx — внутри createProgram.
	progID, status, code, fields := s.createProgram(r, s.currentUserID(r), in)
	if status != http.StatusCreated {
		writeError(w, status, code, fields)
		return
	}
	prog, err := s.q.GetProgram(r.Context(), progID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	dto, err := s.loadProgram(r, prog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}

// createProgram создаёт программу целиком в одной транзакции и закрывает её
// до возврата (ответ вызывающий пишет уже после).
func (s *server) createProgram(r *http.Request, uid int64, in programInput) (int64, int, string, map[string]string) {
	ctx := r.Context()
	all, err := s.q.ListExercisesForUser(ctx, nullInt(uid))
	if err != nil {
		return 0, http.StatusInternalServerError, "internal", nil
	}
	idByName := make(map[string]int64, len(all))
	for _, e := range all {
		idByName[e.Name] = e.ID
	}

	tx, err := s.opts.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, http.StatusInternalServerError, "internal", nil
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.q.WithTx(tx)
	now := s.opts.Now().UTC().Format(time.RFC3339)

	prog, err := qtx.CreateProgram(ctx, gen.CreateProgramParams{
		UserID: uid, Name: strings.TrimSpace(in.Name), Description: in.Description, CreatedAt: now,
	})
	if err != nil {
		return 0, http.StatusInternalServerError, "internal", nil
	}

	for di, d := range in.Days {
		day, err := qtx.CreateProgramDay(ctx, gen.CreateProgramDayParams{
			ProgramID: prog.ID, Position: int64(di), Name: d.Name, Notes: d.Notes,
		})
		if err != nil {
			return 0, http.StatusInternalServerError, "internal", nil
		}
		for pi, p := range d.Exercises {
			exID := p.ExerciseID
			if exID == 0 && p.ExerciseName != "" {
				id, ok := idByName[strings.TrimSpace(p.ExerciseName)]
				if !ok {
					return 0, http.StatusBadRequest, "invalid_input",
						map[string]string{"exercise": "неизвестное упражнение: " + p.ExerciseName}
				}
				exID = id
			}
			if exID == 0 {
				return 0, http.StatusBadRequest, "invalid_input",
					map[string]string{"exercise": "нужен exercise_id или exercise_name"}
			}
			if _, err := qtx.CreatePrescription(ctx, gen.CreatePrescriptionParams{
				ProgramDayID: day.ID, ExerciseID: exID, Position: int64(pi),
				Sets: int64(p.Sets), RepMin: intToNull(p.RepMin), RepMax: intToNull(p.RepMax),
				WeightMinG: kgToGrams(p.WeightMinKg), WeightMaxG: kgToGrams(p.WeightMaxKg),
				RestSec: intToNull(p.RestSec), Tempo: p.Tempo, Notes: p.Notes,
			}); err != nil {
				return 0, http.StatusInternalServerError, "internal", nil
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, http.StatusInternalServerError, "internal", nil
	}
	return prog.ID, http.StatusCreated, "", nil
}

func (s *server) loadProgram(r *http.Request, prog gen.Program) (programDTO, error) {
	dto := programDTO{ID: prog.ID, Name: prog.Name, Description: prog.Description}
	days, err := s.q.ListProgramDays(r.Context(), prog.ID)
	if err != nil {
		return dto, err
	}
	for _, d := range days {
		dd := programDayDTO{ID: d.ID, Position: d.Position, Name: d.Name, Notes: d.Notes}
		rx, err := s.q.ListPrescriptionsForDay(r.Context(), d.ID)
		if err != nil {
			return dto, err
		}
		dd.Exercises = make([]prescriptionDTO, 0, len(rx))
		for _, p := range rx {
			dd.Exercises = append(dd.Exercises, toPrescriptionDTO(p))
		}
		dto.Days = append(dto.Days, dd)
	}
	return dto, nil
}

func (s *server) handleListPrograms(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	var progs []gen.Program
	var err error
	if r.URL.Query().Get("archived") == "1" {
		progs, err = s.q.ListArchivedProgramsForUser(r.Context(), uid)
	} else {
		progs, err = s.q.ListProgramsForUser(r.Context(), uid)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	out := make([]programDTO, 0, len(progs))
	for _, p := range progs {
		out = append(out, programDTO{ID: p.ID, Name: p.Name, Description: p.Description, Archived: p.ArchivedAt.Valid})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) setProgramArchived(w http.ResponseWriter, r *http.Request, archived bool) {
	prog, ok := s.programForUser(w, r)
	if !ok {
		return
	}
	var at sql.NullString
	if archived {
		at = sql.NullString{String: s.opts.Now().UTC().Format(time.RFC3339), Valid: true}
	}
	if _, err := s.q.SetProgramArchived(r.Context(), gen.SetProgramArchivedParams{
		ArchivedAt: at, ID: prog.ID, UserID: prog.UserID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleArchiveProgram(w http.ResponseWriter, r *http.Request) {
	s.setProgramArchived(w, r, true)
}

func (s *server) handleUnarchiveProgram(w http.ResponseWriter, r *http.Request) {
	s.setProgramArchived(w, r, false)
}

func (s *server) programForUser(w http.ResponseWriter, r *http.Request) (gen.Program, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Program{}, false
	}
	prog, err := s.q.GetProgram(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && prog.UserID != s.currentUserID(r)) {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return gen.Program{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return gen.Program{}, false
	}
	return prog, true
}

func (s *server) handleGetProgram(w http.ResponseWriter, r *http.Request) {
	prog, ok := s.programForUser(w, r)
	if !ok {
		return
	}
	dto, err := s.loadProgram(r, prog)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

func (s *server) handleDeleteProgram(w http.ResponseWriter, r *http.Request) {
	prog, ok := s.programForUser(w, r)
	if !ok {
		return
	}
	if _, err := s.q.DeleteProgram(r.Context(), gen.DeleteProgramParams{ID: prog.ID, UserID: prog.UserID}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
