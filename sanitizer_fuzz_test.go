//go:build go1.18
// +build go1.18

package htmlsanitizer_test

import (
	"bytes"
	"testing"

	"github.com/sym01/htmlsanitizer"
)

func FuzzSanitize(f *testing.F) {
	seeds := [][]byte{
		[]byte(`xxls<<< <xx />`),
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, a []byte) {
		data := append([]byte(`abc<script>`), a...)
		data = append(data, `abc</script>def<style>`...)
		data = append(data, a...)
		data = append(data, `</style ...>(g)`...)
		expected := []byte(`abcdef(g)`)
		ret, err := htmlsanitizer.Sanitize(data)
		if err != nil {
			t.Errorf("unable to Sanitize err: %s", err)
			return
		}
		if !bytes.Equal(ret, expected) {
			t.Errorf("test failed for %s, expect %s, got %s", data, expected, ret)
			return
		}
	})
}

func TestSanitize_tmp(t *testing.T) {
	a := []byte(`</A>`)
	data := append([]byte(`abc<script>`), a...)
	data = append(data, `abc</script>def<style>`...)
	data = append(data, a...)
	data = append(data, `</style ...>(g)`...)
	expected := []byte(`abcdef(g)`)
	ret, err := htmlsanitizer.Sanitize(data)
	if err != nil {
		t.Errorf("unable to Sanitize err: %s", err)
		return
	}
	if !bytes.Equal(ret, expected) {
		t.Errorf("test failed for %s, expect %s, got %s", data, expected, ret)
		return
	}
}
