package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/OlegKopeykin/fittrack/internal/server"
	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	static := fstest.MapFS{
		"index.html":    {Data: []byte(`<!doctype html><div id="root">fittrack-spa</div>`)},
		"assets/app.js": {Data: []byte(`console.log("app")`)},
	}
	srv := httptest.NewServer(server.New(server.Options{
		DB:     testutil.NewTestDB(t),
		Static: static,
	}))
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, string(body)
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)

	resp, body := get(t, srv.URL+"/healthz")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("body %q не является JSON: %v", body, err)
	}
	if payload.Status != "ok" {
		t.Errorf("status = %q, want %q", payload.Status, "ok")
	}
}

func TestAPIUnknownRouteReturnsJSONError(t *testing.T) {
	srv := newTestServer(t)

	resp, body := get(t, srv.URL+"/api/v1/nope")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("body %q не является JSON: %v", body, err)
	}
	if payload.Error.Code != "not_found" {
		t.Errorf("error.code = %q, want %q", payload.Error.Code, "not_found")
	}
}

func TestSPADirectoryRequestsFallBackToSPA(t *testing.T) {
	// Каталоги встроенной ФС не должны отдаваться файл-сервером:
	// автолистинг раскрывает структуру ассетов. Ожидание — SPA fallback.
	for _, path := range []string{"/assets/", "/assets"} {
		t.Run(path, func(t *testing.T) {
			srv := newTestServer(t)

			resp, body := get(t, srv.URL+path)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want 200 (SPA fallback)", path, resp.StatusCode)
			}
			if !strings.Contains(body, "fittrack-spa") {
				t.Errorf("GET %s: ожидался index.html (SPA fallback), body = %q", path, body)
			}
			if strings.Contains(body, "app.js") {
				t.Errorf("GET %s: в ответе листинг каталога вместо SPA", path)
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	srv := newTestServer(t)
	for _, path := range []string{"/", "/healthz", "/api/v1/nope", "/assets/app.js"} {
		t.Run(path, func(t *testing.T) {
			resp, _ := get(t, srv.URL+path)

			if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
			if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
				t.Errorf("X-Frame-Options = %q, want DENY", got)
			}
		})
	}

	resp, _ := get(t, srv.URL+"/")
	if csp := resp.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("Content-Security-Policy = %q, want содержит default-src 'self'", csp)
	}
}

func TestAPIUnknownRoutePOSTReturnsJSONError(t *testing.T) {
	// Пин: под /api нет роутов, любой метод на неизвестный путь — JSON-404
	// (а не 405): закрепляет удаление недостижимого MethodNotAllowed.
	srv := newTestServer(t)

	resp, err := http.Post(srv.URL+"/api/v1/nope", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("тело не является JSON: %v", err)
	}
	if payload.Error.Code != "not_found" {
		t.Errorf("error.code = %q, want not_found", payload.Error.Code)
	}
}

func TestSPAServing(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantContains string
	}{
		{"корень отдаёт index.html", "/", http.StatusOK, "fittrack-spa"},
		{"статический ассет отдаётся как есть", "/assets/app.js", http.StatusOK, `console.log("app")`},
		{"клиентский роут падает на index.html (SPA fallback)", "/workouts", http.StatusOK, "fittrack-spa"},
		{"отсутствующий ассет — 404, без fallback", "/assets/missing.js", http.StatusNotFound, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t)

			resp, body := get(t, srv.URL+tt.path)

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("GET %s: status = %d, want %d", tt.path, resp.StatusCode, tt.wantStatus)
			}
			if tt.wantContains != "" && !strings.Contains(body, tt.wantContains) {
				t.Errorf("GET %s: body %q не содержит %q", tt.path, body, tt.wantContains)
			}
		})
	}
}
