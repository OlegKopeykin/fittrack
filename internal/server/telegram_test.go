package server_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

// fakeTG подменяет Bot API без сети.
type fakeTG struct {
	me      string
	chat    string
	meErr   error
	chatErr error
	sendErr error
	sent    int
	lastDoc []byte
}

func (f *fakeTG) GetMe(_ context.Context, _ string) (string, error) { return f.me, f.meErr }
func (f *fakeTG) ResolveChatID(_ context.Context, _ string) (string, error) {
	return f.chat, f.chatErr
}
func (f *fakeTG) SendDocument(_ context.Context, _, _, _ string, data []byte, _ string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sent++
	f.lastDoc = data
	return nil
}

func TestTelegramSetupAndSend(t *testing.T) {
	tg := &fakeTG{me: "oleg_fittrack_bot", chat: "123456"}
	ts := testutil.NewTestServerTG(t, nil, tg)
	ownerSession(t, ts)

	// сначала не настроено
	var st struct {
		Configured bool   `json:"configured"`
		ChatLinked bool   `json:"chat_linked"`
		Frequency  string `json:"frequency"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/profile/telegram"), &st)
	if st.Configured || st.ChatLinked {
		t.Fatalf("до настройки: %+v", st)
	}

	// задаём токен → валидируется через GetMe
	put := putJSON(t, ts, "/api/v1/profile/telegram", map[string]any{"bot_token": "111:AAA"})
	if put.StatusCode != http.StatusOK {
		t.Fatalf("set token: status = %d", put.StatusCode)
	}
	var afterToken struct {
		Configured  bool   `json:"configured"`
		BotUsername string `json:"bot_username"`
		ChatLinked  bool   `json:"chat_linked"`
	}
	testutil.DecodeJSON(t, put, &afterToken)
	if !afterToken.Configured || afterToken.BotUsername != "oleg_fittrack_bot" || afterToken.ChatLinked {
		t.Fatalf("после токена: %+v", afterToken)
	}

	// связываем чат (Start) → ResolveChatID
	link := ts.PostJSON(t, "/api/v1/profile/telegram/link", nil)
	if link.StatusCode != http.StatusOK {
		t.Fatalf("link: status = %d", link.StatusCode)
	}
	var linked struct {
		ChatLinked bool `json:"chat_linked"`
	}
	testutil.DecodeJSON(t, link, &linked)
	if !linked.ChatLinked {
		t.Fatal("чат не привязался")
	}

	// частота + включение
	pref := putJSON(t, ts, "/api/v1/profile/telegram", map[string]any{"frequency": "weekly", "enabled": true})
	var afterPref struct {
		Enabled   bool   `json:"enabled"`
		Frequency string `json:"frequency"`
	}
	testutil.DecodeJSON(t, pref, &afterPref)
	if !afterPref.Enabled || afterPref.Frequency != "weekly" {
		t.Errorf("настройки = %+v", afterPref)
	}

	// тест-отправка шлёт документ
	test := ts.PostJSON(t, "/api/v1/profile/telegram/test", nil)
	if test.StatusCode != http.StatusNoContent {
		t.Fatalf("test send: status = %d, want 204", test.StatusCode)
	}
	if tg.sent != 1 {
		t.Errorf("отправлено документов = %d, want 1", tg.sent)
	}
}

func TestTelegramBadToken(t *testing.T) {
	tg := &fakeTG{meErr: context.Canceled}
	ts := testutil.NewTestServerTG(t, nil, tg)
	ownerSession(t, ts)

	put := putJSON(t, ts, "/api/v1/profile/telegram", map[string]any{"bot_token": "плохой"})
	if put.StatusCode != http.StatusBadRequest {
		t.Errorf("плохой токен: status = %d, want 400", put.StatusCode)
	}
}

func TestExportOwnWorkoutsOnly(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	// у владельца одна тренировка с уникальной заметкой
	wid := newWorkout(t, ts)
	patchJSON(t, ts, "/api/v1/workouts/"+itoa(wid), map[string]any{"notes": "владелец-only"})

	// второй пользователь заводит свою тренировку
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, err := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "sosed2", "password": "надёжный-пароль"}))
	if err != nil {
		t.Fatal(err)
	}
	reg.Body.Close()
	cr, err := other.Post(ts.URL+"/api/v1/workouts", "application/json", mustJSON(map[string]any{}))
	if err != nil {
		t.Fatal(err)
	}
	var nb struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, cr, &nb)
	cr.Body.Close()
	pr, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/workouts/"+itoa(nb.ID), mustJSON(map[string]any{"notes": "сосед-only"}))
	pr.Header.Set("Content-Type", "application/json")
	presp, err := other.Do(pr)
	if err != nil {
		t.Fatal(err)
	}
	presp.Body.Close()

	// экспорт соседа содержит только его тренировку
	er, err := other.Get(ts.URL + "/api/v1/profile/export")
	if err != nil {
		t.Fatal(err)
	}
	var exp struct {
		Workouts []struct {
			Notes string `json:"notes"`
		} `json:"workouts"`
	}
	testutil.DecodeJSON(t, er, &exp)
	er.Body.Close()
	if len(exp.Workouts) != 1 || exp.Workouts[0].Notes != "сосед-only" {
		t.Fatalf("экспорт соседа = %+v, ожидалась только его тренировка", exp)
	}
}

func TestImportRoundtrip(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	// owner: завершённая тренировка с подходом
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 80, "reps": 8,
	})
	finish(t, ts, wid, "2026-07-10T12:00:00Z")

	var exp map[string]any
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/profile/export"), &exp)

	// второй пользователь импортирует бэкап
	jar, _ := cookiejar.New(nil)
	other := &http.Client{Jar: jar}
	code := ts.CreateInvite(t, "user", "")
	reg, err := other.Post(ts.URL+"/api/v1/auth/register", "application/json",
		mustJSON(map[string]string{"invite_code": code, "username": "restorer", "password": "надёжный-пароль"}))
	if err != nil {
		t.Fatal(err)
	}
	reg.Body.Close()

	post := func() struct{ Imported, Skipped int } {
		resp, err := other.Post(ts.URL+"/api/v1/profile/import", "application/json", mustJSON(exp))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("import: status = %d", resp.StatusCode)
		}
		var res struct{ Imported, Skipped int }
		testutil.DecodeJSON(t, resp, &res)
		return res
	}

	if res := post(); res.Imported != 1 {
		t.Fatalf("первый импорт: imported = %d, want 1", res.Imported)
	}

	// у второго пользователя появилась тренировка
	var page struct {
		Items []map[string]any `json:"items"`
	}
	wr, err := other.Get(ts.URL + "/api/v1/workouts")
	if err != nil {
		t.Fatal(err)
	}
	testutil.DecodeJSON(t, wr, &page)
	wr.Body.Close()
	if len(page.Items) != 1 {
		t.Errorf("тренировок после импорта = %d, want 1", len(page.Items))
	}

	// повторный импорт того же файла ничего не дублирует
	if res := post(); res.Imported != 0 || res.Skipped != 1 {
		t.Errorf("повторный импорт: imported=%d skipped=%d, want 0/1", res.Imported, res.Skipped)
	}
}

func TestExportDownload(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	// одна тренировка с подходом
	wid := newWorkout(t, ts)
	ex := anExerciseID(t, ts, "Присед")
	ts.PostJSON(t, "/api/v1/workouts/"+itoa(wid)+"/sets", map[string]any{
		"exercise_id": ex, "role": "working", "weight_kg": 80, "reps": 8,
	})

	resp := ts.Get(t, "/api/v1/profile/export")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export: status = %d", resp.StatusCode)
	}
	var exp struct {
		User     string `json:"user"`
		Workouts []struct {
			Sets []struct {
				Exercise string  `json:"exercise"`
				WeightKg float64 `json:"weight_kg"`
			} `json:"sets"`
		} `json:"workouts"`
	}
	testutil.DecodeJSON(t, resp, &exp)
	if len(exp.Workouts) != 1 || len(exp.Workouts[0].Sets) != 1 {
		t.Fatalf("экспорт = %+v", exp)
	}
	if exp.Workouts[0].Sets[0].Exercise == "" || exp.Workouts[0].Sets[0].WeightKg != 80 {
		t.Errorf("подход в экспорте = %+v", exp.Workouts[0].Sets[0])
	}
}
