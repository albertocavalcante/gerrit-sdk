# gerrit-sdk

A high-level Go client for [Gerrit Code Review](https://www.gerritcodereview.com/), with first-class support for [Googlesource.com](https://gerrit-review.googlesource.com/) instances.

Built on top of [andygrunwald/go-gerrit](https://github.com/andygrunwald/go-gerrit), this SDK provides:

- **Googlesource authentication** via `.gitcookies`, HTTP Basic, or Bearer tokens
- **Fluent query builder** for Gerrit's change search syntax
- **Simplified types** (`CL`, `CLFile`, `FileDiff`, `Project`) mapped from go-gerrit internals
- **Diff retrieval** with structured content segments

## Install

```bash
go get github.com/albertocavalcante/gerrit-sdk
```

## Usage

### Anonymous access (public instances)

```go
package main

import (
    "context"
    "fmt"

    gerritsdk "github.com/albertocavalcante/gerrit-sdk"
)

func main() {
    ctx := context.Background()
    client, err := gerritsdk.NewAnonymousClient(ctx, "bazel-review.googlesource.com")
    if err != nil {
        panic(err)
    }

    query := gerritsdk.NewQuery().
        Project("bazel").
        Status("merged").
        File("docs/").
        String()

    cls, err := client.QueryChanges(ctx, query, gerritsdk.WithLimit(10))
    if err != nil {
        panic(err)
    }

    for _, cl := range cls {
        fmt.Printf("CL/%d  %s  (%s)\n", cl.Number, cl.Subject, cl.Owner)
    }
}
```

### Authenticated access (gitcookies)

```go
client, err := gerritsdk.NewClient(ctx, "bazel-review.googlesource.com",
    gerritsdk.WithAuth(gerritsdk.GitCookiesAuth{}), // reads ~/.gitcookies
)
```

### Authenticated access (HTTP Basic)

```go
client, err := gerritsdk.NewClient(ctx, "review.example.com",
    gerritsdk.WithAuth(gerritsdk.HTTPBasicAuth{
        Username: "user",
        Password: os.Getenv("GERRIT_PASSWORD"),
    }),
)
```

### Change details and files

```go
cl, err := client.GetChangeDetail(ctx, "318050")
fmt.Println(cl.Subject, cl.Files)

files, err := client.GetChangeFiles(ctx, "318050")
for _, f := range files {
    fmt.Printf("  %s (+%d -%d)\n", f.Path, f.Insertions, f.Deletions)
}
```

### File diffs

```go
diff, err := client.GetChangeDiff(ctx, "318050", "docs/run/build.mdx")
for _, seg := range diff.Content {
    // seg.Common, seg.OldOnly, seg.NewOnly
}
```

### Query builder

```go
q := gerritsdk.NewQuery().
    Project("bazel").
    Status("merged").
    After(time.Now().Add(-30 * 24 * time.Hour)).
    File("docs/").
    Owner("self").
    String()
// "project:bazel status:merged after:2026-02-24 file:docs/ owner:self"
```

## Authentication

| Method | Use case |
|--------|----------|
| `GitCookiesAuth{Path: "~/.gitcookies"}` | Googlesource.com (default path: `~/.gitcookies`) |
| `HTTPBasicAuth{Username, Password}` | Standard Gerrit with HTTP auth |
| `BearerTokenAuth{Token}` | OAuth2 / service accounts |
| None (`NewAnonymousClient`) | Public read-only access |

Auth errors are returned immediately from `NewClient` -- the SDK never silently falls back to unauthenticated access.

## Advanced usage

Access the underlying [go-gerrit](https://github.com/andygrunwald/go-gerrit) client for endpoints not yet wrapped:

```go
inner := client.Inner()
reviewers, _, err := inner.Changes.ListReviewers(ctx, "318050")
```

## License

[MIT](LICENSE)
