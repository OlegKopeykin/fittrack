package server

import (
	"io"
	"net/http"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

const maxImageBytes = 5 << 20 // 5 MiB

var allowedImageTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/webp": true, "image/gif": true,
}

// handleSetExerciseImage принимает картинку тела запроса (Content-Type image/*)
// и сохраняет её в БД (blob). Только для своих упражнений.
func (s *server) handleSetExerciseImage(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	if !ex.OwnerID.Valid {
		writeError(w, http.StatusForbidden, "forbidden", map[string]string{"exercise": "глобальное нельзя менять"})
		return
	}
	ct := r.Header.Get("Content-Type")
	if !allowedImageTypes[ct] {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type",
			map[string]string{"content_type": "png, jpeg, webp или gif"})
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxImageBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", nil)
		return
	}
	if len(body) == 0 || len(body) > maxImageBytes {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"image": "пусто или больше 5 МБ"})
		return
	}
	if err := s.q.SetExerciseImage(r.Context(), gen.SetExerciseImageParams{
		ExerciseID: ex.ID, ContentType: ct, Bytes: body,
		UpdatedAt: s.opts.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGetExerciseImage отдаёт сохранённую картинку упражнения.
func (s *server) handleGetExerciseImage(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	img, err := s.q.GetExerciseImage(r.Context(), ex.ID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", img.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	_, _ = w.Write(img.Bytes)
}

func (s *server) handleDeleteExerciseImage(w http.ResponseWriter, r *http.Request) {
	ex, ok := s.exerciseForUser(w, r)
	if !ok {
		return
	}
	if !ex.OwnerID.Valid {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	if _, err := s.q.DeleteExerciseImage(r.Context(), ex.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
