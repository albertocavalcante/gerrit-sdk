// Deprecated: unmaintained; it can send one review host's credential to another, and silently drops googlesource fields. Do not use. See the README.
//
// The first line above is deliberately one long line: `go get` prints only the
// first line of a deprecation message, so anything wrapped onto a second line
// is invisible to the person being warned.
//
// Superseded by a client that owns its own types. That successor is not
// published yet, so this deliberately names no replacement path rather than
// one that fails to resolve.
module github.com/albertocavalcante/gerrit-sdk

go 1.26.1

require github.com/andygrunwald/go-gerrit v1.1.1

require github.com/google/go-querystring v1.1.0 // indirect
