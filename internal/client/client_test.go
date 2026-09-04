package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkspaceHeader(t *testing.T) {
	t.Parallel()

	const workspaceID = "22222222-2222-4222-8222-222222222222"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-openroadie-workspace-id"); got != workspaceID {
			t.Fatalf("workspace header = %q, want %q", got, workspaceID)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := New(server.URL, "token", workspaceID, "test")
	if _, err := c.Get(context.Background(), "/api/actions"); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultWorkspaceOmitsHeader(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-openroadie-workspace-id"); got != "" {
			t.Fatalf("workspace header = %q, want it omitted", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := New(server.URL, "token", "", "test")
	if _, err := c.Get(context.Background(), "/api/actions"); err != nil {
		t.Fatal(err)
	}
}
