package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRouter_Health(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	buf := make([]byte, 64)
	n, _ := res.Body.Read(buf)
	if got := strings.TrimSpace(string(buf[:n])); got != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}
