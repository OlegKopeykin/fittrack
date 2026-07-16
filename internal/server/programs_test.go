package server_test

import (
	"net/http"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func TestCreateAndGetProgram(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	body := map[string]any{
		"name":        "Full Body A",
		"description": "тяжёлая",
		"days": []map[string]any{{
			"name": "День A",
			"exercises": []map[string]any{
				{"exercise_name": "Присед в Смите", "sets": 3, "rep_min": 6, "rep_max": 10, "weight_min_kg": 70, "weight_max_kg": 90, "tempo": "3-0-1"},
				{"exercise_name": "Жим гантелей лёжа", "sets": 3, "rep_min": 6, "rep_max": 10},
			},
		}},
	}
	resp := ts.PostJSON(t, "/api/v1/programs", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", resp.StatusCode)
	}
	var prog struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Days []struct {
			Name      string `json:"name"`
			Exercises []struct {
				ExerciseID  int64   `json:"exercise_id"`
				Sets        int     `json:"sets"`
				RepMin      int     `json:"rep_min"`
				WeightMinKg float64 `json:"weight_min_kg"`
				Tempo       string  `json:"tempo"`
			} `json:"exercises"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, resp, &prog)
	if len(prog.Days) != 1 || len(prog.Days[0].Exercises) != 2 {
		t.Fatalf("структура программы = %+v", prog)
	}
	first := prog.Days[0].Exercises[0]
	if first.Sets != 3 || first.RepMin != 6 || first.WeightMinKg != 70 || first.Tempo != "3-0-1" {
		t.Errorf("предписание = %+v, want sets3 rep6 вес70 темп3-0-1", first)
	}

	// повторное чтение отдаёт то же
	var got struct {
		Days []struct {
			Exercises []map[string]any `json:"exercises"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)), &got)
	if len(got.Days) != 1 || len(got.Days[0].Exercises) != 2 {
		t.Errorf("GET программы вернул иную структуру: %+v", got)
	}
}

// aProgramDayID создаёт программу с одним днём (2 упражнения) и возвращает id дня.
func aProgramDayID(t *testing.T, ts *testutil.TestServer) int64 {
	t.Helper()
	body := map[string]any{
		"name": "Фул бади",
		"days": []map[string]any{{
			"name": "День A",
			"exercises": []map[string]any{
				{"exercise_name": "Присед в Смите", "sets": 3, "rep_min": 6, "rep_max": 10, "weight_min_kg": 70, "weight_max_kg": 90},
				{"exercise_name": "Жим гантелей лёжа", "sets": 3, "rep_min": 8, "rep_max": 12},
			},
		}},
	}
	resp := ts.PostJSON(t, "/api/v1/programs", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create program: status = %d, want 201", resp.StatusCode)
	}
	var prog struct {
		ID   int64 `json:"id"`
		Days []struct {
			ID int64 `json:"id"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, resp, &prog)
	if len(prog.Days) != 1 {
		t.Fatalf("ожидался 1 день, получено %d", len(prog.Days))
	}
	return prog.Days[0].ID
}

func TestGetProgramDay(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	dayID := aProgramDayID(t, ts)

	var day struct {
		ID          int64  `json:"id"`
		ProgramName string `json:"program_name"`
		Name        string `json:"name"`
		Exercises   []struct {
			ExerciseID int64 `json:"exercise_id"`
			Sets       int64 `json:"sets"`
		} `json:"exercises"`
	}
	resp := ts.Get(t, "/api/v1/program-days/"+itoa(dayID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get day: status = %d, want 200", resp.StatusCode)
	}
	testutil.DecodeJSON(t, resp, &day)
	if day.ProgramName != "Фул бади" || day.Name != "День A" || len(day.Exercises) != 2 {
		t.Errorf("день = %+v, want «Фул бади»/«День A»/2 упражнения", day)
	}

	if miss := ts.Get(t, "/api/v1/program-days/999999"); miss.StatusCode != http.StatusNotFound {
		t.Errorf("несуществующий день: status = %d, want 404", miss.StatusCode)
	}
}

func TestCreateProgramUnknownExercise(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	resp := ts.PostJSON(t, "/api/v1/programs", map[string]any{
		"name": "P", "days": []map[string]any{{"name": "d", "exercises": []map[string]any{
			{"exercise_name": "Несуществующее упражнение", "sets": 3},
		}}},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestListPrograms(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	for _, n := range []string{"A", "B"} {
		ts.PostJSON(t, "/api/v1/programs", map[string]any{"name": n})
	}
	var list []map[string]any
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs"), &list)
	if len(list) != 2 {
		t.Errorf("программ = %d, want 2", len(list))
	}
}

func TestProgramNameRequired(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	resp := ts.PostJSON(t, "/api/v1/programs", map[string]any{"name": ""})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestDeleteProgram(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	var prog struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, ts.PostJSON(t, "/api/v1/programs", map[string]any{"name": "X"}), &prog)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/programs/"+itoa(prog.ID), nil)
	resp, _ := ts.Client.Do(req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: status = %d, want 204", resp.StatusCode)
	}
	if after := ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)); after.StatusCode != http.StatusNotFound {
		t.Errorf("после удаления GET: status = %d, want 404", after.StatusCode)
	}
}

func TestBearerCreatesProgram(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	token := issueToken(t, ts, "sync")
	resp := bearer(t, ts, http.MethodPost, "/api/v1/programs", token, map[string]any{"name": "по API"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("bearer create program: status = %d, want 201", resp.StatusCode)
	}
}

func TestProgramArchiveFlow(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	var prog struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, ts.PostJSON(t, "/api/v1/programs", map[string]any{"name": "Старая"}), &prog)

	// в архив
	if resp := ts.PostJSON(t, "/api/v1/programs/"+itoa(prog.ID)+"/archive", nil); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("archive: status = %d, want 204", resp.StatusCode)
	}
	// пропала из активного списка
	var active []map[string]any
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs"), &active)
	if len(active) != 0 {
		t.Errorf("активных программ = %d, want 0", len(active))
	}
	// видна в архиве
	var archived []struct {
		Archived bool `json:"archived"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs?archived=1"), &archived)
	if len(archived) != 1 || !archived[0].Archived {
		t.Errorf("архивных программ = %+v, want 1 archived", archived)
	}
	// вернуть из архива
	if resp := ts.PostJSON(t, "/api/v1/programs/"+itoa(prog.ID)+"/unarchive", nil); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("unarchive: status = %d, want 204", resp.StatusCode)
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs"), &active)
	if len(active) != 1 {
		t.Errorf("после возврата активных = %d, want 1", len(active))
	}
}
