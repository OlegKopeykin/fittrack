package server_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func createOwnExercise(t *testing.T, ts *testutil.TestServer, body map[string]any) int64 {
	t.Helper()
	resp := ts.PostJSON(t, "/api/v1/exercises", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create exercise: status = %d, want 201", resp.StatusCode)
	}
	var ex struct {
		ID int64 `json:"id"`
	}
	testutil.DecodeJSON(t, resp, &ex)
	return ex.ID
}

func TestCreateExerciseWithFields(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)

	id := createOwnExercise(t, ts, map[string]any{
		"name": "Жим штанги лёжа", "muscle_group": "chest", "kind": "compound",
		"equipment": "barbell", "secondary_muscles": []string{"triceps", "delts-side"},
		"instructions": "Опустить к груди, выжать.", "video_url": "https://example.com/v",
	})

	var got struct {
		Equipment        string   `json:"equipment"`
		SecondaryMuscles []string `json:"secondary_muscles"`
		Instructions     string   `json:"instructions"`
		VideoURL         string   `json:"video_url"`
		HasImage         bool     `json:"has_image"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises/"+itoa(id)), &got)
	if got.Equipment != "barbell" || got.Instructions == "" || got.VideoURL == "" {
		t.Errorf("поля не сохранились: %+v", got)
	}
	if len(got.SecondaryMuscles) != 2 {
		t.Errorf("вторичных мышц = %d, want 2", len(got.SecondaryMuscles))
	}
	if got.HasImage {
		t.Error("has_image должно быть false без картинки")
	}
}

func TestCreateExerciseBadEquipment(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	resp := ts.PostJSON(t, "/api/v1/exercises", map[string]any{
		"name": "X", "muscle_group": "chest", "kind": "compound", "equipment": "лапшерезка",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func putImage(t *testing.T, ts *testutil.TestServer, id int64, ct string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/exercises/"+itoa(id)+"/image", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", ct)
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestExerciseImageLifecycle(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	id := createOwnExercise(t, ts, map[string]any{"name": "Тяга-X", "muscle_group": "lats", "kind": "compound"})

	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}
	if resp := putImage(t, ts, id, "image/png", png); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("upload: status = %d, want 204", resp.StatusCode)
	}

	img := ts.Get(t, "/api/v1/exercises/"+itoa(id)+"/image")
	if img.StatusCode != http.StatusOK {
		t.Fatalf("get image: status = %d, want 200", img.StatusCode)
	}
	if ct := img.Header.Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}

	var ex struct {
		HasImage bool   `json:"has_image"`
		ImageURL string `json:"image_url"`
	}
	testutil.DecodeJSON(t, ts.Get(t, "/api/v1/exercises/"+itoa(id)), &ex)
	if !ex.HasImage || ex.ImageURL == "" {
		t.Errorf("после загрузки has_image/image_url = %+v", ex)
	}

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/exercises/"+itoa(id)+"/image", nil)
	del, _ := ts.Client.Do(req)
	del.Body.Close()
	if del.StatusCode != http.StatusNoContent {
		t.Fatalf("delete image: status = %d, want 204", del.StatusCode)
	}
	if after := ts.Get(t, "/api/v1/exercises/"+itoa(id)+"/image"); after.StatusCode != http.StatusNotFound {
		t.Errorf("после удаления get image: status = %d, want 404", after.StatusCode)
	}
}

func TestExerciseImageRejectsNonImage(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	id := createOwnExercise(t, ts, map[string]any{"name": "Y", "muscle_group": "chest", "kind": "compound"})
	if resp := putImage(t, ts, id, "image/bmp", []byte{1, 2, 3}); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("image/bmp: status = %d, want 415", resp.StatusCode)
	}
}

func TestGlobalExerciseImageForbidden(t *testing.T) {
	ts := testutil.NewTestServer(t, nil)
	ownerSession(t, ts)
	gid := anExerciseID(t, ts, "Присед")
	if resp := putImage(t, ts, gid, "image/png", []byte{0x89, 0x50}); resp.StatusCode != http.StatusForbidden {
		t.Errorf("картинка на глобальное упражнение: status = %d, want 403", resp.StatusCode)
	}
}
