package server_test

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func TestRegisterFlow(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	code := ts.CreateInvite(t, "user", "")

	resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
		"invite_code": code, "username": "oleg", "password": "надёжный-пароль",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: status = %d, want 201", resp.StatusCode)
	}

	me := ts.Get(t, "/api/v1/auth/me")
	if me.StatusCode != http.StatusOK {
		t.Fatalf("me после регистрации: status = %d, want 200", me.StatusCode)
	}
	var user struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	testutil.DecodeJSON(t, me, &user)
	if user.Username != "oleg" || user.Role != "user" {
		t.Errorf("me = %+v, want username=oleg role=user", user)
	}
}

func TestRegisterValidation(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	code := ts.CreateInvite(t, "user", "")

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"короткий username", "ab", "надёжный-пароль"},
		{"пробел в username", "ol eg", "надёжный-пароль"},
		{"короткий пароль", "oleg", "1234567"},
		{"пустой пароль", "oleg", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
				"invite_code": code, "username": tt.username, "password": tt.password,
			})
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			if got := testutil.ErrorCode(t, resp); got != "invalid_input" {
				t.Errorf("error.code = %q, want invalid_input", got)
			}
		})
	}
}

func TestRegisterInviteRejections(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)

	used := ts.CreateInvite(t, "user", "")
	ts.Register(t, used, "first", "надёжный-пароль")

	expired := ts.CreateInvite(t, "user", "2020-01-01T00:00:00Z")

	tests := []struct {
		name string
		code string
	}{
		{"несуществующий код", "NOSUCHCODE22"},
		{"уже использованный код", used},
		{"просроченный код", expired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
				"invite_code": tt.code, "username": "second", "password": "надёжный-пароль",
			})
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", resp.StatusCode)
			}
			if got := testutil.ErrorCode(t, resp); got != "invalid_invite" {
				t.Errorf("error.code = %q, want invalid_invite", got)
			}
		})
	}
}

func TestRegisterDuplicateUsernameCaseInsensitive(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "Oleg", "надёжный-пароль")

	resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
		"invite_code": ts.CreateInvite(t, "user", ""), "username": "oleg", "password": "надёжный-пароль",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
	if got := testutil.ErrorCode(t, resp); got != "conflict" {
		t.Errorf("error.code = %q, want conflict", got)
	}
}

func TestRegisterInviteRace(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	code := ts.CreateInvite(t, "user", "")

	const n = 4
	statuses := make([]int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp := ts.PostJSON(t, "/api/v1/auth/register", map[string]string{
				"invite_code": code,
				"username":    "racer" + string(rune('a'+i)),
				"password":    "надёжный-пароль",
			})
			statuses[i] = resp.StatusCode
		}(i)
	}
	wg.Wait()

	created := 0
	for _, st := range statuses {
		if st == http.StatusCreated {
			created++
		}
	}
	if created != 1 {
		t.Errorf("инвайт сработал %d раз, want ровно 1 (statuses=%v)", created, statuses)
	}
}

func TestLoginLogoutFlow(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "oleg", "надёжный-пароль")

	logout := ts.PostJSON(t, "/api/v1/auth/logout", nil)
	if logout.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: status = %d, want 204", logout.StatusCode)
	}
	if me := ts.Get(t, "/api/v1/auth/me"); me.StatusCode != http.StatusUnauthorized {
		t.Fatalf("me после logout: status = %d, want 401", me.StatusCode)
	}

	login := ts.PostJSON(t, "/api/v1/auth/login", map[string]string{
		"username": "oleg", "password": "надёжный-пароль",
	})
	if login.StatusCode != http.StatusOK {
		t.Fatalf("login: status = %d, want 200", login.StatusCode)
	}
	if me := ts.Get(t, "/api/v1/auth/me"); me.StatusCode != http.StatusOK {
		t.Fatalf("me после login: status = %d, want 200", me.StatusCode)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "oleg", "надёжный-пароль")
	ts.PostJSON(t, "/api/v1/auth/logout", nil)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"неверный пароль", "oleg", "не тот пароль"},
		{"несуществующий пользователь", "ghost", "надёжный-пароль"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.PostJSON(t, "/api/v1/auth/login", map[string]string{
				"username": tt.username, "password": tt.password,
			})
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", resp.StatusCode)
			}
			if got := testutil.ErrorCode(t, resp); got != "invalid_credentials" {
				t.Errorf("error.code = %q, want invalid_credentials", got)
			}
		})
	}
}

func TestLoginRateLimit(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	ts := testutil.NewTestServer(t, clock)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "oleg", "надёжный-пароль")
	ts.PostJSON(t, "/api/v1/auth/logout", nil)

	for i := 0; i < 10; i++ {
		resp := ts.PostJSON(t, "/api/v1/auth/login", map[string]string{
			"username": "oleg", "password": "перебор",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("попытка %d: status = %d, want 401", i+1, resp.StatusCode)
		}
	}

	blocked := ts.PostJSON(t, "/api/v1/auth/login", map[string]string{
		"username": "oleg", "password": "надёжный-пароль",
	})
	if blocked.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("после 10 неудач: status = %d, want 429", blocked.StatusCode)
	}
	if got := testutil.ErrorCode(t, blocked); got != "rate_limited" {
		t.Errorf("error.code = %q, want rate_limited", got)
	}

	now = now.Add(16 * time.Minute)
	ok := ts.PostJSON(t, "/api/v1/auth/login", map[string]string{
		"username": "oleg", "password": "надёжный-пароль",
	})
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("после окна: status = %d, want 200", ok.StatusCode)
	}
}

func TestCSRFOriginGuard(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "oleg", "надёжный-пароль")
	ts.PostJSON(t, "/api/v1/auth/logout", nil)

	makeLogin := func(origin, contentType string) *http.Response {
		req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth/login",
			strings.NewReader(`{"username":"oleg","password":"надёжный-пароль"}`))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Content-Type", contentType)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := ts.Client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}

	if resp := makeLogin("https://evil.example", "application/json"); resp.StatusCode != http.StatusForbidden {
		t.Errorf("чужой Origin: status = %d, want 403", resp.StatusCode)
	}
	if resp := makeLogin(ts.URL, "application/json"); resp.StatusCode != http.StatusOK {
		t.Errorf("свой Origin: status = %d, want 200", resp.StatusCode)
	}
	if resp := makeLogin("", "text/plain"); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("не-JSON Content-Type: status = %d, want 415", resp.StatusCode)
	}
}

func TestMeRequiresAuth(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	resp := ts.Get(t, "/api/v1/auth/me")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	if got := testutil.ErrorCode(t, resp); got != "unauthorized" {
		t.Errorf("error.code = %q, want unauthorized", got)
	}
}

func TestInvitesRequireOwner(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "user", ""), "user1", "надёжный-пароль")

	resp := ts.Get(t, "/api/v1/invites")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("обычный пользователь: status = %d, want 403", resp.StatusCode)
	}
}

func TestInvitesOwnerCRUD(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.Register(t, ts.CreateInvite(t, "owner", ""), "boss", "надёжный-пароль")

	create := ts.PostJSON(t, "/api/v1/invites", map[string]any{"expires_in_days": 14})
	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", create.StatusCode)
	}
	var invite struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	testutil.DecodeJSON(t, create, &invite)
	if !regexp.MustCompile(`^[A-Z2-7]{12}$`).MatchString(invite.Code) {
		t.Errorf("code = %q, want 12 символов base32", invite.Code)
	}

	list := ts.Get(t, "/api/v1/invites")
	if list.StatusCode != http.StatusOK {
		t.Fatalf("list: status = %d, want 200", list.StatusCode)
	}
	var invites []struct {
		Code string `json:"code"`
	}
	testutil.DecodeJSON(t, list, &invites)
	if len(invites) < 2 { // бутстрап-инвайт владельца + созданный
		t.Errorf("list: %d инвайтов, want ≥ 2", len(invites))
	}

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/invites/"+itoa(invite.ID), nil)
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want 204", resp.StatusCode)
	}

	req2, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/invites/"+itoa(invite.ID), nil)
	resp2, err := ts.Client.Do(req2)
	if err != nil {
		t.Fatalf("DELETE #2: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("повторный delete: status = %d, want 404", resp2.StatusCode)
	}
}

func TestHealthzDegradedOnClosedDB(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ts.DB.Close()

	resp := ts.Get(t, "/healthz")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
