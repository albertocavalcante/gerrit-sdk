// Deprecated: gerrit-sdk is superseded and no longer maintained. It wraps
// go-gerrit in a way that drops googlesource.com's ChangeInfo extensions, and
// its .gitcookies handling can return one host's credential to another. A
// successor that owns its own types is in progress and not yet published.
module github.com/albertocavalcante/gerrit-sdk

go 1.26.1

require github.com/andygrunwald/go-gerrit v1.1.1

require github.com/google/go-querystring v1.1.0 // indirect
