package server

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/telegram"
)

// Options — зависимости HTTP-сервера.
type Options struct {
	DB     *sql.DB
	Static fs.FS
	// PublicOrigin — внешний origin приложения (https://host); используется
	// CSRF-проверкой Origin. Пусто — проверяется совпадение с Host запроса.
	PublicOrigin string
	// SecureCookies — ставить Secure на сессионную куку (прод за TLS).
	SecureCookies bool
	// Now — источник времени (тесты подменяют).
	Now func() time.Time
	// Telegram — клиент Bot API (тесты подменяют фейком). nil — реальный HTTP.
	Telegram telegram.Client
	// Version — версия приложения, попадает в экспорт.
	Version string
}

type server struct {
	opts     Options
	q        *gen.Queries
	sessions *scs.SessionManager
	limiter  *rateLimiter
}

// New собирает HTTP-роутер приложения: health-проба, API под /api/v1
// и раздача SPA.
func New(opts Options) http.Handler {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Telegram == nil {
		opts.Telegram = telegram.New()
	}

	sessions := scs.New()
	sessions.Store = sqlite3store.New(opts.DB)
	sessions.Lifetime = 90 * 24 * time.Hour
	sessions.IdleTimeout = 30 * 24 * time.Hour
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.Secure = opts.SecureCookies
	sessions.Cookie.SameSite = http.SameSiteLaxMode

	s := &server{
		opts:     opts,
		q:        gen.New(opts.DB),
		sessions: sessions,
		limiter:  newRateLimiter(10, 15*time.Minute, opts.Now),
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer, securityHeaders)

	r.Get("/healthz", s.handleHealthz)
	r.Route("/api", func(api chi.Router) {
		api.NotFound(jsonError(http.StatusNotFound, "not_found"))
		api.Route("/v1", func(v1 chi.Router) {
			v1.Use(s.sessions.LoadAndSave, s.resolveBearer, s.csrfGuard)
			v1.Post("/auth/register", s.handleRegister)
			v1.Post("/auth/login", s.handleLogin)
			v1.Group(func(priv chi.Router) {
				priv.Use(s.requireAuth)
				priv.Post("/auth/logout", s.handleLogout)
				priv.Get("/auth/me", s.handleMe)

				// Каталог упражнений (cookie или bearer).
				priv.Get("/muscle-groups", s.handleListMuscleGroups)
				priv.Get("/exercises", s.handleListExercises)
				priv.Post("/exercises", s.handleCreateExercise)
				priv.Post("/exercises/bulk", s.handleBulkExercises)
				priv.Get("/exercises/{id}", s.handleGetExercise)
				priv.Patch("/exercises/{id}", s.handleUpdateExercise)
				priv.Delete("/exercises/{id}", s.handleArchiveExercise)
				priv.Post("/exercises/{id}/aliases", s.handleAddAlias)
				priv.Put("/exercises/{id}/image", s.handleSetExerciseImage)
				priv.Get("/exercises/{id}/image", s.handleGetExerciseImage)
				priv.Delete("/exercises/{id}/image", s.handleDeleteExerciseImage)
				priv.Get("/exercises/{id}/history", s.handleExerciseHistory)

				// Дневник тренировок (cookie или bearer).
				priv.Post("/workouts", s.handleCreateWorkout)
				priv.Get("/workouts", s.handleListWorkouts)
				priv.Get("/workouts/{id}", s.handleGetWorkout)
				priv.Patch("/workouts/{id}", s.handleUpdateWorkout)
				priv.Delete("/workouts/{id}", s.handleDeleteWorkout)
				priv.Post("/workouts/{id}/sets", s.handleAddSet)
				priv.Patch("/sets/{id}", s.handleUpdateSet)
				priv.Delete("/sets/{id}", s.handleDeleteSet)

				// Программы (cookie или bearer).
				priv.Get("/programs", s.handleListPrograms)
				priv.Post("/programs", s.handleCreateProgram)
				priv.Get("/programs/{id}", s.handleGetProgram)
				priv.Put("/programs/{id}", s.handleUpdateProgram)
				priv.Delete("/programs/{id}", s.handleDeleteProgram)
				priv.Post("/programs/{id}/archive", s.handleArchiveProgram)
				priv.Post("/programs/{id}/unarchive", s.handleUnarchiveProgram)
				priv.Get("/program-days/{id}", s.handleGetProgramDay)

				// Экспорт лога и настройки Telegram.
				priv.Get("/profile/telegram", s.handleGetTelegram)
				priv.Put("/profile/telegram", s.handleSetTelegram)
				priv.Post("/profile/telegram/link", s.handleLinkTelegram)
				priv.Post("/profile/telegram/test", s.handleTestTelegram)
				priv.Delete("/profile/telegram", s.handleDeleteTelegram)
				priv.Get("/profile/export", s.handleExportDownload)
				priv.Post("/profile/import", s.handleImport)

				// Персональные API-токены (только из cookie-сессии).
				priv.Get("/tokens", s.handleListTokens)
				priv.Post("/tokens", s.handleCreateToken)
				priv.Delete("/tokens/{id}", s.handleRevokeToken)

				priv.Group(func(own chi.Router) {
					own.Use(s.requireOwner)
					own.Get("/invites", s.handleListInvites)
					own.Post("/invites", s.handleCreateInvite)
					own.Delete("/invites/{id}", s.handleDeleteInvite)
				})
			})
		})
	})
	r.NotFound(spaHandler(opts.Static))

	return r
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := s.opts.DB.PingContext(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded"}`))
		return
	}
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// securityHeaders — базовые защитные заголовки на все ответы. CSP расширять
// осознанно по мере роста фронтенда.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

// jsonError отдаёт ошибку в едином конверте {"error":{code,message}}.
func jsonError(status int, code string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    code,
				"message": http.StatusText(status),
			},
		})
	}
}

// spaHandler отдаёт файлы фронтенда; для клиентских роутов (путей без
// расширения) — index.html, чтобы работал роутинг на стороне SPA.
// Отсутствующие ассеты (пути с расширением) получают честный 404.
// Каталоги файл-сервером не отдаются (авто-листинг раскрывает структуру).
func spaHandler(static fs.FS) http.HandlerFunc {
	fileServer := http.FileServerFS(static)
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" {
			p = "index.html"
		}
		if st, err := fs.Stat(static, p); err == nil && !st.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		if path.Ext(p) != "" {
			http.NotFound(w, r)
			return
		}
		http.ServeFileFS(w, r, static, "index.html")
	}
}
