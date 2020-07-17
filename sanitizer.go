package htmlsanitizer

import (
	"bytes"
	"io"
	"net/url"
)

// DefaultURLSanitizer is a default and strict sanitizer.
// It only accepts
//  * URL with scheme http or https
//  * relative URL, such as abc, abc?xxx=1, abc#123
//  * absolute URL, such as /abc, /abc?xxx=1, /abc#123
func DefaultURLSanitizer(rawURL string) (sanitzed string, ok bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	if len(u.Opaque) > 0 {
		return
	}

	switch u.Scheme {
	case "http", "https", "":
	default:
		return
	}

	sanitzed = u.String()
	ok = true
	return
}

// HTMLSanitizer is a super fast HTML sanitizer for arbitrary HTML content.
// This is a allowlist-based santizer, of which the time complexity is O(n).
type HTMLSanitizer struct {
	*AllowList

	// URLSanitizer is a func used to sanitize all the URLAttr.
	// URLSanitizer returns a sanitzed URL and a bool var indicating
	// whether the current attribute is acceptable. If not acceptable,
	// the current attribute will be ignored.
	// If the func is nil, then DefaultURLSanitizer will be used.
	URLSanitizer func(rawURL string) (sanitzed string, ok bool)
}

// NewHTMLSanitizer creates a new HTMLSanitizer with the clone of
// the DefaultAllowList.
func NewHTMLSanitizer() *HTMLSanitizer {
	return &HTMLSanitizer{
		AllowList: DefaultAllowList.Clone(),
	}
}

func (f *HTMLSanitizer) urlSanitizer(rawURL string) (sanitzed string, ok bool) {
	if f.URLSanitizer != nil {
		return f.URLSanitizer(rawURL)
	}

	return DefaultURLSanitizer(rawURL)
}

// NewWriter returns a new Writer writing sanitized HTML content to w.
func (f *HTMLSanitizer) NewWriter(w io.Writer) io.Writer {
	return &writer{
		HTMLSanitizer: f,
		w:             w,
	}
}

// Sanitize the HTML data and return the sanitzed HTML.
func (f *HTMLSanitizer) Sanitize(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	if _, err := f.NewWriter(buf).Write(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SanitizeString sanitizes the HTML string and return the sanitzed HTML.
func (f *HTMLSanitizer) SanitizeString(data string) (string, error) {
	ret, err := f.Sanitize([]byte(data))
	var retStr string
	if ret != nil {
		retStr = string(ret)
	}

	return retStr, err
}

var defaultHTMLSanitizer = NewHTMLSanitizer()

// NewWriter returns a new Writer, with DefaultAllowList,
// writing sanitized HTML content to w.
func NewWriter(w io.Writer) io.Writer {
	return defaultHTMLSanitizer.NewWriter(w)
}

// Sanitize uses the DefaultAllowList to sanitize the HTML data.
func Sanitize(data []byte) ([]byte, error) {
	return defaultHTMLSanitizer.Sanitize(data)
}

// SanitizeString uses the DefaultAllowList to sanitize the HTML string.
func SanitizeString(data string) (string, error) {
	return defaultHTMLSanitizer.SanitizeString(data)
}
