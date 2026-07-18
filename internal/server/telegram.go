package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/backup"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
)

var validFrequency = map[string]bool{"daily": true, "weekly": true, "monthly": true}

type telegramDTO struct {
	Configured  bool   `json:"configured"`
	BotUsername string `json:"bot_username,omitempty"`
	ChatLinked  bool   `json:"chat_linked"`
	Enabled     bool   `json:"enabled"`
	Frequency   string `json:"frequency"`
	LastSentAt  string `json:"last_sent_at,omitempty"`
}

func toTelegramDTO(s gen.TelegramSetting) telegramDTO {
	return telegramDTO{
		Configured:  s.BotToken != "",
		BotUsername: s.BotUsername,
		ChatLinked:  s.ChatID != "",
		Enabled:     s.Enabled != 0,
		Frequency:   s.Frequency,
		LastSentAt:  s.LastSentAt,
	}
}

// tgSettings возвращает настройки пользователя или дефолт (без строки в БД).
func (s *server) tgSettings(r *http.Request, uid int64) (gen.TelegramSetting, error) {
	st, err := s.q.GetTelegramSettings(r.Context(), uid)
	if errors.Is(err, sql.ErrNoRows) {
		return gen.TelegramSetting{UserID: uid, Frequency: "daily"}, nil
	}
	return st, err
}

func (s *server) saveTg(r *http.Request, st gen.TelegramSetting) error {
	return s.q.UpsertTelegramSettings(r.Context(), gen.UpsertTelegramSettingsParams{
		UserID: st.UserID, BotToken: st.BotToken, BotUsername: st.BotUsername, ChatID: st.ChatID,
		Enabled: st.Enabled, Frequency: st.Frequency, LastSentAt: st.LastSentAt,
		UpdatedAt: s.opts.Now().UTC().Format(time.RFC3339),
	})
}

func (s *server) handleGetTelegram(w http.ResponseWriter, r *http.Request) {
	st, err := s.tgSettings(r, s.currentUserID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toTelegramDTO(st))
}

func (s *server) handleSetTelegram(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	var in struct {
		BotToken  *string `json:"bot_token"`
		Frequency *string `json:"frequency"`
		Enabled   *bool   `json:"enabled"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	st, err := s.tgSettings(r, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}

	if in.BotToken != nil {
		token := strings.TrimSpace(*in.BotToken)
		if token == "" {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"bot_token": "обязательно"})
			return
		}
		username, err := s.opts.Telegram.GetMe(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"bot_token": "токен не принят Telegram"})
			return
		}
		// Смена токена сбрасывает привязку чата.
		if token != st.BotToken {
			st.ChatID = ""
		}
		st.BotToken = token
		st.BotUsername = username
	}
	if in.Frequency != nil {
		if !validFrequency[*in.Frequency] {
			writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"frequency": "daily|weekly|monthly"})
			return
		}
		st.Frequency = *in.Frequency
	}
	if in.Enabled != nil {
		if *in.Enabled {
			st.Enabled = 1
		} else {
			st.Enabled = 0
		}
	}
	if err := s.saveTg(r, st); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toTelegramDTO(st))
}

func (s *server) handleLinkTelegram(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	st, err := s.tgSettings(r, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	if st.BotToken == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"bot_token": "сначала укажите токен"})
		return
	}
	chatID, err := s.opts.Telegram.ResolveChatID(r.Context(), st.BotToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "not_linked", map[string]string{"chat": "откройте бота и нажмите Start"})
		return
	}
	st.ChatID = chatID
	if err := s.saveTg(r, st); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, toTelegramDTO(st))
}

func (s *server) handleDeleteTelegram(w http.ResponseWriter, r *http.Request) {
	if err := s.q.DeleteTelegramSettings(r.Context(), s.currentUserID(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleTestTelegram(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	st, err := s.tgSettings(r, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	if st.BotToken == "" || st.ChatID == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"chat": "бот не подключён"})
		return
	}
	user, err := s.q.GetUserByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	now := s.opts.Now().UTC().Format(time.RFC3339)
	data, filename, err := buildBackupFile(r, s.q, uid, user.Username, s.opts.Version, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	if err := s.opts.Telegram.SendDocument(r.Context(), st.BotToken, st.ChatID, filename, data, ""); err != nil {
		writeError(w, http.StatusBadGateway, "telegram_failed", map[string]string{"telegram": err.Error()})
		return
	}
	st.LastSentAt = now
	_ = s.saveTg(r, st)
	w.WriteHeader(http.StatusNoContent)
}

// handleExportDownload отдаёт JSON-бэкап лога для ручного скачивания.
func (s *server) handleExportDownload(w http.ResponseWriter, r *http.Request) {
	uid := s.currentUserID(r)
	user, err := s.q.GetUserByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	now := s.opts.Now().UTC().Format(time.RFC3339)
	data, filename, err := buildBackupFile(r, s.q, uid, user.Username, s.opts.Version, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, _ = w.Write(data)
}

// handleImport восстанавливает лог из загруженного JSON-бэкапа.
func (s *server) handleImport(w http.ResponseWriter, r *http.Request) {
	var exp backup.Export
	if !decodeJSON(w, r, &exp) {
		return
	}
	if len(exp.Workouts) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", map[string]string{"file": "в файле нет тренировок"})
		return
	}
	now := s.opts.Now().UTC().Format(time.RFC3339)
	imported, skipped, err := backup.Restore(r.Context(), s.opts.DB, s.q, s.currentUserID(r), now, exp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"imported": imported, "skipped": skipped})
}

func buildBackupFile(r *http.Request, q *gen.Queries, uid int64, username, version, now string) ([]byte, string, error) {
	exp, err := backup.Build(r.Context(), q, uid, username, version, now)
	if err != nil {
		return nil, "", err
	}
	data, err := backup.Marshal(exp)
	if err != nil {
		return nil, "", err
	}
	filename := "fittrack-" + username + "-" + now[:10] + ".json"
	return data, filename, nil
}
