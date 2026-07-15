package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

// issueToken выпускает персональный токен через cookie-сессию и возвращает
// его открытое значение.
func issueToken(t *testing.T, ts *testutil.TestServer, name string) string {
	t.Helper()
	resp := ts.PostJSON(t, "/api/v1/tokens", map[string]any{"name": name})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create token: status = %d, want 201", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if out.Token == "" {
		t.Fatal("токен пуст")
	}
	return out.Token
}

// bearer выполняет запрос с заголовком Authorization: Bearer <token>
// через отдельный клиент БЕЗ cookie (чистый машинный доступ).
func bearer(t *testing.T, ts *testutil.TestServer, method, path, token string, body any) *http.Response {
	t.Helper()
	var r *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, ts.URL+path, r)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{}).Do(req) // без cookie-jar
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestTokenLifecycleAndList(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "obsidian-sync")

	var list []struct {
		Name   string `json:"name"`
		Prefix string `json:"prefix"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/tokens"), &list)
	if len(list) != 1 || list[0].Name != "obsidian-sync" {
		t.Fatalf("список токенов = %+v", list)
	}
	// список не должен содержать открытое значение токена
	if list[0].Prefix == token {
		t.Error("в списке лежит полный токен, а должен только префикс")
	}
}

func TestBearerAccessesData(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "machine")

	resp := bearer(t, ts, http.MethodGet, "/api/v1/exercises", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bearer GET exercises: status = %d, want 200", resp.StatusCode)
	}
}

func TestBearerBulkUploadIdempotent(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "importer")

	items := []map[string]any{
		{"name": "Апи-упражнение 1", "muscle_group": "chest", "kind": "compound"},
		{"name": "Апи-упражнение 2", "muscle_group": "quads", "kind": "compound"},
	}

	first := bearer(t, ts, http.MethodPost, "/api/v1/exercises/bulk", token, items)
	if first.StatusCode != http.StatusOK {
		t.Fatalf("bulk #1: status = %d, want 200", first.StatusCode)
	}
	var r1 struct{ Created, Skipped int }
	testutil.DecodeJSON(t, first, &r1)
	if r1.Created != 2 || r1.Skipped != 0 {
		t.Errorf("bulk #1 = %+v, want created=2 skipped=0", r1)
	}

	second := bearer(t, ts, http.MethodPost, "/api/v1/exercises/bulk", token, items)
	var r2 struct{ Created, Skipped int }
	testutil.DecodeJSON(t, second, &r2)
	if r2.Created != 0 || r2.Skipped != 2 {
		t.Errorf("bulk #2 (повтор) = %+v, want created=0 skipped=2", r2)
	}
}

func TestRevokedTokenRejected(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.PostJSON(t, "/api/v1/tokens", map[string]any{"name": "temp"})
	var out struct {
		Token string `json:"token"`
		Info  struct {
			ID int64 `json:"id"`
		} `json:"info"`
	}
	testutil.DecodeJSON(t, resp, &out)

	// отзыв через cookie-сессию
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/tokens/"+itoa(out.Info.ID), nil)
	rr, _ := ts.Client.Do(req)
	rr.Body.Close()
	if rr.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke: status = %d, want 204", rr.StatusCode)
	}

	after := bearer(t, ts, http.MethodGet, "/api/v1/exercises", out.Token, nil)
	if after.StatusCode != http.StatusUnauthorized {
		t.Errorf("отозванный токен: status = %d, want 401", after.StatusCode)
	}
}

func TestInvalidBearerRejected(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	resp := bearer(t, ts, http.MethodGet, "/api/v1/exercises", "fit_totallyfake", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("несуществующий токен: status = %d, want 401", resp.StatusCode)
	}
}

func TestBearerCannotIssueToken(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "m2m")

	resp := bearer(t, ts, http.MethodPost, "/api/v1/tokens", token, map[string]any{"name": "child"})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("выпуск токена по bearer: status = %d, want 403", resp.StatusCode)
	}
}
