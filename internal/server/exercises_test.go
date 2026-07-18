package server_test

import (
	"net/http"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func TestExerciseNoteRoundtrip(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	ex := anExerciseID(t, ts, "Присед") // глобальное

	var g struct {
		Note string `json:"note"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises/"+itoa(ex)+"/note"), &g)
	if g.Note != "" {
		t.Errorf("заметка по умолчанию = %q, want пусто", g.Note)
	}

	put := putJSON(t, ts, "/api/v1/exercises/"+itoa(ex)+"/note", map[string]any{"note": "тянуть лопатками"})
	if put.StatusCode != http.StatusOK {
		t.Fatalf("put note: status = %d", put.StatusCode)
	}

	var g2 struct {
		Note string `json:"note"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises/"+itoa(ex)+"/note"), &g2)
	if g2.Note != "тянуть лопатками" {
		t.Errorf("заметка после PUT = %q", g2.Note)
	}
}

// ownerSession регистрирует owner-пользователя и оставляет сессию в jar.
func ownerSession(t *testing.T, ts *testutil.TestServer) {
	t.Helper()
	ts.Register(t, ts.CreateInvite(t, "owner", ""), "owner1", "надёжный-пароль")
}

func TestListMuscleGroups(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.Get(t, "/api/v1/muscle-groups")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var groups []struct {
		Slug   string `json:"slug"`
		NameRu string `json:"name_ru"`
	}
	testutil.DecodeJSON(t, resp, &groups)
	if len(groups) != 16 {
		t.Errorf("групп %d, want 16", len(groups))
	}
}

func TestExercisesRequireAuth(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	resp := ts.Get(t, "/api/v1/exercises")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

type exDTO struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Global   bool     `json:"global"`
	Archived bool     `json:"archived"`
	Aliases  []string `json:"aliases"`
}

func TestListGlobalCatalog(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.Get(t, "/api/v1/exercises")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var list []exDTO
	testutil.DecodeJSON(t, resp, &list)
	if len(list) < 50 {
		t.Errorf("упражнений %d, want >= 50", len(list))
	}
	found := false
	for _, e := range list {
		if e.Name == "Присед в Смите" {
			found = true
			if !e.Global {
				t.Error("«Присед в Смите» должно быть глобальным")
			}
		}
	}
	if !found {
		t.Error("в каталоге нет «Присед в Смите»")
	}
}

func TestSearchByAlias(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.Get(t, "/api/v1/exercises?q=РДЛ")
	var list []exDTO
	testutil.DecodeJSON(t, resp, &list)
	if len(list) != 1 || list[0].Name != "Румынская тяга с гантелями" {
		t.Fatalf("поиск РДЛ → %+v, want [Румынская тяга с гантелями]", list)
	}
}

func TestFilterByMuscleGroup(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	resp := ts.Get(t, "/api/v1/exercises?muscle_group=cardio")
	var list []exDTO
	testutil.DecodeJSON(t, resp, &list)
	if len(list) == 0 {
		t.Fatal("cardio-группа пуста")
	}
	for _, e := range list {
		if e.Kind != "cardio" {
			t.Errorf("в фильтре cardio попало %q (kind=%s)", e.Name, e.Kind)
		}
	}
}

func TestCreateExercise(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	body := map[string]any{"name": "Моё упражнение", "muscle_group": "chest", "kind": "isolation"}
	resp := ts.PostJSON(t, "/api/v1/exercises", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want 201", resp.StatusCode)
	}
	var created exDTO
	testutil.DecodeJSON(t, resp, &created)
	if created.Global {
		t.Error("своё упражнение не должно быть глобальным")
	}

	// дубликат по имени → 409
	dup := ts.PostJSON(t, "/api/v1/exercises", body)
	if dup.StatusCode != http.StatusConflict {
		t.Fatalf("дубликат: status = %d, want 409", dup.StatusCode)
	}
}

func TestGlobalExerciseImmutable(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	var list []exDTO
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises?q=Присед"), &list)
	if len(list) == 0 {
		t.Fatal("не найдено глобальное упражнение")
	}
	id := itoa(list[0].ID)

	resp := ts.PostJSON(t, "/api/v1/exercises/"+id+"/aliases", map[string]string{"alias": "х"})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("изменение глобального: status = %d, want 403", resp.StatusCode)
	}
}

func TestArchiveOwnExercise(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	var created exDTO
	testutil.DecodeJSON(t, ts.PostJSON(t, "/api/v1/exercises",
		map[string]any{"name": "Временное", "muscle_group": "core", "kind": "isometric"}), &created)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/exercises/"+itoa(created.ID), nil)
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("archive: status = %d, want 204", resp.StatusCode)
	}

	var active []exDTO
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises?q=Временное"), &active)
	if len(active) != 0 {
		t.Error("архивное упражнение не должно быть в обычном списке")
	}
	var withArchived []exDTO
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises?q=Временное&include_archived=1"), &withArchived)
	if len(withArchived) != 1 {
		t.Errorf("с include_archived=1 ожидалось 1, получено %d", len(withArchived))
	}
}
