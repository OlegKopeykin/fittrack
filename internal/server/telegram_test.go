package server_test

import (
	"context"
	"net/http"
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
