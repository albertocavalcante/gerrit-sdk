# gerrit-sdk

> [!CAUTION]
> **Deprecated and unmaintained as of 2026-09-22. Do not use this module.**
>
> A successor that owns its own types is in progress; it is not yet published,
> so there is no replacement path to point at yet. This notice will be updated
> when there is one.
>
> The repository stays up, read-only, so that `v0.1.0` keeps resolving for
> anyone who already depends on it. It is not being deleted and `v0.1.0` is not
> retracted — it works, it is merely wrong in the ways listed below.
>
> **Known defects, any of which produces a silently wrong result:**
>
> 1. **Cross-host credential leak.** `parseGitCookies` matches with
>    `strings.Contains(host, ".googlesource.com")` and returns the *first*
>    googlesource credential in the file regardless of which host was asked for.
>    With cookies for more than one review host, it sends the wrong identity to
>    the wrong server.
> 2. **`ParseGitCookiesForHost` fails on a stock Google `.gitcookies`.** It
>    matches only `host` or `"."+host`, so the fleet-wide `.googlesource.com`
>    line that Google actually issues never matches a specific review host. It
>    also never inspects the cookie *name* and never checks expiry.
> 3. **googlesource extensions are dropped.** `triplet_id`,
>    `virtual_id_number` and `full_branch` are returned by googlesource and are
>    absent from the underlying `go-gerrit` types, with no
>    `DisallowUnknownFields`, so they vanish without an error.
> 4. **`_more_changes` is discarded**, so a caller cannot tell a truncated
>    result set from a complete one.
> 5. **`NewClient` mutates the caller's `http.Client`.** Two clients built from
>    one `http.Client` stack auth transports and send both credentials.
> 6. **`NewAnonymousClient` accepts and silently ignores `WithAuth`** — directly
>    contradicting the claim below that the SDK never silently falls back to
>    unauthenticated access.
> 7. **Nondeterministic output.** `GetChangeDetail`, `GetChangeFiles` and
>    `ListProjects` build slices by ranging over a map, so ordering changes on
>    every call.
> 8. **The query builder does not quote values**, so `Project("my project")`
>    silently becomes two query terms, and `file:` is documented as a prefix
>    match when Gerrit treats it as exact — the example below returns nothing.

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
