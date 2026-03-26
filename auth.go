package gerritsdk

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// AuthMethod configures how the client authenticates with Gerrit.
type AuthMethod interface {
	apply(transport http.RoundTripper) (http.RoundTripper, error)
}

// GitCookiesAuth authenticates using a .gitcookies file,
// the standard method for googlesource.com instances.
type GitCookiesAuth struct {
	Path string // path to .gitcookies file; empty = ~/.gitcookies
}

func (a GitCookiesAuth) apply(transport http.RoundTripper) (http.RoundTripper, error) {
	path := a.Path
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve gitcookies path: %w", err)
		}
		path = home + "/.gitcookies"
	}
	user, pass, err := parseGitCookies(path)
	if err != nil {
		return nil, fmt.Errorf("gitcookies auth: %w", err)
	}
	return &basicAuthTransport{
		username:  user,
		password:  pass,
		transport: transport,
	}, nil
}

// HTTPBasicAuth authenticates with a username and password/token.
type HTTPBasicAuth struct {
	Username string
	Password string
}

func (a HTTPBasicAuth) apply(transport http.RoundTripper) (http.RoundTripper, error) {
	return &basicAuthTransport{
		username:  a.Username,
		password:  a.Password,
		transport: transport,
	}, nil
}

// BearerTokenAuth authenticates with an OAuth2 bearer token.
type BearerTokenAuth struct {
	Token string
}

func (a BearerTokenAuth) apply(transport http.RoundTripper) (http.RoundTripper, error) {
	return &bearerTransport{
		token:     a.Token,
		transport: transport,
	}, nil
}

// parseGitCookies reads a Netscape-format cookies file and extracts
// the first googlesource.com credential it finds.
//
// Format: host \t FALSE \t / \t TRUE \t 0 \t o \t git-user.googlesource.com=TOKEN
func parseGitCookies(path string) (user, pass string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("open gitcookies: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		host := fields[0]
		if !strings.Contains(host, ".googlesource.com") {
			continue
		}
		// The value field (index 6) is in the format: git-user.googlesource.com=TOKEN
		value := fields[6]
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			continue
		}
		return parts[0], parts[1], nil
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("scan gitcookies: %w", err)
	}
	return "", "", fmt.Errorf("no googlesource.com credential found in %s", path)
}

// ParseGitCookiesForHost reads a .gitcookies file and returns credentials
// for the specified host.
func ParseGitCookiesForHost(path, host string) (user, pass string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("open gitcookies: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		cookieHost := fields[0]
		// Match exact host or .host (domain cookie)
		if cookieHost != host && cookieHost != "."+host {
			continue
		}
		value := fields[6]
		parts := strings.SplitN(value, "=", 2)
		if len(parts) != 2 {
			continue
		}
		return parts[0], parts[1], nil
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("scan gitcookies: %w", err)
	}
	return "", "", fmt.Errorf("no credential for host %q in %s", host, path)
}

type basicAuthTransport struct {
	username  string
	password  string
	transport http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.SetBasicAuth(t.username, t.password)
	return t.transport.RoundTrip(req2)
}

type bearerTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+t.token)
	return t.transport.RoundTrip(req2)
}
