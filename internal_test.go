package htmlsanitizer

import "testing"

func TestNilTagAttrExists(t *testing.T) {
	var tag *Tag
	ok, urlAttr := tag.attrExists([]byte("href"))
	if ok || urlAttr {
		t.Error("expected false for nil Tag.attrExists")
	}
}

func TestNilAllowListAttrExists(t *testing.T) {
	var al *AllowList
	if al.attrExists([]byte("class")) {
		t.Error("expected false for nil AllowList.attrExists")
	}
}

func TestNilAllowListClone(t *testing.T) {
	var al *AllowList
	if al.Clone() != nil {
		t.Error("expected nil for nil AllowList.Clone()")
	}
}

func TestNilAllowListCheckNonHTMLTag(t *testing.T) {
	var al *AllowList
	if al.checkNonHTMLTag("script") != nil {
		t.Error("expected nil for nil AllowList.checkNonHTMLTag")
	}
}

func TestNilAllowListFindTag(t *testing.T) {
	var al *AllowList
	if al.FindTag("div") != nil {
		t.Error("expected nil for nil AllowList.FindTag")
	}
}

func TestIsEndTagOfNonHTMLElement(t *testing.T) {
	// Length mismatch
	w := &writer{
		HTMLSanitizer: NewHTMLSanitizer(),
	}
	w.nonHTMLTag = &Tag{Name: "script"}
	if w.isEndTagOfNonHTMLElement([]byte("style")) {
		t.Error("expected false for different length tag name")
	}
	// Byte mismatch (same length)
	if w.isEndTagOfNonHTMLElement([]byte("scopet")) {
		t.Error("expected false for mismatched bytes with same length")
	}
	// Match with uppercase
	if !w.isEndTagOfNonHTMLElement([]byte("SCRIPT")) {
		t.Error("expected true for case-insensitive match")
	}
}
