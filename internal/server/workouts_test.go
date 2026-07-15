package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func mustJSON(v any) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func patchJSON(t *testing.T, ts *testutil.TestServer, path string, body any) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, ts.URL+path, mustJSON(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func finish(t *testing.T, ts *testutil.TestServer, id int64, at string) {
	t.Helper()
	resp := patchJSON(t, ts, "/api/v1/workouts/"+itoa(id), map[string]any{"finished_at": at})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("finish workout: status = %d", resp.StatusCode)
	}
}

// anExerciseID возвращает id глобального упражнения по части имени.
func anExerciseID(t *testing.T, ts *testutil.TestServer, q string) int64 {
	t.Helper()
	var list []exDTO
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises?q="+q), &list)
	if len(list) == 0 {
		t.Fatalf("нет упражнения по запросу %q", q)
	}
	return list[0].ID
}

func newWorkout(t *testing.T, ts *testutil.TestServer) int64 {
	t.Helper()
	resp := ts.PostJSON(t, "/api/v1/workouts", map[string]any{})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create workout: status = %d, want 201", resp.StatusCode)
	}
	var wk struct {
		ID   int64  `json:"id"`
		Date string `json:"date"`
	}
	testutil.DecodeJSON(t, resp, &wk)
	if wk.Date == "" {
		t.Error("у тренировки нет даты по умолчанию")
	}
	return wk.ID
}

func TestWorkoutCreateGetList(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	id := newWorkout(t, ts)

	resp := ts.Get(t, "/api/v1/workouts/"+itoa(id))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: status = %d, want 200", resp.StatusCode)
	}

	var page struct {
		Items      []map[string]any `json:"items"`
		NextCursor string           `json:"next_cursor"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts"), &page)
	if len(page.Items) != 1 {
		t.Errorf("список тренировок = %d, want 1", len(page.Items))
	}
}

func TestAddSetAndReadBack(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")

	resp := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 60.5, "reps": 8,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add set: status = %d, want 201", resp.StatusCode)
	}
	var st struct {
		WeightKg float64 `json:"weight_kg"`
		Reps     int     `json:"reps"`
	}
	testutil.DecodeJSON(t, resp, &st)
	if st.WeightKg != 60.5 || st.Reps != 8 {
		t.Errorf("подход = %.2f×%d, want 60.50×8 (проверка кг↔граммы)", st.WeightKg, st.Reps)
	}

	var full struct {
		Sets []map[string]any `json:"sets"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wid)), &full)
	if len(full.Sets) != 1 {
		t.Errorf("подходов в тренировке = %d, want 1", len(full.Sets))
	}
}

func TestSetIdempotentByClientID(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	body := map[string]any{"exercise_id": ex, "weight_kg": 60, "reps": 8, "client_id": "uuid-1"}

	first := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", body)
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first: status = %d, want 201", first.StatusCode)
	}
	var s1 struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, first, &s1)

	second := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", body)
	if second.StatusCode != http.StatusOK {
		t.Fatalf("повтор: status = %d, want 200", second.StatusCode)
	}
	var s2 struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, second, &s2)
	if s1.ID != s2.ID {
		t.Errorf("повтор создал новый подход (%d != %d)", s1.ID, s2.ID)
	}

	var full struct {
		Sets []map[string]any `json:"sets"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wid)), &full)
	if len(full.Sets) != 1 {
		t.Errorf("после идемпотентного повтора подходов = %d, want 1", len(full.Sets))
	}
}

func TestWorkoutOwnership(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)

	// второй пользователь в отдельной сессии
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, err := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "someoneelse", "password": "надёжный-пароль"}))
	if err != nil {
		t.Fatal(err)
	}
	reg.Body.Close()

	resp, err := other.Get(ts.URL + "/api/v1/workouts/" + itoa(wid))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("чужая тренировка: status = %d, want 404", resp.StatusCode)
	}
}

func TestExerciseHistory(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	ex := anExerciseID(t, ts, "Присед")

	// первая (завершённая) тренировка: 60×8, 60×7
	w1 := newWorkout(t, ts)
	for _, reps := range []int{8, 7} {
		ts.PostJSON(t, "/api/v1/workouts/"+itoa(w1)+"/sets", map[string]any{"exercise_id": ex, "weight_kg": 60, "reps": reps})
	}
	finish(t, ts, w1, "2026-07-10T12:00:00Z")

	// незавершённая тренировка не должна попадать в историю
	w2 := newWorkout(t, ts)
	ts.PostJSON(t, "/api/v1/workouts/"+itoa(w2)+"/sets", map[string]any{"exercise_id": ex, "weight_kg": 62, "reps": 6})

	var hist []struct {
		Date string `json:"date"`
		Sets []struct {
			WeightKg float64 `json:"weight_kg"`
			Reps     int     `json:"reps"`
		} `json:"sets"`
	}
	resp := ts.Get(t, "/api/v1/exercises/"+itoa(ex)+"/history")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history: status = %d, want 200", resp.StatusCode)
	}
	testutil.DecodeJSON(t, resp, &hist)
	if len(hist) != 1 {
		t.Fatalf("сессий в истории = %d, want 1 (только завершённые)", len(hist))
	}
	if len(hist[0].Sets) != 2 || hist[0].Sets[0].Reps != 8 {
		t.Errorf("подходы истории = %+v, want [60×8, 60×7]", hist[0].Sets)
	}
}

func TestWorkoutBodyweightRoundtrip(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)

	req := patchJSON(t, ts, "/api/v1/workouts/"+itoa(wid), map[string]any{"bodyweight_kg": 86.5, "feeling": "бодро"})
	if req.StatusCode != http.StatusOK {
		t.Fatalf("patch: status = %d, want 200", req.StatusCode)
	}
	var wk struct {
		BodyweightKg float64 `json:"bodyweight_kg"`
		Feeling      string  `json:"feeling"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wid)), &wk)
	if wk.BodyweightKg != 86.5 || wk.Feeling != "бодро" {
		t.Errorf("вес/самочувствие = %.2f/%q, want 86.50/бодро", wk.BodyweightKg, wk.Feeling)
	}
}

func TestBearerCreatesWorkout(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "sync")

	resp := bearer(t, ts, http.MethodPost, "/api/v1/workouts", token, map[string]any{"date": "2026-07-01"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("bearer create workout: status = %d, want 201", resp.StatusCode)
	}
}
