// Package testutil — реальная временная SQLite и полный HTTP-сервер для тестов.
package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/auth"
	"github.com/OlegKopeykin/fittrack/internal/db"
	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/seed"
	"github.com/OlegKopeykin/fittrack/internal/server"
)

// NullInt оборачивает id в sql.NullInt64 (Valid=true) для параметров sqlc.
func NullInt(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

// NewTestDB открывает мигрированную SQLite во временном файле
// (как в проде: файл + WAL, не :memory:).
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("testutil: db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// TestServer — поднятый httptest-сервер с реальной БД и клиентом с cookie-jar.
type TestServer struct {
	*httptest.Server
	DB     *sql.DB
	Q      *gen.Queries
	Client *http.Client
	Now    func() time.Time
}

// NewTestServer собирает полный роутер поверх свежей БД.
// clock == nil — реальное время.
func NewTestServer(t *testing.T, clock func() time.Time) *TestServer {
	t.Helper()
	conn := NewTestDB(t)
	static := fstest.MapFS{
		"index.html": {Data: []byte(`<!doctype html><div id="root">fittrack-spa</div>`)},
	}
	if err := seed.LoadCatalog(t.Context(), conn); err != nil {
		t.Fatalf("testutil: seed: %v", err)
	}
	h := server.New(server.Options{DB: conn, Static: static, Now: clock})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	client := &http.Client{Jar: jar}

	return &TestServer{Server: srv, DB: conn, Q: gen.New(conn), Client: client, Now: clock}
}

// CreateInvite вставляет инвайт напрямую в БД и возвращает код.
func (ts *TestServer) CreateInvite(t *testing.T, role string, expiresAt string) string {
	t.Helper()
	code, err := auth.NewInviteCode()
	if err != nil {
		t.Fatalf("NewInviteCode: %v", err)
	}
	var expires sql.NullString
	if expiresAt != "" {
		expires = sql.NullString{String: expiresAt, Valid: true}
	}
	_, err = ts.Q.CreateInvite(t.Context(), gen.CreateInviteParams{
		Code:      code,
		Role:      role,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		ExpiresAt: expires,
	})
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	return code
}

// PostJSON шлёт POST с JSON-телом через клиент с cookie-jar.
func (ts *TestServer) PostJSON(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := ts.Client.Post(ts.URL+path, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// Get шлёт GET через клиент с cookie-jar.
func (ts *TestServer) Get(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := ts.Client.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// Register регистрирует пользователя по инвайту и оставляет сессию в jar.
func (ts *TestServer) Register(t *testing.T, code, username, password string) {
	t.Helper()
	resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
		"invite_code": code, "username": username, "password": password,
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("register %s: status %d, body %s", username, resp.StatusCode, body)
	}
}

// DecodeJSON декодирует тело ответа в v.
func DecodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// ErrorCode достаёт error.code из конверта ошибки.
func ErrorCode(t *testing.T, resp *http.Response) string {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	DecodeJSON(t, resp, &payload)
	return payload.Error.Code
}
