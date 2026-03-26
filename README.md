# htmlsanitizer

[![Go Reference](https://pkg.go.dev/badge/github.com/sym01/htmlsanitizer.svg)](https://pkg.go.dev/github.com/sym01/htmlsanitizer)
[![Go](https://github.com/SYM01/htmlsanitizer/workflows/Go/badge.svg)](https://github.com/SYM01/htmlsanitizer/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/SYM01/htmlsanitizer/branch/master/graph/badge.svg)](https://codecov.io/gh/SYM01/htmlsanitizer)

A fast, allowlist-based HTML sanitizer written in Go. Secure-by-default with a built-in allowlist that strips dangerous HTML content.

- **Fast** -- O(n) time complexity via an internal Finite State Machine
- **Customizable** -- modify the allowlist, add/remove tags, or disable all HTML
- **Zero dependencies**

> Also available in **Rust / npm**: [htmlsanitizer-rs](https://github.com/SYM01/htmlsanitizer-rs)

## Install

```bash
go get github.com/sym01/htmlsanitizer
```

## Usage

### Basic

```go
sanitizedHTML, err := htmlsanitizer.SanitizeString(rawHTML)
```

### Disable the `id` attribute globally

```go
s := htmlsanitizer.NewHTMLSanitizer()
s.GlobalAttr = []string{"class"}

sanitizedHTML, err := s.SanitizeString(rawHTML)
```

### Add or remove tags

```go
s := htmlsanitizer.NewHTMLSanitizer()
// remove <a> tag
s.RemoveTag("a")

// add a custom tag
s.AllowList.Tags = append(s.AllowList.Tags, &htmlsanitizer.Tag{
    Name: "my-tag",
    Attr: []string{"my-attr"},
})

sanitizedHTML, err := s.SanitizeString(rawHTML)
```

### Strip all HTML

```go
s := htmlsanitizer.NewHTMLSanitizer()
s.AllowList = nil

sanitizedHTML, err := s.SanitizeString(rawHTML)
```

## Testing

```bash
go test ./...              # run tests
go test -race ./...        # with race detection
go test -bench=. -benchmem ./...  # benchmarks
go test -fuzz=FuzzSanitize -fuzztime=30s .  # fuzz testing
```
