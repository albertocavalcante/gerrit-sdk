package gerritsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProjects(t *testing.T) {
	projects := map[string]any{
		"bazel": map[string]any{
			"id":          "bazel",
			"state":       "ACTIVE",
			"description": "Bazel build tool",
		},
		"abseil-cpp": map[string]any{
			"id":    "abseil-cpp",
			"state": "ACTIVE",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(gerritResponse(t, projects)))
	}))
	defer server.Close()

	client, err := newProjectsTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d projects, want 2", len(result))
	}
}

func TestGetProject(t *testing.T) {
	project := map[string]any{
		"id":          "bazel",
		"state":       "ACTIVE",
		"description": "Bazel build tool",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(gerritResponse(t, project)))
	}))
	defer server.Close()

	client, err := newProjectsTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.GetProject(context.Background(), "bazel")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if result.ID != "bazel" {
		t.Errorf("ID = %q, want %q", result.ID, "bazel")
	}
	if result.State != "ACTIVE" {
		t.Errorf("State = %q, want %q", result.State, "ACTIVE")
	}
}

func newProjectsTestClient(t *testing.T, server *httptest.Server) (*Client, error) {
	t.Helper()
	return newClientFromURL(context.Background(), server.URL+"/", server.Client())
}
