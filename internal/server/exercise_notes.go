package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

// Персональная заметка пользователя к упражнению (работает и для глобальных).

func (s *server) handleGetExerciseNote(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	note, err := s.q.GetUserExerciseNote(r.Context(), gen.GetUserExerciseNoteParams{
		UserID: s.currentUserID(r), ExerciseID: ex.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		note = ""
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"note": note})
}

func (s *server) handleSetExerciseNote(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	var in struct {
		Note string `json:"note"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if err := s.q.UpsertUserExerciseNote(r.Context(), gen.UpsertUserExerciseNoteParams{
		UserID: s.currentUserID(r), ExerciseID: ex.ID, Note: in.Note,
		UpdatedAt: s.opts.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"note": in.Note})
}
