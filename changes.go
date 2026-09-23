package gerritsdk

import (
	"context"
	"fmt"
	"strings"
	"time"

	gogerrit "github.com/andygrunwald/go-gerrit"
)

// QueryChanges searches for changes matching the given Gerrit query string.
// Example queries:
//
//	"project:bazel status:merged file:docs/"
//	"status:open owner:self"
func (c *Client) QueryChanges(ctx context.Context, query string, opts ...QueryOption) ([]CL, error) {
	cfg := &queryConfig{limit: 25}
	for _, o := range opts {
		o(cfg)
	}

	gerritOpts := &gogerrit.QueryChangeOptions{}
	gerritOpts.Query = []string{query}
	gerritOpts.Limit = cfg.limit
	gerritOpts.Start = cfg.offset
	for _, f := range cfg.fields {
		gerritOpts.AdditionalFields = append(gerritOpts.AdditionalFields, f)
	}

	changes, _, err := c.inner.Changes.QueryChanges(ctx, gerritOpts)
	if err != nil {
		return nil, fmt.Errorf("query changes: %w", err)
	}

	result := make([]CL, 0, len(*changes))
	for _, ch := range *changes {
		cl := c.changeInfoToCL(&ch)
		result = append(result, cl)
	}
	return result, nil
}

// GetChangeDetail returns detailed information about a single change.
func (c *Client) GetChangeDetail(ctx context.Context, changeID string) (*CL, error) {
	opts := &gogerrit.ChangeOptions{}
	opts.AdditionalFields = []string{"CURRENT_REVISION", "CURRENT_FILES", "DETAILED_ACCOUNTS"}

	ch, _, err := c.inner.Changes.GetChange(ctx, changeID, opts)
	if err != nil {
		return nil, fmt.Errorf("get change %s: %w", changeID, err)
	}

	cl := c.changeInfoToCL(ch)

	// Populate files from current revision.
	if ch.CurrentRevision != "" {
		if rev, ok := ch.Revisions[ch.CurrentRevision]; ok {
			files := make([]string, 0, len(rev.Files))
			for path := range rev.Files {
				if path == "/COMMIT_MSG" {
					continue
				}
				files = append(files, path)
			}
			cl.Files = files
		}
	}

	return &cl, nil
}

// GetChangeFiles returns the list of files modified in the current patchset.
func (c *Client) GetChangeFiles(ctx context.Context, changeID string) ([]CLFile, error) {
	filesMap, _, err := c.inner.Changes.ListFiles(ctx, changeID, "current", nil)
	if err != nil {
		return nil, fmt.Errorf("list files for %s: %w", changeID, err)
	}

	result := make([]CLFile, 0, len(filesMap))
	for path, info := range filesMap {
		if path == "/COMMIT_MSG" {
			continue
		}
		result = append(result, CLFile{
			Path:       path,
			Status:     info.Status,
			OldPath:    info.OldPath,
			Insertions: info.LinesInserted,
			Deletions:  info.LinesDeleted,
		})
	}
	return result, nil
}

// GetChangeDiff returns the diff for a specific file in a change.
func (c *Client) GetChangeDiff(ctx context.Context, changeID, filePath string) (*FileDiff, error) {
	diff, _, err := c.inner.Changes.GetDiff(ctx, changeID, "current", filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("get diff for %s in %s: %w", filePath, changeID, err)
	}

	fd := &FileDiff{Path: filePath}
	for _, seg := range diff.Content {
		dc := DiffContent{}
		if len(seg.AB) > 0 {
			dc.Common = seg.AB
		}
		if len(seg.A) > 0 {
			dc.OldOnly = seg.A
		}
		if len(seg.B) > 0 {
			dc.NewOnly = seg.B
		}
		fd.Content = append(fd.Content, dc)
	}
	return fd, nil
}

// GetChangeCommitMessage returns the commit message for a change.
func (c *Client) GetChangeCommitMessage(ctx context.Context, changeID string) (string, error) {
	commit, _, err := c.inner.Changes.GetCommit(ctx, changeID, "current", nil)
	if err != nil {
		return "", fmt.Errorf("get commit for %s: %w", changeID, err)
	}
	return commit.Message, nil
}

// BuildQuery is a helper for constructing Gerrit query strings.
type BuildQuery struct {
	parts []string
}

// NewQuery starts building a query.
//
// Deprecated: this module is unmaintained. The builder does not quote values,
// so any value containing a space is silently split into two query terms, and
// callers can inject arbitrary Gerrit operators through Project, Owner or File.
// See the README.
func NewQuery() *BuildQuery { return &BuildQuery{} }

// Project restricts to a project.
func (q *BuildQuery) Project(name string) *BuildQuery {
	q.parts = append(q.parts, "project:"+name)
	return q
}

// Status restricts to a change status.
func (q *BuildQuery) Status(status string) *BuildQuery {
	q.parts = append(q.parts, "status:"+status)
	return q
}

// File restricts to changes touching a file path (prefix match).
func (q *BuildQuery) File(path string) *BuildQuery {
	q.parts = append(q.parts, "file:"+path)
	return q
}

// After restricts to changes updated after a time.
func (q *BuildQuery) After(t time.Time) *BuildQuery {
	q.parts = append(q.parts, "after:"+t.Format("2006-01-02"))
	return q
}

// Before restricts to changes updated before a time.
func (q *BuildQuery) Before(t time.Time) *BuildQuery {
	q.parts = append(q.parts, "before:"+t.Format("2006-01-02"))
	return q
}

// Owner restricts to changes owned by a user.
func (q *BuildQuery) Owner(user string) *BuildQuery {
	q.parts = append(q.parts, "owner:"+user)
	return q
}

// Raw adds a raw query fragment.
func (q *BuildQuery) Raw(fragment string) *BuildQuery {
	q.parts = append(q.parts, fragment)
	return q
}

// String returns the assembled query.
func (q *BuildQuery) String() string {
	return strings.Join(q.parts, " ")
}

// changeInfoToCL converts a go-gerrit ChangeInfo to our CL type.
func (c *Client) changeInfoToCL(ch *gogerrit.ChangeInfo) CL {
	cl := CL{
		Number:   ch.Number,
		Project:  ch.Project,
		Branch:   ch.Branch,
		ChangeID: ch.ChangeID,
		Subject:  ch.Subject,
		Status:   ch.Status,
		Created:  ch.Created.Time,
		Updated:  ch.Updated.Time,
		URL:      c.changeURL(ch.Project, ch.Number),
	}

	// Owner is a value type (AccountInfo), check if populated.
	if ch.Owner.Name != "" {
		cl.Owner = ch.Owner.Name
	} else if ch.Owner.Email != "" {
		cl.Owner = ch.Owner.Email
	}

	if ch.Submitted != nil {
		cl.Submitted = ch.Submitted.Time
	}

	cl.Mergeable = ch.Mergeable

	return cl
}
