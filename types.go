// Package gerritsdk provides a high-level Go client for Gerrit Code Review
// instances, with first-class support for Googlesource.com authentication.
package gerritsdk

import "time"

// CL represents a Gerrit change (changelist).
type CL struct {
	Number    int       `json:"number"`
	Project   string    `json:"project"`
	Branch    string    `json:"branch"`
	ChangeID  string    `json:"change_id"`
	Subject   string    `json:"subject"`
	Status    string    `json:"status"` // NEW, MERGED, ABANDONED
	Owner     string    `json:"owner"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Submitted time.Time `json:"submitted,omitempty"`
	Mergeable bool      `json:"mergeable,omitempty"`
	URL       string    `json:"url"`
	Files     []string  `json:"files,omitempty"`
}

// CLFile represents a file changed in a CL revision.
type CLFile struct {
	Path       string `json:"path"`
	Status     string `json:"status"` // ADDED, MODIFIED, DELETED, RENAMED, COPIED
	OldPath    string `json:"old_path,omitempty"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
}

// DiffContent represents a segment of a unified diff returned by Gerrit.
type DiffContent struct {
	Common  []string `json:"ab,omitempty"` // lines common to both
	OldOnly []string `json:"a,omitempty"`  // lines only in old version
	NewOnly []string `json:"b,omitempty"`  // lines only in new version
}

// FileDiff represents the diff for a single file.
type FileDiff struct {
	Path    string        `json:"path"`
	Content []DiffContent `json:"content"`
}

// Project represents a Gerrit project (repository).
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	State       string `json:"state"` // ACTIVE, READ_ONLY, HIDDEN
	WebLinks    []Link `json:"web_links,omitempty"`
}

// Link is a Gerrit web link.
type Link struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// QueryOption modifies a change query.
type QueryOption func(*queryConfig)

type queryConfig struct {
	limit  int
	offset int
	fields []string
}

// WithLimit sets the maximum number of results.
func WithLimit(n int) QueryOption {
	return func(c *queryConfig) { c.limit = n }
}

// WithOffset sets the result offset for pagination.
func WithOffset(n int) QueryOption {
	return func(c *queryConfig) { c.offset = n }
}

// WithFields requests additional fields (e.g. "CURRENT_REVISION", "ALL_FILES").
func WithFields(fields ...string) QueryOption {
	return func(c *queryConfig) { c.fields = append(c.fields, fields...) }
}
