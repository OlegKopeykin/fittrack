package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

// nullInt оборачивает id в sql.NullInt64 (Valid=true).
func nullInt(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

type tokenDTO struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	Revoked    bool   `json:"revoked"`
}

func toTokenDTO(t gen.ApiToken) tokenDTO {
	return tokenDTO{
		ID: t.ID, Name: t.Name, Prefix: t.Prefix, CreatedAt: t.CreatedAt,
		LastUsedAt: t.LastUsedAt.String, ExpiresAt: t.ExpiresAt.String,
		Revoked: t.RevokedAt.Valid,
	}
}

func (s *server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	toks, err := s.q.ListApiTokens(r.Context(), nullInt(s.currentUserID(r)).Int64)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	out := make([]tokenDTO, 0, len(toks))
	for _, t := range toks {
		out = append(out, toTokenDTO(t))
	}
	writeJSON(w, http.StatusOK, out)
}

// handleCreateToken выпускает персональный токен; открытое значение
// показывается ОДИН раз в ответе.
func (s *server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	// Токены выпускаются только из cookie-сессии (не по bearer).
	if isBearer(r) {
		writeError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	var in struct {
		Name          string `json:"name"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"name": "обязательно"})
		return
	}
	plaintext, hash, prefix, err := auth.NewAPIToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	now := s.opts.Now().UTC()
	var expires sql.NullString
	if in.ExpiresInDays > 0 {
		expires = sql.NullString{String: now.AddDate(0, 0, in.ExpiresInDays).Format(time.RFC3339), Valid: true}
	}
	tok, err := s.q.CreateApiToken(r.Context(), gen.CreateApiTokenParams{
		UserID:    s.currentUserID(r),
		Name:      in.Name,
		TokenHash: hash,
		Prefix:    prefix,
		CreatedAt: now.Format(time.RFC3339),
		ExpiresAt: expires,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	dto := toTokenDTO(tok)
	writeJSON(w, http.StatusCreated, map[string]any{"token": plaintext, "info": dto})
}

func (s *server) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return
	}
	rows, err := s.q.RevokeApiToken(r.Context(), gen.RevokeApiTokenParams{
		RevokedAt: sql.NullString{String: s.opts.Now().UTC().Format(time.RFC3339), Valid: true},
		ID:        id,
		UserID:    s.currentUserID(r),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	if rows == 0 {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
