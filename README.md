# Golang HTML Sanitizer

![Go](https://github.com/SYM01/htmlsanitizer/workflows/Go/badge.svg?branch=develop)
[![codecov](https://codecov.io/gh/SYM01/htmlsanitizer/branch/develop/graph/badge.svg)](https://codecov.io/gh/SYM01/htmlsanitizer)


htmlsanitizer is a super fast, whitelist-based HTML sanitizer written in Golang. A built-in, secure-by-default whitelist helps you filter out any dangerous HTML content.

Why use htmlsanitizer?

- [x] Fast, a Finite State Machine was implemented internally, making the time complexity always O(n).
- [x] Highly customizable, allows you to modify the whitelist, or simply disable all HTML tags.
- [x] Dependency free.


## Install

```bash
go get -u github.com/sym01/htmlsanitizer
```


## Getting Started

### Use the secure-by-default whitelist

Simply use the secure-by-default whitelist to sanitize untrusted HTML.

```golang
sanitizedHTML, err := htmlsanitizer.SanitizeString(rawHTML)
```


### Disable the `id` attribute globally

By default, htmlsanitizer allows the `id` attribute globally. If we do NOT allow the `id` attribute, we can simply override the `GlobalAttr`.

```golang
s := htmlsanitizer.NewHTMLSanitizer()
s.GlobalAttr = []string{"class"}

sanitizedHTML, err := s.SanitizeString(rawHTML)
```

### Disable or add HTML tag

```golang
s := htmlsanitizer.NewHTMLSanitizer()
// remove <a> tag
s.RemoveTag("a")

// add a custom tag named my-tag, which allows my-attr attribute
customTag := &htmlsanitizer.Tag{
    Name: "my-tag",
    Attr: []string{"my-attr"},
}
s.Whitelist.Tags = append(s.Whitelist.Tags, customTag)

sanitizedHTML, err := s.SanitizeString(rawHTML)
```

### Disable all HTML tags

You can also use htmlsanitizer to remove all HTML tags.

```golang
s := htmlsanitizer.NewHTMLSanitizer()
// just set Whitelist to nil to disable all tags
s.Whitelist = nil

sanitizedHTML, err := s.SanitizeString(rawHTML)
```