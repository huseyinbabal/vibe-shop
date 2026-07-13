package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vibe-shop/internal/httpx"
)

func TestWriteJSON_SetsStatusContentTypeAndBody(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteJSON(rec, http.StatusCreated, map[string]int{"id": 7})

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusCreated)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]int
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["id"] != 7 {
		t.Errorf("body[id] = %d, want 7", body["id"])
	}
}

func TestWriteError_WrapsMessageInErrorField(t *testing.T) {
	rec := httptest.NewRecorder()

	httpx.WriteError(rec, http.StatusNotFound, "product not found")

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "product not found" {
		t.Errorf("body[error] = %q, want %q", body["error"], "product not found")
	}
}
