package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

const sessionUserKey = "userID"

const minPasswordLen = 8

type ctxKey int

const bearerUserKey ctxKey = iota

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)

// dummyHash выравнивает время ответа login для несуществующих пользователей.
var dummyHash = func() string {
	h, err := auth.HashPassword("fittrack-dummy")
	if err != nil {
		panic(err)
	}
	return h
}()

type userDTO struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func toUserDTO(u gen.User) userDTO {
	return userDTO{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, Role: u.Role}
}

type inviteDTO struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`
	UsedAt    string `json:"used_at,omitempty"`
}

func toInviteDTO(i gen.Invite) inviteDTO {
	return inviteDTO{
		ID: i.ID, Code: i.Code, Role: i.Role, CreatedAt: i.CreatedAt,
		ExpiresAt: i.ExpiresAt.String, UsedAt: i.UsedAt.String,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string, fields map[string]string) {
	body := map[string]any{"code": code, "message": http.StatusText(status)}
	if len(fields) > 0 {
		body["fields"] = fields
	}
	writeJSON(w, status, map[string]any{"error": body})
}

// clientIP — реальный адрес клиента: X-Real-IP от доверенного loopback-прокси,
// иначе host из RemoteAddr.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// csrfGuard защищает cookie-мутации: Origin (если прислан) обязан совпадать
// с PublicOrigin или Host запроса; тело мутаций — только application/json.
func (s *server) csrfGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			next.ServeHTTP(w, r)
			return
		}

		// Машинный API (bearer) не подвержен CSRF — куки не участвуют.
		if isBearer(r) {
			next.ServeHTTP(w, r)
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" && !s.originAllowed(origin, r.Host) {
			writeError(w, http.StatusForbidden, "forbidden", nil)
			return
		}

		// Отклоняем только «простые» content-type, которые кросс-сайтовая
		// <form> может отправить без preflight (единственный CSRF-вектор).
		// JSON, image/* и прочие не-простые типы проходят (их защищает
		// CORS-preflight + проверка Origin выше).
		if r.ContentLength > 0 {
			mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			switch mt {
			case "", "text/plain", "application/x-www-form-urlencoded", "multipart/form-data":
				writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", nil)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) originAllowed(origin, host string) bool {
	if s.opts.PublicOrigin != "" && origin == s.opts.PublicOrigin {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == host
}

// currentUserID возвращает id пользователя из bearer-токена (если запрос
// пришёл по машинному API) либо из cookie-сессии.
func (s *server) currentUserID(r *http.Request) int64 {
	if v, ok := r.Context().Value(bearerUserKey).(int64); ok && v != 0 {
		return v
	}
	return s.sessions.GetInt64(r.Context(), sessionUserKey)
}

// isBearer сообщает, аутентифицирован ли запрос по токену (а не по cookie).
func isBearer(r *http.Request) bool {
	v, ok := r.Context().Value(bearerUserKey).(int64)
	return ok && v != 0
}

// resolveBearer: если есть заголовок Authorization: Bearer fit_…, проверяет
// токен (не отозван, не истёк), кладёт userID в контекст и обновляет
// last_used_at. Невалидный токен — 401. Отсутствие заголовка — пропуск к cookie.
func (s *server) resolveBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		rec, err := s.q.GetApiTokenByHash(r.Context(), auth.HashToken(token))
		now := s.opts.Now().UTC().Format(time.RFC3339)
		if err != nil || rec.RevokedAt.Valid || (rec.ExpiresAt.Valid && rec.ExpiresAt.String <= now) {
			writeError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		_ = s.q.TouchApiToken(r.Context(), gen.TouchApiTokenParams{
			LastUsedAt: sql.NullString{String: now, Valid: true},
			ID:         rec.ID,
		})
		ctx := context.WithValue(r.Context(), bearerUserKey, rec.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.currentUserID(r) == 0 {
			writeError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) requireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.q.GetUserByID(r.Context(), s.currentUserID(r))
		if err != nil || user.Role != "owner" {
			writeError(w, http.StatusForbidden, "forbidden", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", nil)
		return false
	}
	return true
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields := map[string]string{}
	if !usernameRe.MatchString(req.Username) {
		fields["username"] = "3–32 символа: латиница, цифры, . _ -"
	}
	if utf8.RuneCountInString(req.Password) < minPasswordLen {
		fields["password"] = "минимум 8 символов"
	}
	if len(fields) > 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", fields)
		return
	}

	key := clientIP(r) + "|" + strings.ToLower(req.Username)
	if s.limiter.tooMany(key) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", nil)
		return
	}

	// ВАЖНО: HTTP-ответ нельзя писать при открытой транзакции — scs
	// коммитит сессию в момент записи ответа и берёт коннекцию из пула,
	// а пул из одной коннекции занят транзакцией (дедлок). Вся работа
	// с транзакцией — внутри s.register, ответ — строго после.
	user, status, code, errFields := s.register(r, req, key)
	if status != http.StatusCreated {
		writeError(w, status, code, errFields)
		return
	}

	if err := s.sessions.RenewToken(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	s.sessions.Put(r.Context(), sessionUserKey, user.ID)
	writeJSON(w, http.StatusCreated, toUserDTO(user))
}

type registerRequest struct {
	InviteCode string `json:"invite_code"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// register выполняет регистрацию в одной транзакции (инвайт гасится тем же
// коммитом, что и создание пользователя) и закрывает её до возврата.
func (s *server) register(r *http.Request, req registerRequest, limiterKey string) (gen.User, int, string, map[string]string) {
	ctx := r.Context()
	now := s.opts.Now().UTC().Format(time.RFC3339)

	tx, err := s.opts.DB.BeginTx(ctx, nil)
	if err != nil {
		return gen.User{}, http.StatusInternalServerError, "internal", nil
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.q.WithTx(tx)

	inv, err := qtx.GetInviteByCode(ctx, req.InviteCode)
	if errors.Is(err, sql.ErrNoRows) || err == nil && inviteUnusable(inv, now) {
		s.limiter.fail(limiterKey)
		return gen.User{}, http.StatusBadRequest, "invalid_invite", nil
	}
	if err != nil {
		return gen.User{}, http.StatusInternalServerError, "internal", nil
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return gen.User{}, http.StatusInternalServerError, "internal", nil
	}

	user, err := qtx.CreateUser(ctx, gen.CreateUserParams{
		Username:     req.Username,
		PasswordHash: hash,
		DisplayName:  "",
		Role:         inv.Role,
		CreatedAt:    now,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return gen.User{}, http.StatusConflict, "conflict", map[string]string{"username": "занят"}
		}
		return gen.User{}, http.StatusInternalServerError, "internal", nil
	}

	rows, err := qtx.ConsumeInvite(ctx, gen.ConsumeInviteParams{
		UsedBy:    sql.NullInt64{Int64: user.ID, Valid: true},
		UsedAt:    sql.NullString{String: now, Valid: true},
		Code:      req.InviteCode,
		ExpiresAt: sql.NullString{String: now, Valid: true},
	})
	if err != nil || rows != 1 {
		return gen.User{}, http.StatusBadRequest, "invalid_invite", nil
	}
	if err := tx.Commit(); err != nil {
		return gen.User{}, http.StatusInternalServerError, "internal", nil
	}
	return user, http.StatusCreated, "", nil
}

func inviteUnusable(inv gen.Invite, now string) bool {
	if inv.UsedBy.Valid {
		return true
	}
	return inv.ExpiresAt.Valid && inv.ExpiresAt.String <= now
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	key := clientIP(r) + "|" + strings.ToLower(req.Username)
	if s.limiter.tooMany(key) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", nil)
		return
	}

	user, err := s.q.GetUserByUsername(r.Context(), req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		auth.VerifyPassword(req.Password, dummyHash) // выравнивание времени ответа
		s.limiter.fail(key)
		writeError(w, http.StatusUnauthorized, "invalid_credentials", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		s.limiter.fail(key)
		writeError(w, http.StatusUnauthorized, "invalid_credentials", nil)
		return
	}

	s.limiter.reset(key)
	if err := s.sessions.RenewToken(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	s.sessions.Put(r.Context(), sessionUserKey, user.ID)
	writeJSON(w, http.StatusOK, toUserDTO(user))
}

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.Destroy(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.q.GetUserByID(r.Context(), s.currentUserID(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(user))
}

func (s *server) handleListInvites(w http.ResponseWriter, r *http.Request) {
	invites, err := s.q.ListInvites(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	out := make([]inviteDTO, 0, len(invites))
	for _, inv := range invites {
		out = append(out, toInviteDTO(inv))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role          string `json:"role"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Role != "user" && req.Role != "owner" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"role": "user или owner"})
		return
	}

	code, err := auth.NewInviteCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	now := s.opts.Now().UTC()
	var expires sql.NullString
	if req.ExpiresInDays > 0 {
		expires = sql.NullString{String: now.AddDate(0, 0, req.ExpiresInDays).Format(time.RFC3339), Valid: true}
	}
	inv, err := s.q.CreateInvite(r.Context(), gen.CreateInviteParams{
		Code:      code,
		Role:      req.Role,
		CreatedBy: sql.NullInt64{Int64: s.currentUserID(r), Valid: true},
		CreatedAt: now.Format(time.RFC3339),
		ExpiresAt: expires,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusCreated, toInviteDTO(inv))
}

func (s *server) handleDeleteInvite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", nil)
		return
	}
	rows, err := s.q.DeleteUnusedInvite(r.Context(), id)
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
