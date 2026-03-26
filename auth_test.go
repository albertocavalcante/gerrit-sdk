package gerritsdk

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGitCookies(t *testing.T) {
	content := `# Netscape HTTP Cookie File
.googlesource.com	FALSE	/	TRUE	0	o	git-alice.googlesource.com=secret123token
.example.com	FALSE	/	TRUE	0	o	git-bob.example.com=othertoken
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitcookies")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	user, pass, err := parseGitCookies(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "git-alice.googlesource.com" {
		t.Errorf("user = %q, want %q", user, "git-alice.googlesource.com")
	}
	if pass != "secret123token" {
		t.Errorf("pass = %q, want %q", pass, "secret123token")
	}
}

func TestParseGitCookiesForHost(t *testing.T) {
	content := `.bazel-review.googlesource.com	FALSE	/	TRUE	0	o	git-alice.googlesource.com=bazeltoken
.chromium-review.googlesource.com	FALSE	/	TRUE	0	o	git-bob.googlesource.com=chromiumtoken
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitcookies")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	user, pass, err := ParseGitCookiesForHost(path, "bazel-review.googlesource.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != "git-alice.googlesource.com" {
		t.Errorf("user = %q, want %q", user, "git-alice.googlesource.com")
	}
	if pass != "bazeltoken" {
		t.Errorf("pass = %q, want %q", pass, "bazeltoken")
	}
}

func TestParseGitCookiesForHost_NotFound(t *testing.T) {
	content := `.example.com	FALSE	/	TRUE	0	o	git-user=token123
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitcookies")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := ParseGitCookiesForHost(path, "bazel-review.googlesource.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseGitCookies_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitcookies")
	if err := os.WriteFile(path, []byte("# empty\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := parseGitCookies(path)
	if err == nil {
		t.Fatal("expected error for empty cookies, got nil")
	}
}

func TestParseGitCookies_MissingFile(t *testing.T) {
	_, _, err := parseGitCookies("/nonexistent/.gitcookies")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseGitCookies_MalformedLines(t *testing.T) {
	content := `.googlesource.com	FALSE	/	TRUE	0	o	noequals
.googlesource.com	too-few-fields
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitcookies")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := parseGitCookies(path)
	if err == nil {
		t.Fatal("expected error for malformed lines, got nil")
	}
}

func TestGitCookiesAuth_ApplyError(t *testing.T) {
	auth := GitCookiesAuth{Path: "/nonexistent/.gitcookies"}
	_, err := auth.apply(http.DefaultTransport)
	if err == nil {
		t.Fatal("expected error when gitcookies file doesn't exist")
	}
	if !strings.Contains(err.Error(), "gitcookies") {
		t.Errorf("error should mention gitcookies, got: %v", err)
	}
}

func TestHTTPBasicAuth_Apply(t *testing.T) {
	auth := HTTPBasicAuth{Username: "user", Password: "pass"}
	transport, err := auth.apply(http.DefaultTransport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("transport should not be nil")
	}
}

func TestBearerTokenAuth_Apply(t *testing.T) {
	auth := BearerTokenAuth{Token: "tok123"}
	transport, err := auth.apply(http.DefaultTransport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("transport should not be nil")
	}
}
