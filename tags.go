package htmlsanitizer

// Tag with its attributes.
type Tag struct {
	// Name for current tag, must be lowercase.
	Name string

	// Attr specifies the allowed attributes for current tag,
	// must be lowercase.
	//
	// e.g. colspan, rowspan
	Attr []string

	// URLAttr specifies the allowed, URL-related attributes for current tag,
	// must be lowercase.
	//
	// e.g. src, href
	URLAttr []string
}

// attrExists checks whether attr exists. Case sensitive
func (t *Tag) attrExists(p []byte) (ok, urlAttr bool) {
	name := string(p)

	if t == nil {
		return
	}

	for _, attr := range t.URLAttr {
		if attr == name {
			ok, urlAttr = true, true
			return
		}
	}

	for _, attr := range t.Attr {
		if attr == name {
			ok = true
			return
		}
	}

	return
}

// AllowList specifies all the allowed HTML tags and its attributes for
// the sanitizer.
type AllowList struct {
	// Tags specifies all the allow tags.
	Tags []*Tag

	// GlobalAttr specifies the allowed attributes for all the tag.
	// It's very useful for some common attributes, such as `class`, `id`.
	// For security reasons, it's not recommended to set a global attr for
	// any URL-related attribute.
	GlobalAttr []string

	// NonHTMLTags defines a set of special tags, such as <script> and <style>.
	// The content of these kind of tags is actually not a real HTML content.
	// So we should treat it as a single element, without any child elements.
	// TODO: rename this one
	NonHTMLTags []*Tag

	// internal lookup maps, built lazily
	tagMap        map[string]*Tag
	nonHTMLTagMap map[string]*Tag
}

func buildTagLookup(tags []*Tag) map[string]*Tag {
	m := make(map[string]*Tag, len(tags))
	for _, tag := range tags {
		m[tag.Name] = tag
	}
	return m
}

// attrExists checks whether global attr exists. Case sensitive
func (l *AllowList) attrExists(p []byte) bool {
	if l == nil {
		return false
	}

	name := string(p)
	for _, attr := range l.GlobalAttr {
		if attr == name {
			return true
		}
	}

	return false
}

// checkNonHTMLTag checks if the given tag name is a non-html tag,
// such as `script` and `style`. Return nil if it's not a non-html tag.
// The name parameter must already be lowercased.
func (l *AllowList) checkNonHTMLTag(name string) *Tag {
	if l == nil {
		return nil
	}

	if l.nonHTMLTagMap == nil {
		l.nonHTMLTagMap = buildTagLookup(l.NonHTMLTags)
	}
	return l.nonHTMLTagMap[name]
}

// RemoveTag removes all tags name `name`, must be lowercase
// It is not recommended to modify the default list directly, use .Clone() and
// then modify the new one instead.
func (l *AllowList) RemoveTag(name string) {
	if l == nil || l.Tags == nil {
		return
	}

	idx := -1
	for i := 0; i < len(l.Tags); i++ {
		if l.Tags[i].Name == name {
			idx = i
			break
		}
	}

	if idx >= 0 {
		l.Tags = append(l.Tags[:idx], l.Tags[idx+1:]...)
		l.RemoveTag(name)
	}

	// invalidate maps
	l.tagMap = nil
	l.nonHTMLTagMap = nil
}

// FindTag finds and returns tag by its name.
// The name parameter must already be lowercased.
func (l *AllowList) FindTag(name string) *Tag {
	if l == nil {
		return nil
	}

	if l.tagMap == nil {
		l.tagMap = buildTagLookup(l.Tags)
	}
	return l.tagMap[name]
}

func cloneTagSlice(tags []*Tag) []*Tag {
	out := make([]*Tag, len(tags))
	for i, t := range tags {
		copied := *t
		copied.Attr = append([]string(nil), t.Attr...)
		copied.URLAttr = append([]string(nil), t.URLAttr...)
		out[i] = &copied
	}
	return out
}

// Clone a new AllowList. Tags and NonHTMLTags are deep-copied so that
// mutating the clone does not affect the original.
func (l *AllowList) Clone() *AllowList {
	if l == nil {
		return l
	}

	newList := new(AllowList)
	newList.Tags = cloneTagSlice(l.Tags)
	newList.GlobalAttr = append(newList.GlobalAttr, l.GlobalAttr...)
	newList.NonHTMLTags = cloneTagSlice(l.NonHTMLTags)

	return newList
}

func asciiLowerByte(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b | 0x20
	}
	return b
}

// asciiLowerInPlace lowercases ASCII A-Z bytes in-place. Since HTML tag and
// attribute names are ASCII-only, this avoids allocating via bytes.ToLower.
func asciiLowerInPlace(p []byte) {
	for i, b := range p {
		p[i] = asciiLowerByte(b)
	}
}

// DefaultAllowList for HTML filter.
//
// The allowlist contains most tags listed in
// https://developer.mozilla.org/en-US/docs/Web/HTML/Element .
// It is not recommended to modify the default list directly, use .Clone() and
// then modify the new one instead.
var DefaultAllowList = &AllowList{
	Tags: []*Tag{
		{"address", []string{}, []string{}},
		{"article", []string{}, []string{}},
		{"aside", []string{}, []string{}},
		{"footer", []string{}, []string{}},
		{"header", []string{}, []string{}},
		{"h1", []string{}, []string{}},
		{"h2", []string{}, []string{}},
		{"h3", []string{}, []string{}},
		{"h4", []string{}, []string{}},
		{"h5", []string{}, []string{}},
		{"h6", []string{}, []string{}},
		{"hgroup", []string{}, []string{}},
		{"main", []string{}, []string{}},
		{"nav", []string{}, []string{}},
		{"section", []string{}, []string{}},
		{"blockquote", []string{}, []string{"cite"}},
		{"dd", []string{}, []string{}},
		{"div", []string{}, []string{}},
		{"dl", []string{}, []string{}},
		{"dt", []string{}, []string{}},
		{"figcaption", []string{}, []string{}},
		{"figure", []string{}, []string{}},
		{"hr", []string{}, []string{}},
		{"li", []string{}, []string{}},
		{"ol", []string{}, []string{}},
		{"p", []string{}, []string{}},
		{"pre", []string{}, []string{}},
		{"ul", []string{}, []string{}},
		{"a", []string{"rel", "target", "referrerpolicy"}, []string{"href"}},
		{"abbr", []string{"title"}, []string{}},
		{"b", []string{}, []string{}},
		{"bdi", []string{}, []string{}},
		{"bdo", []string{}, []string{}},
		{"br", []string{}, []string{}},
		{"cite", []string{}, []string{}},
		{"code", []string{}, []string{}},
		{"data", []string{"value"}, []string{}},
		{"em", []string{}, []string{}},
		{"i", []string{}, []string{}},
		{"kbd", []string{}, []string{}},
		{"mark", []string{}, []string{}},
		{"q", []string{}, []string{"cite"}},
		{"s", []string{}, []string{}},
		{"small", []string{}, []string{}},
		{"span", []string{}, []string{}},
		{"strong", []string{}, []string{}},
		{"sub", []string{}, []string{}},
		{"sup", []string{}, []string{}},
		{"time", []string{"datetime"}, []string{}},
		{"u", []string{}, []string{}},
		{"area", []string{"alt", "coords", "shape", "target", "rel", "referrerpolicy"}, []string{"href"}},
		{"audio", []string{"autoplay", "controls", "crossorigin", "duration", "loop", "muted", "preload"}, []string{"src"}},
		{"img", []string{"alt", "crossorigin", "height", "width", "loading", "referrerpolicy"}, []string{"src"}},
		{"map", []string{"name"}, []string{}},
		{"track", []string{"default", "kind", "label", "srclang"}, []string{"src"}},
		{"video", []string{"autoplay", "buffered", "controls", "crossorigin", "duration", "loop", "muted", "preload", "height", "width"}, []string{"src", "poster"}},
		// no embed
		// no iframe
		// no object
		// no param
		{"picture", []string{}, []string{}},
		{"source", []string{"type"}, []string{"src"}},
		// no canvas
		// no script
		{"del", []string{}, []string{}},
		{"ins", []string{}, []string{}},
		{"caption", []string{}, []string{}},
		{"col", []string{"span"}, []string{}},
		{"colgroup", []string{}, []string{}},
		{"table", []string{}, []string{}},
		{"tbody", []string{}, []string{}},
		{"td", []string{"colspan", "rowspan"}, []string{}},
		{"tfoot", []string{}, []string{}},
		{"th", []string{"colspan", "rowspan", "scope"}, []string{}},
		{"thead", []string{}, []string{}},
		{"tr", []string{}, []string{}},
		// no Forms
		{"details", []string{"open"}, []string{}},
		{"summary", []string{}, []string{}},
		// no Web Components
	},
	GlobalAttr: []string{
		"class",
		"id",
	},
	NonHTMLTags: []*Tag{
		{Name: "script"},
		{Name: "style"},
		{Name: "object"},
	},
}
