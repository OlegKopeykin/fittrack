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

func TestWorkoutProgramDayLink(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	dayID := aProgramDayID(t, ts)

	// старт «дня программы» связывает тренировку с днём
	resp := ts.PostJSON(t, "/api/v1/workouts", map[string]any{"program_day_id": dayID, "title": "Фул бади A"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", resp.StatusCode)
	}
	var wk struct {
		ProgramDayID *int64 `json:"program_day_id"`
	}
	testutil.DecodeJSON(t, resp, &wk)
	if wk.ProgramDayID == nil || *wk.ProgramDayID != dayID {
		t.Errorf("program_day_id = %v, want %d", wk.ProgramDayID, dayID)
	}

	// несуществующий день → 400
	bad := ts.PostJSON(t, "/api/v1/workouts", map[string]any{"program_day_id": int64(999999)})
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("несуществующий день: status = %d, want 400", bad.StatusCode)
	}

	// чужой день нельзя привязать и нельзя прочитать
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, err := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "sosed", "password": "надёжный-пароль"}))
	if err != nil {
		t.Fatal(err)
	}
	reg.Body.Close()

	cr, err := other.Post(ts.URL+"/api/v1/workouts", "application/json", mustJSON(map[string]any{"program_day_id": dayID}))
	if err != nil {
		t.Fatal(err)
	}
	cr.Body.Close()
	if cr.StatusCode != http.StatusBadRequest {
		t.Errorf("чужой день при создании: status = %d, want 400", cr.StatusCode)
	}
	rd, err := other.Get(ts.URL + "/api/v1/program-days/" + itoa(dayID))
	if err != nil {
		t.Fatal(err)
	}
	rd.Body.Close()
	if rd.StatusCode != http.StatusNotFound {
		t.Errorf("чужой день при чтении: status = %d, want 404", rd.StatusCode)
	}
}

func TestWorkoutTitleRoundtrip(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.PostJSON(t, "/api/v1/workouts", map[string]any{"title": "Full-A"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", resp.StatusCode)
	}
	var wk struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	testutil.DecodeJSON(t, resp, &wk)
	if wk.Title != "Full-A" {
		t.Errorf("title после создания = %q, want Full-A", wk.Title)
	}

	upd := patchJSON(t, ts, "/api/v1/workouts/"+itoa(wk.ID), map[string]any{"title": "Full-B"})
	if upd.StatusCode != http.StatusOK {
		t.Fatalf("update: status = %d, want 200", upd.StatusCode)
	}
	var got struct {
		Title string `json:"title"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wk.ID)), &got)
	if got.Title != "Full-B" {
		t.Errorf("title после обновления = %q, want Full-B", got.Title)
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

func TestUpdateSetPartialPreservesFields(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")

	resp := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 80.0, "reps": 5, "note": "легко",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add set: status = %d, want 201", resp.StatusCode)
	}
	var st struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &st)

	// Частичный PATCH: меняем только role. Вес, повторы и заметку не шлём —
	// они не должны обнулиться.
	upd := patchJSON(t, ts, "/api/v1/sets/"+itoa(st.ID), map[string]any{"role": "warmup"})
	if upd.StatusCode != http.StatusOK {
		t.Fatalf("patch set: status = %d, want 200", upd.StatusCode)
	}
	var got struct {
		Role     string   `json:"role"`
		WeightKg *float64 `json:"weight_kg"`
		Reps     *int     `json:"reps"`
		Note     string   `json:"note"`
	}
	testutil.DecodeJSON(t, upd, &got)
	if got.Role != "warmup" {
		t.Errorf("role = %q, want warmup", got.Role)
	}
	if got.WeightKg == nil || *got.WeightKg != 80.0 {
		t.Errorf("weight_kg = %v, want 80 (частичный PATCH не должен затирать)", got.WeightKg)
	}
	if got.Reps == nil || *got.Reps != 5 {
		t.Errorf("reps = %v, want 5 (частичный PATCH не должен затирать)", got.Reps)
	}
	if got.Note != "легко" {
		t.Errorf("note = %q, want «легко» (частичный PATCH не должен затирать)", got.Note)
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

// --- добор покрытия: удаление, валидация, bearer add-set, пагинация ---

func deleteReq(t *testing.T, ts *testutil.TestServer, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestDeleteSet(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	resp := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 60, "reps": 8,
	})
	var st struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &st)

	del := deleteReq(t, ts, "/api/v1/sets/"+itoa(st.ID))
	if del.StatusCode != http.StatusNoContent {
		t.Fatalf("delete set: status = %d, want 204", del.StatusCode)
	}
	var full struct {
		Sets []map[string]any `json:"sets"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wid)), &full)
	if len(full.Sets) != 0 {
		t.Errorf("после удаления подходов = %d, want 0", len(full.Sets))
	}
}

func TestDeleteWorkout(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)

	del := deleteReq(t, ts, "/api/v1/workouts/"+itoa(wid))
	if del.StatusCode != http.StatusNoContent {
		t.Fatalf("delete workout: status = %d, want 204", del.StatusCode)
	}
	if got := ts.Get(t, "/api/v1/workouts/"+itoa(wid)); got.StatusCode != http.StatusNotFound {
		t.Errorf("после удаления GET: status = %d, want 404", got.StatusCode)
	}
}

func TestSetMutationOwnership(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	resp := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 60, "reps": 8,
	})
	var st struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &st)

	// второй пользователь в отдельной сессии не видит чужой подход
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, err := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "chuzhoy", "password": "надёжный-пароль"}))
	if err != nil {
		t.Fatal(err)
	}
	reg.Body.Close()

	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/sets/"+itoa(st.ID), mustJSON(map[string]any{"role": "warmup"}))
	req.Header.Set("Content-Type", "application/json")
	patch, err := other.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	patch.Body.Close()
	if patch.StatusCode != http.StatusNotFound {
		t.Errorf("чужой PATCH подхода: status = %d, want 404", patch.StatusCode)
	}

	dreq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/sets/"+itoa(st.ID), nil)
	dresp, err := other.Do(dreq)
	if err != nil {
		t.Fatal(err)
	}
	dresp.Body.Close()
	if dresp.StatusCode != http.StatusNotFound {
		t.Errorf("чужой DELETE подхода: status = %d, want 404", dresp.StatusCode)
	}
}

func TestAddSetValidation(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")

	// нет exercise_id → 400
	noEx := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{"role": "working", "reps": 8})
	if noEx.StatusCode != http.StatusBadRequest {
		t.Errorf("без exercise_id: status = %d, want 400", noEx.StatusCode)
	}
	// неверный role → 400
	badRole := ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{"exercise_id": ex, "role": "bogus", "reps": 8})
	if badRole.StatusCode != http.StatusBadRequest {
		t.Errorf("неверный role: status = %d, want 400", badRole.StatusCode)
	}
}

func TestBearerAddsSet(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	token := issueToken(t, ts, "sync")

	resp := bearer(t, ts, http.MethodPost, "/api/v1/workouts", token, map[string]any{"date": "2026-07-01"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("bearer create workout: status = %d, want 201", resp.StatusCode)
	}
	var wk struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &wk)

	add := bearer(t, ts, http.MethodPost, "/api/v1/workouts/"+itoa(wk.ID)+"/sets", token,
		map[string]any{"exercise_id": ex, "role": "working", "weight_kg": 70, "reps": 6})
	if add.StatusCode != http.StatusCreated {
		t.Fatalf("bearer add set: status = %d, want 201", add.StatusCode)
	}
	var full struct {
		Sets []map[string]any `json:"sets"`
	}
	testutil.DecodeJSON(t, bearer(t, ts, http.MethodGet, "/api/v1/workouts/"+itoa(wk.ID), token, nil), &full)
	if len(full.Sets) != 1 {
		t.Errorf("подходов по bearer = %d, want 1", len(full.Sets))
	}
}

func TestWorkoutListPagination(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	dates := []string{"2026-06-01", "2026-06-02", "2026-06-03"}
	for _, d := range dates {
		resp := ts.PostJSON(t, "/api/v1/workouts", map[string]any{"date": d})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create %s: status = %d", d, resp.StatusCode)
		}
	}

	type page struct {
		Items []struct {
			Date string `json:"date"`
		} `json:"items"`
		NextCursor string `json:"next_cursor"`
	}
	var p1 page
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts?limit=2"), &p1)
	if len(p1.Items) != 2 {
		t.Fatalf("страница 1: items = %d, want 2", len(p1.Items))
	}
	if p1.NextCursor == "" {
		t.Fatal("страница 1: пустой next_cursor при остатке")
	}

	var p2 page
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts?limit=2&cursor="+p1.NextCursor), &p2)
	if len(p2.Items) != 1 {
		t.Fatalf("страница 2: items = %d, want 1", len(p2.Items))
	}
	if p2.NextCursor != "" {
		t.Errorf("страница 2: next_cursor = %q, want пусто (данные кончились)", p2.NextCursor)
	}
	// нет пересечения дат между страницами
	seen := map[string]bool{}
	for _, it := range p1.Items {
		seen[it.Date] = true
	}
	for _, it := range p2.Items {
		if seen[it.Date] {
			t.Errorf("дата %s встречается на обеих страницах", it.Date)
		}
	}

	// битый курсор → 400
	if bad := ts.Get(t, "/api/v1/workouts?cursor=not-a-cursor"); bad.StatusCode != http.StatusBadRequest {
		t.Errorf("битый курсор: status = %d, want 400", bad.StatusCode)
	}
}
