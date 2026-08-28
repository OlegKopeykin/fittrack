package server_test

import (
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func putJSON(t *testing.T, ts *testutil.TestServer, path string, body any) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, ts.URL+path, mustJSON(body))
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

func TestUpdateProgramReplacesContent(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	// создаём программу с одним днём
	resp := ts.PostJSON(t, "/api/v1/programs", map[string]any{
		"name": "Черновик",
		"days": []map[string]any{{"name": "День 1", "exercises": []map[string]any{
			{"exercise_name": "Присед в Смите"},
		}}},
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d", resp.StatusCode)
	}
	var prog struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &prog)

	// PUT: новое имя и два дня
	upd := putJSON(t, ts, "/api/v1/programs/"+itoa(prog.ID), map[string]any{
		"name": "Фул бади",
		"days": []map[string]any{
			{"name": "День A", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
			{"name": "День B", "exercises": []map[string]any{{"exercise_name": "Жим гантелей лёжа"}}},
		},
	})
	if upd.StatusCode != http.StatusOK {
		t.Fatalf("update: status = %d, want 200", upd.StatusCode)
	}

	var got struct {
		Name string `json:"name"`
		Days []struct {
			Name      string           `json:"name"`
			Exercises []map[string]any `json:"exercises"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)), &got)
	if got.Name != "Фул бади" || len(got.Days) != 2 {
		t.Fatalf("после PUT = %+v, want «Фул бади»/2 дня", got)
	}
	if got.Days[1].Name != "День B" || len(got.Days[1].Exercises) != 1 {
		t.Errorf("второй день = %+v", got.Days[1])
	}
}

// programDaysWithFields описывает ответ GET программы с полным набором полей,
// нужных, чтобы проверить, что правка ничего не теряет.
type programWithFields struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Days        []struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Notes     string `json:"notes"`
		Exercises []struct {
			ExerciseID  int64   `json:"exercise_id"`
			Sets        int64   `json:"sets"`
			RepMin      int     `json:"rep_min"`
			RepMax      int     `json:"rep_max"`
			WeightMinKg float64 `json:"weight_min_kg"`
			WeightMaxKg float64 `json:"weight_max_kg"`
			RestSec     int     `json:"rest_sec"`
			Tempo       string  `json:"tempo"`
			Notes       string  `json:"notes"`
		} `json:"exercises"`
	} `json:"days"`
}

// TestUpdateProgramPreservesDayIdentity — правка программы должна сохранять
// id существующих дней (позиционное сопоставление) и переносить все поля
// (описание, заметки дня, sets/rep/вес/notes предписаний), а не только
// name/exercise_id. Иначе тренировки, ссылающиеся на program_day_id дня,
// теряют привязку (ON DELETE SET NULL), а описание/заметки/предписания
// стираются в 0/NULL/''.
func TestUpdateProgramPreservesDayIdentity(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	create := ts.PostJSON(t, "/api/v1/programs", map[string]any{
		"name":        "Черновик",
		"description": "старое описание",
		"days": []map[string]any{
			{
				"name": "День 1", "notes": "старая заметка дня",
				"exercises": []map[string]any{
					{
						"exercise_name": "Присед в Смите", "sets": 3, "rep_min": 6, "rep_max": 10,
						"weight_min_kg": 60, "weight_max_kg": 80, "rest_sec": 90, "tempo": "3-0-1",
						"notes": "старая заметка упражнения",
					},
				},
			},
			{"name": "День 2", "exercises": []map[string]any{{"exercise_name": "Жим гантелей лёжа"}}},
		},
	})
	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", create.StatusCode)
	}
	var prog struct {
		ID   int64 `json:"id"`
		Days []struct {
			ID int64 `json:"id"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, create, &prog)
	if len(prog.Days) != 2 {
		t.Fatalf("ожидалось 2 дня, получено %d", len(prog.Days))
	}
	day1ID, day2ID := prog.Days[0].ID, prog.Days[1].ID

	// тренировка привязана к первому дню
	wkResp := ts.PostJSON(t, "/api/v1/workouts", map[string]any{"program_day_id": day1ID})
	if wkResp.StatusCode != http.StatusCreated {
		t.Fatalf("create workout: status = %d, want 201", wkResp.StatusCode)
	}
	var wk struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, wkResp, &wk)

	// правка: новое имя/описание дня и упражнения, новый набор полей предписания
	upd := putJSON(t, ts, "/api/v1/programs/"+itoa(prog.ID), map[string]any{
		"name":        "Фул бади",
		"description": "новое описание",
		"days": []map[string]any{
			{
				"name": "День A", "notes": "новая заметка дня",
				"exercises": []map[string]any{
					{
						"exercise_name": "Жим гантелей лёжа", "sets": 4, "rep_min": 8, "rep_max": 12,
						"weight_min_kg": 20, "weight_max_kg": 30, "rest_sec": 60, "tempo": "2-0-1",
						"notes": "новая заметка упражнения",
					},
				},
			},
			{"name": "День B", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
		},
	})
	if upd.StatusCode != http.StatusOK {
		t.Fatalf("update: status = %d, want 200", upd.StatusCode)
	}

	var got programWithFields
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)), &got)
	if got.Name != "Фул бади" || got.Description != "новое описание" {
		t.Errorf("программа = %+v, want name=Фул бади description=новое описание", got)
	}
	if len(got.Days) != 2 {
		t.Fatalf("дней = %d, want 2", len(got.Days))
	}
	if got.Days[0].ID != day1ID || got.Days[1].ID != day2ID {
		t.Errorf("id дней изменились: было %d/%d, стало %d/%d", day1ID, day2ID, got.Days[0].ID, got.Days[1].ID)
	}
	d1 := got.Days[0]
	if d1.Name != "День A" || d1.Notes != "новая заметка дня" {
		t.Errorf("день 1 = %+v", d1)
	}
	if len(d1.Exercises) != 1 {
		t.Fatalf("предписаний в дне 1 = %d, want 1", len(d1.Exercises))
	}
	e1 := d1.Exercises[0]
	if e1.Sets != 4 || e1.RepMin != 8 || e1.RepMax != 12 || e1.WeightMinKg != 20 || e1.WeightMaxKg != 30 ||
		e1.RestSec != 60 || e1.Tempo != "2-0-1" || e1.Notes != "новая заметка упражнения" {
		t.Errorf("предписание дня 1 = %+v", e1)
	}

	// тренировка не отвязалась от дня 1 (тот же id, содержимое дня заменилось)
	var wkAfter struct {
		ProgramDayID *int64 `json:"program_day_id"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/workouts/"+itoa(wk.ID)), &wkAfter)
	if wkAfter.ProgramDayID == nil || *wkAfter.ProgramDayID != day1ID {
		t.Errorf("тренировка отвязалась от дня программы: program_day_id = %v, want %d", wkAfter.ProgramDayID, day1ID)
	}
}

// TestUpdateProgramDayCountChanges — уменьшение числа дней удаляет лишние,
// увеличение — добавляет новые, id общих дней не меняется.
func TestUpdateProgramDayCountChanges(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	create := ts.PostJSON(t, "/api/v1/programs", map[string]any{
		"name": "П",
		"days": []map[string]any{
			{"name": "День 1", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
			{"name": "День 2", "exercises": []map[string]any{{"exercise_name": "Жим гантелей лёжа"}}},
		},
	})
	var prog struct {
		ID   int64 `json:"id"`
		Days []struct {
			ID int64 `json:"id"`
		} `json:"days"`
	}
	testutil.DecodeJSON(t, create, &prog)
	day1ID := prog.Days[0].ID

	// уменьшили до 1 дня — второй удалён
	shrink := putJSON(t, ts, "/api/v1/programs/"+itoa(prog.ID), map[string]any{
		"name": "П",
		"days": []map[string]any{
			{"name": "День 1", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
		},
	})
	if shrink.StatusCode != http.StatusOK {
		t.Fatalf("shrink: status = %d, want 200", shrink.StatusCode)
	}
	var afterShrink programWithFields
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)), &afterShrink)
	if len(afterShrink.Days) != 1 || afterShrink.Days[0].ID != day1ID {
		t.Fatalf("после уменьшения дней = %+v, want 1 день с id %d", afterShrink.Days, day1ID)
	}

	// увеличили до 3 дней — первый сохранил id, добавились новые
	grow := putJSON(t, ts, "/api/v1/programs/"+itoa(prog.ID), map[string]any{
		"name": "П",
		"days": []map[string]any{
			{"name": "День 1", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
			{"name": "День 2", "exercises": []map[string]any{{"exercise_name": "Жим гантелей лёжа"}}},
			{"name": "День 3", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}},
		},
	})
	if grow.StatusCode != http.StatusOK {
		t.Fatalf("grow: status = %d, want 200", grow.StatusCode)
	}
	var afterGrow programWithFields
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/programs/"+itoa(prog.ID)), &afterGrow)
	if len(afterGrow.Days) != 3 {
		t.Fatalf("после увеличения дней = %d, want 3", len(afterGrow.Days))
	}
	if afterGrow.Days[0].ID != day1ID {
		t.Errorf("id первого дня изменился: было %d, стало %d", day1ID, afterGrow.Days[0].ID)
	}
	if afterGrow.Days[1].Name != "День 2" || afterGrow.Days[2].Name != "День 3" {
		t.Errorf("новые дни = %+v", afterGrow.Days[1:])
	}
}

func TestUpdateProgramOwnership(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	resp := ts.PostJSON(t, "/api/v1/programs", map[string]any{
		"name": "Моя", "days": []map[string]any{{"name": "Д", "exercises": []map[string]any{{"exercise_name": "Присед в Смите"}}}},
	})
	var prog struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &prog)

	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, _ := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "chuzhak", "password": "надёжный-пароль"}))
	reg.Body.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/programs/"+itoa(prog.ID),
		mustJSON(map[string]any{"name": "Взлом", "days": []map[string]any{}}))
	req.Header.Set("Content-Type", "application/json")
	r, err := other.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	if r.StatusCode != http.StatusNotFound {
		t.Errorf("чужой PUT: status = %d, want 404", r.StatusCode)
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
