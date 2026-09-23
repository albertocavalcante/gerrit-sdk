package gerritsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// gerritResponse wraps response with the magic prefix Gerrit uses.
func gerritResponse(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return ")]}'\n" + string(b)
}

func TestQueryChanges(t *testing.T) {
	changes := []map[string]any{
		{
			"_number": 318050,
			"project": "bazel",
			"branch":  "master",
			"status":  "MERGED",
			"subject": "Bazel Docs : Fix various mdx syntax errors (part 2)",
			"owner":   map[string]any{"name": "Florian Weikert", "email": "fwe@google.com"},
			"created": "2026-03-20 21:12:05.000000000",
			"updated": "2026-03-23 12:52:31.000000000",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/changes/") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(gerritResponse(t, changes)))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := newTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}

	cls, err := client.QueryChanges(context.Background(), "project:bazel status:merged")
	if err != nil {
		t.Fatalf("QueryChanges: %v", err)
	}
	if len(cls) != 1 {
		t.Fatalf("got %d CLs, want 1", len(cls))
	}
	cl := cls[0]
	if cl.Number != 318050 {
		t.Errorf("Number = %d, want 318050", cl.Number)
	}
	if cl.Status != "MERGED" {
		t.Errorf("Status = %q, want MERGED", cl.Status)
	}
	if cl.Owner != "Florian Weikert" {
		t.Errorf("Owner = %q, want %q", cl.Owner, "Florian Weikert")
	}
}

func TestBuildQuery(t *testing.T) {
	q := NewQuery().
		Project("bazel").
		Status("merged").
		File("docs/").
		String()

	want := "project:bazel status:merged file:docs/"
	if q != want {
		t.Errorf("query = %q, want %q", q, want)
	}
}

func TestGetChangeDetail(t *testing.T) {
	change := map[string]any{
		"_number":          318050,
		"project":          "bazel",
		"branch":           "master",
		"status":           "MERGED",
		"subject":          "Bazel Docs : Fix various mdx syntax errors (part 2)",
		"owner":            map[string]any{"name": "fwe"},
		"created":          "2026-03-20 21:12:05.000000000",
		"updated":          "2026-03-23 12:52:31.000000000",
		"current_revision": "abc123",
		"revisions": map[string]any{
			"abc123": map[string]any{
				"files": map[string]any{
					"/COMMIT_MSG":         map[string]any{},
					"docs/run/build.mdx":  map[string]any{"lines_inserted": 5, "lines_deleted": 3},
					"docs/query/lang.mdx": map[string]any{"lines_inserted": 10, "lines_deleted": 8},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(gerritResponse(t, change)))
	}))
	defer server.Close()

	client, err := newTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}

	cl, err := client.GetChangeDetail(context.Background(), "318050")
	if err != nil {
		t.Fatalf("GetChangeDetail: %v", err)
	}
	if cl.Number != 318050 {
		t.Errorf("Number = %d, want 318050", cl.Number)
	}
	if len(cl.Files) != 2 {
		t.Errorf("Files count = %d, want 2 (excluding /COMMIT_MSG)", len(cl.Files))
	}
}

func TestGetChangeFiles(t *testing.T) {
	files := map[string]any{
		"/COMMIT_MSG":               map[string]any{},
		"docs/extending/config.mdx": map[string]any{"status": "M", "lines_inserted": 5, "lines_deleted": 3},
		"docs/run/build.mdx":        map[string]any{"status": "M", "lines_inserted": 10, "lines_deleted": 8},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(gerritResponse(t, files)))
	}))
	defer server.Close()

	client, err := newTestClient(t, server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.GetChangeFiles(context.Background(), "318050")
	if err != nil {
		t.Fatalf("GetChangeFiles: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("got %d files, want 2 (excluding /COMMIT_MSG)", len(result))
	}
}

func newTestClient(t *testing.T, server *httptest.Server) (*Client, error) {
	t.Helper()
	return newClientFromURL(context.Background(), server.URL+"/", server.Client())
}
