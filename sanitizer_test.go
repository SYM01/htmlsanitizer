package htmlsanitizer_test

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/sym01/htmlsanitizer"
)

func ExampleNewWriter() {
	// demo data
	data := strings.Repeat(`abc-->
<a href="javascript:alert(1)">link1</a>
<a href=http://example.com>link2<script>xxx</script></a>
<!--`, 1024)
	expected := "abc--&gt;" + strings.Repeat(`
<a>link1</a>
<a href="http://example.com">link2</a>
`, 1024)

	// underlying writer for demo
	o := new(bytes.Buffer)

	// source reader for demo
	r := bytes.NewBufferString(data)

	sanitizedWriter := htmlsanitizer.NewWriter(o)
	_, _ = io.Copy(sanitizedWriter, r)

	// check the result, for demo only
	fmt.Print(o.String() == expected)
	// Output:
	// true
}

func ExampleHTMLSanitizer_keepStyleSheet() {
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	sanitizer.AllowList.Tags = append(sanitizer.AllowList.Tags,
		&htmlsanitizer.Tag{Name: "style"},
		&htmlsanitizer.Tag{Name: "head"},
		&htmlsanitizer.Tag{Name: "body"},
		&htmlsanitizer.Tag{Name: "html"},
	)

	data := `<!doctype html>
<html>
<head>
	<style type="text/css">
	body {
		background-color: #f0f0f2;
		margin: 0;
		padding: 0;
		bad-attr: <body></body>;
		bad-attr: <body></body >;
		bad-attr: <body></ body>;
		font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
	}
	</style>
</head>
<body>
	<div>
	<h1>Example Domain</h1>
	<p><a href="https://www.iana.org/domains/example">More information...</a></p>
	</div>
</body>
</html>`
	output, _ := sanitizer.SanitizeString(data)
	fmt.Print(output)
	// Output:
	//
	// <html>
	// <head>
	// 	<style>
	//	body {
	//		background-color: #f0f0f2;
	//		margin: 0;
	//		padding: 0;
	// 		bad-attr: &lt;body&gt;&lt;/body&gt;;
	// 		bad-attr: &lt;body&gt;&lt;/body &gt;;
	// 		bad-attr: &lt;body&gt;&lt;/ body&gt;;
	//		font-family: -apple-system, system-ui, BlinkMacSystemFont, "Segoe UI", "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
	//	}
	// 	</style>
	// </head>
	// <body>
	// 	<div>
	// 	<h1>Example Domain</h1>
	// 	<p><a href="https://www.iana.org/domains/example">More information...</a></p>
	// 	</div>
	// </body>
	// </html>
}

func ExampleHTMLSanitizer_noTagsAllowed() {
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	// just set AllowList to nil to disable all tags
	sanitizer.AllowList = nil

	// of course nothing will happen here
	sanitizer.RemoveTag("a")

	data := `
<a href="http://others.com">Link</a>
<a href="https://example.com/xxx">Link with example.com</a>
	`
	output, _ := sanitizer.SanitizeString(data)
	fmt.Print(output)
	// Output:
	//
	// Link
	// Link with example.com
}

func ExampleHTMLSanitizer_onlyAllowHrefTag() {
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	sanitizer.AllowList.Tags = []*htmlsanitizer.Tag{
		{"a", nil, []string{"href"}},
	}

	data := `
<details/open/ontoggle=alert(1)></details>
<a href="http://others.com" target="_blank">Link</a>
<a href="https://example.com/xxx">Link with example.com</a>
	`
	output, _ := sanitizer.SanitizeString(data)
	fmt.Print(output)
	// Output:
	//
	// <a href="http://others.com">Link</a>
	// <a href="https://example.com/xxx">Link with example.com</a>
}

func ExampleHTMLSanitizer_customURLSanitizer() {
	// only links with domain name example.com are allowed.
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	sanitizer.URLSanitizer = func(rawURL string) (newURL string, ok bool) {
		newURL, ok = htmlsanitizer.DefaultURLSanitizer(rawURL)
		if !ok {
			return
		}

		u, err := url.Parse(newURL)
		if err != nil {
			ok = false
			return
		}

		if u.Host == "example.com" {
			ok = true
			return
		}
		ok = false
		return
	}

	data := `
<a href="http://others.com">Link</a>
<a href="https://example.com/xxx">Link with example.com</a>
	`
	output, _ := sanitizer.SanitizeString(data)
	fmt.Print(output)
	// Output:
	//
	// <a>Link</a>
	// <a href="https://example.com/xxx">Link with example.com</a>
}

func TestSanitize(t *testing.T) {
	data := []byte(`<a class=x id= 123 href="javascript:alert(1)">demo</a>`)
	expected := []byte(`<a class="x" id="123">demo</a>`)
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

func TestSanitizeString(t *testing.T) {
	for _, item := range testCases {
		ret, err := htmlsanitizer.SanitizeString(item.in)

		if err != nil {
			t.Errorf("unable to SanitizeString(%#v) err: %s", item.in, err)
			break
		}

		if ret != item.out {
			t.Errorf("test failed for %#v, expect %#v, got %#v", item.in, item.out, ret)
			break
		}
	}
}

var testCases = []struct {
	in  string
	out string
}{
	{
		in:  `<a class="'<>" rel='aaa"'>test</a>`,
		out: "<a class=\"&#39;&lt;&gt;\" rel=\"aaa&#34;\">test</a>",
	},
	{
		in:  `<a href="ftp://example.com/xxx">test</a xxx>`,
		out: "<a>test</a>",
	},
	{
		in:  `<audio autoplay class=x>`,
		out: "<audio autoplay class=\"x\">",
	},
	{
		in:  `<audio autoplay />`,
		out: "<audio autoplay />",
	},
	{
		in:  `<audio autoplay/class="a">`,
		out: "<audio autoplay class=\"a\">",
	},
	{
		in:  `<span class=>`,
		out: "<span class=\"\">",
	},
	{
		in:  `<span`,
		out: "",
	},
	{
		in:  `</span`,
		out: "",
	},
	{
		in:  `<span class`,
		out: "",
	},
	{
		in:  `</span class`,
		out: "",
	},
	{
		in:  `<span class  `,
		out: "",
	},
	{
		in:  `<span class=  `,
		out: "",
	},
	{
		in:  `<span class="  `,
		out: "",
	},
	{
		in:  `<span class=  >`,
		out: "<span class=\"\">",
	},
	{
		in:  `<span class=  />`,
		out: "<span class=\"/\">",
	},
	{
		in:  `<span class  =   abc`,
		out: "",
	},
	{
		in:  `<span class  =   a>`,
		out: "<span class=\"a\">",
	},
	{
		in:  `<//>`,
		out: "",
	},
	{
		in:  `<//`,
		out: "",
	},

	// test cases from https://owasp.org/www-community/xss-filter-evasion-cheatsheet
	{
		in:  `<SCRIPT SRC=http://xss.rocks/xss.js></SCRIPT xxx>`,
		out: ``,
	},
	{
		in:  `javascript:/*--></title></style></textarea></script></xmp><svg/onload='+/"/+/onmouseover=1/+/[*/[]/+alert(1)//'>`,
		out: `javascript:/*--&gt;`,
	},
	{
		in:  `<IMG SRC="javascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=javascript:alert('XSS')>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=JaVaScRiPt:alert('XSS')>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=javascript:alert(&quot;XSS&quot;)>`,
		out: `<img>`,
	},
	{
		in:  "<IMG SRC=`javascript:alert(\"RSnake says, 'XSS'\")`>",
		out: `<img>`,
	},
	{
		in:  `\<a onmouseover="alert(document.cookie)"\>xxs link\</a\>`,
		out: `\<a>xxs link\</a>`,
	},
	{
		in:  `\<a onmouseover=alert(document.cookie)\>xxs link\</a\>`,
		out: `\<a>xxs link\</a>`,
	},
	{
		in:  `<IMG """><SCRIPT>alert("XSS")</SCRIPT>"\>`,
		out: `<img>"\&gt;`,
	},
	{
		in:  `<IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=#abc onmouseover="alert('xxs')">`,
		out: `<img src="#abc">`,
	},
	{
		in:  `<IMG SRC= onmouseover="alert('xxs')">`,
		out: `<img src="onmouseover=%22alert%28%27xxs%27%29%22">`,
	},
	{
		in:  `<IMG onmouseover="alert('xxs')">`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=/ onerror="alert(String.fromCharCode(88,83,83))"></img>`,
		out: "<img src=\"/\"></img>",
	},
	{
		in:  `<img src=x onerror="&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041">`,
		out: "<img src=\"x\">",
	},
	{
		in:  `<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;&#39;&#88;&#83;&#83;&#39;&#41;>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC="jav	ascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC="jav&#x09;ascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC="jav&#x0A;ascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  `<IMG SRC="jav&#x0D;ascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  "<IMG SRC=java\x00script:alert(\"XSS\")>",
		out: `<img>`,
	},
	{
		in:  `<IMG SRC=" &#14;  javascript:alert('XSS');">`,
		out: `<img>`,
	},
	{
		in:  `<SCRIPT/XSS SRC="http://xss.rocks/xss.js"></SCRIPT>`,
		out: ``,
	},
	{
		in:  "<BODY onload!#$%&()*~+-_.,:;?@[/|\\]^`=alert(\"XSS\")>",
		out: ``,
	},
	{
		in:  `<SCRIPT/SRC="http://xss.rocks/xss.js"></SCRIPT>`,
		out: ``,
	},
	{
		in:  `<<SCRIPT>alert("XSS");//\<</SCRIPT>`,
		out: `alert("XSS");//\`,
	},
	{
		in:  `<SCRIPT SRC=http://xss.rocks/xss.js?< B >`,
		out: ``,
	},
	{
		in:  "<IMG SRC=\"`<javascript:alert>`('XSS')\">",
		out: `<img>`,
	},
	{
		in:  `<iframe src=http://xss.rocks/scriptlet.html <`,
		out: ``,
	},
	{
		in:  `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');">`,
		out: ``,
	},
	{
		in:  `<IMG DYNSRC="javascript:alert('XSS')">`,
		out: `<img>`,
	},
	{
		in:  `<IMG LOWSRC="javascript:alert('XSS')">`,
		out: `<img>`,
	},
	{
		in:  `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}</STYLE><UL><LI>XSS</br>`,
		out: "<ul><li>XSS</br>",
	},
	{
		in:  `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}`,
		out: "",
	},
	{
		in:  `<svg/onload=alert('XSS')>`,
		out: ``,
	},
	{
		in:  `<BR SIZE="&{alert('XSS')}">`,
		out: `<br>`,
	},
	{
		in:  `<LINK REL="stylesheet" HREF="http://xss.rocks/xss.css">`,
		out: ``,
	},
	{
		in:  `<IMG STYLE="xss:expr/*XSS*/ession(alert('XSS'))">`,
		out: `<img>`,
	},
	{
		in:  `<XSS STYLE="xss:expression(alert('XSS'))">`,
		out: ``,
	},
	{
		in:  `¼script¾alert(¢XSS¢)¼/script¾`,
		out: `¼script¾alert(¢XSS¢)¼/script¾`,
	},
	{
		in:  `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`,
		out: ``,
	},
	{
		in:  `<IFRAME SRC=# onmouseover="alert(document.cookie)"></IFRAME>`,
		out: ``,
	},
	{
		in: `<!--[if gte IE 4]>
<SCRIPT>alert('XSS');</SCRIPT>
		<![endif]-->`,
		out: `

		`,
	},
	{
		in:  `<BASE HREF="javascript:alert('XSS');//">`,
		out: ``,
	},
	{
		in:  `<OBJECT TYPE="text/x-scriptlet" DATA="http://xss.rocks/scriptlet.html"></OBJECT>`,
		out: ``,
	},
	{
		in:  `<EMBED SRC="data:image/svg+xml;base64,PHN2ZyB4bWxuczpzdmc9Imh0dH A6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcv MjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hs aW5rIiB2ZXJzaW9uPSIxLjAiIHg9IjAiIHk9IjAiIHdpZHRoPSIxOTQiIGhlaWdodD0iMjAw IiBpZD0ieHNzIj48c2NyaXB0IHR5cGU9InRleHQvZWNtYXNjcmlwdCI+YWxlcnQoIlh TUyIpOzwvc2NyaXB0Pjwvc3ZnPg==" type="image/svg+xml" AllowScriptAccess="always"></EMBED>`,
		out: ``,
	},
	{
		in:  `<SCRIPT a=">" SRC="httx://xss.rocks/xss.js"></SCRIPT>`,
		out: ``,
	},
	{
		in:  `<SCRIPT =">" SRC="httx://xss.rocks/xss.js"></SCRIPT>`,
		out: "",
	},
	{
		in:  `<A HREF="http://66.102.7.147/">XSS</A>`,
		out: "<a href=\"http://66.102.7.147/\">XSS</a>",
	},
	{
		in:  `<A HREF="http://%77%77%77%2E%67%6F%6F%67%6C%65%2E%63%6F%6D">XSS</A>`,
		out: `<a>XSS</a>`,
	},
	{
		in: `<A HREF="h
		tt  p://6	6.000146.0x7.147/">XSS</A>`,
		out: `<a>XSS</a>`,
	},
	{
		in:  `<A HREF="javascript:document.location='http://www.google.com/'">XSS</A>`,
		out: `<a>XSS</a>`,
	},
	{
		in:  `<span>func <a class= "Documentation-source" href="https://cs.opensource.google/go/go/+/go1.21.5:src/os/path.go;l=66">RemoveAll</a> <a class="Documentation-idLink" href="#RemoveAll" aria-label="Go to RemoveAll">¶</a></span>`,
		out: `<span>func <a class="Documentation-source" href="https://cs.opensource.google/go/go/+/go1.21.5:src/os/path.go;l=66">RemoveAll</a> <a class="Documentation-idLink" href="#RemoveAll">¶</a></span>`,
	},
	{
		in: `
<Img src = x onerror = "javascript: window.onerror = alert; throw XSS">
<Video> <source onerror = "javascript: alert (XSS)">
<Input value = "XSS" type = text>
<applet code="javascript:confirm(document.cookie);">
<isindex x="javascript:" onmouseover="alert(XSS)">
"></SCRIPT>”>’><SCRIPT>alert(String.fromCharCode(88,83,83))</SCRIPT>
"><img src="x:x" onerror="alert(XSS)">
"><iframe src="javascript:alert(XSS)">
<object data="javascript:alert(XSS)" />
<isindex type=image src=1 onerror=alert(XSS)>
<img src=x:alert(alt) onerror=eval(src) alt=0>
<img  src="x:gif" onerror="window['al\u0065rt'](0)"></img>
<iframe/src="data:text/html,<svg onload=alert(1)>">
<meta content="&NewLine; 1 &NewLine;; JAVASCRIPT&colon; alert(1)" http-equiv="refresh"/>
<svg><script xlink:href=data&colon;,window.open('https://www.google.com/')></script
<meta http-equiv="refresh" content="0;url=javascript:confirm(1)">
<iframe src=javascript&colon;alert&lpar;document&period;location&rpar;>
<form><a href="javascript:\u0061lert(1)">X
</script><img/*%00/src="worksinchrome&colon;prompt(1)"/%00*/onerror='eval(src)'>
<style>//*{x:expression(alert(/xss/))}//<style></style>
On Mouse Over​
<img src="/" =_=" title="onerror='prompt(1)'">
<a aa aaa aaaa aaaaa aaaaaa aaaaaaa aaaaaaaa aaaaaaaaa aaaaaaaaaa href=j&#97v&#97script:&#97lert(1)>ClickMe
<script x> alert(1) </script 1=2
<form><button formaction=javascript&colon;alert(1)>CLICKME
<input/onmouseover="javaSCRIPT&colon;confirm&lpar;1&rpar;"
<iframe src="data:text/html,%3C%73%63%72%69%70%74%3E%61%6C%65%72%74%28%31%29%3C%2F%73%63%72%69%70%74%3E"></iframe>
<OBJECT CLASSID="clsid:333C7BC4-460F-11D0-BC04-0080C7055A83"><PARAM NAME="DataURL" VALUE="javascript:alert(1)"></OBJECT>
`,
		out: `
<img src="x">
<video> <source>



"&gt;”&gt;’&gt;
"&gt;<img>
"&gt;


<img alt="0">
<img></img>




<a>X
<img>

On Mouse Over​
<img src="/">
<a>ClickMe
CLICKME


`,
	},
}

func TestChunkedWrite(t *testing.T) {
	// Verify that splitting input across multiple Write calls produces the
	// same output as a single Write call.
	inputs := []string{
		`<a href="http://example.com" class="link">text</a>`,
		`<script>alert(1)</script><p>safe</p>`,
		`<div class="c"><img src="http://x.com/i.png" alt="pic" /></div>`,
		`<span style="color:red">hello</span>`,
		`<a href='http://example.com'>link</a>`,
	}

	for _, input := range inputs {
		expected, err := htmlsanitizer.SanitizeString(input)
		if err != nil {
			t.Fatalf("SanitizeString(%q) error: %v", input, err)
		}

		// Try splitting at every possible byte position
		data := []byte(input)
		for split := 1; split < len(data); split++ {
			var buf bytes.Buffer
			w := htmlsanitizer.NewWriter(&buf)

			_, err1 := w.Write(data[:split])
			_, err2 := w.Write(data[split:])
			if err1 != nil || err2 != nil {
				t.Fatalf("chunked Write(%q) split=%d errors: %v, %v", input, split, err1, err2)
			}

			got := buf.String()
			if got != expected {
				t.Errorf("chunked Write(%q) split=%d\nexpected: %q\ngot:      %q", input, split, expected, got)
			}
		}
	}
}

func TestCloneDeepCopy(t *testing.T) {
	original := htmlsanitizer.DefaultAllowList.Clone()

	// Record original state
	clone := original.Clone()
	origTagCount := len(original.Tags)
	origFirstTagAttrCount := len(original.Tags[0].Attr)

	// Mutate the clone's tag slice
	clone.Tags = append(clone.Tags, &htmlsanitizer.Tag{Name: "custom"})
	if len(original.Tags) != origTagCount {
		t.Errorf("appending to clone.Tags changed original: got %d, want %d", len(original.Tags), origTagCount)
	}

	// Mutate a Tag struct inside the clone
	clone.Tags[0].Attr = append(clone.Tags[0].Attr, "style")
	if len(original.Tags[0].Attr) != origFirstTagAttrCount {
		t.Errorf("mutating clone Tag.Attr changed original: got %d, want %d", len(original.Tags[0].Attr), origFirstTagAttrCount)
	}

	// Mutate GlobalAttr
	origGlobalCount := len(original.GlobalAttr)
	clone.GlobalAttr = append(clone.GlobalAttr, "data-x")
	if len(original.GlobalAttr) != origGlobalCount {
		t.Errorf("appending to clone.GlobalAttr changed original: got %d, want %d", len(original.GlobalAttr), origGlobalCount)
	}
}

var benchHTML = strings.Repeat(`<div class="container"><p>Hello <strong>world</strong>!</p><a href="http://example.com">link</a><img src="http://example.com/img.png" alt="test" /><script>alert('xss')</script></div>`, 100)
var benchPlain = strings.Repeat("Hello world, this is plain text with no HTML tags at all. ", 100)

func BenchmarkSanitize(b *testing.B) {
	data := []byte(benchHTML)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = htmlsanitizer.Sanitize(data)
	}
}

func BenchmarkSanitizePlainText(b *testing.B) {
	data := []byte(benchPlain)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = htmlsanitizer.Sanitize(data)
	}
}

func BenchmarkSanitizeString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = htmlsanitizer.SanitizeString(benchHTML)
	}
}

func BenchmarkSanitizeWriter(b *testing.B) {
	data := []byte(benchHTML)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := htmlsanitizer.NewWriter(io.Discard)
		_, _ = w.Write(data)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) {
	return 0, io.ErrShortWrite
}

func TestSanitizeWriteError(t *testing.T) {
	// Covers flush error propagation via NewWriter
	w := htmlsanitizer.NewWriter(failWriter{})
	_, err := w.Write([]byte("hello"))
	if err == nil {
		t.Error("expected error from failing writer")
	}

	// Covers Sanitize method error path (sanitizer.go:78)
	s := htmlsanitizer.NewHTMLSanitizer()
	s.URLSanitizer = func(rawURL string) (string, bool) {
		return rawURL, true
	}
	fw := failWriter{}
	sw := s.NewWriter(&fw)
	_, err = sw.Write([]byte(`<p>text</p>`))
	if err == nil {
		t.Error("expected error from Sanitize with failing writer")
	}
}

func TestEdgeCaseDFACoverage(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "invalid tag character triggers sERRTAG",
			in:   `<+invalid>text`,
			out:  `text`,
		},
		{
			name: "unquoted attr value after space in sVALSPACE",
			in:   `<a href= http://example.com>test</a>`,
			out:  `<a href="http://example.com">test</a>`,
		},
		{
			name: "consecutive closing tags reset tagName in sETAGSTART",
			in:   `<div></div></xyz>`,
			out:  `<div></div>`,
		},
		{
			name: "self-closing non-HTML tag",
			in:   `<script/>safe`,
			out:  `safe`,
		},
		{
			name: "comment-like tag",
			in:   `<!--comment-->visible`,
			out:  `visible`,
		},
		{
			name: "nil allowlist strips everything",
			in:   `<div class="x">hello</div>`,
			out:  `hello`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s *htmlsanitizer.HTMLSanitizer
			if tc.name == "nil allowlist strips everything" {
				s = htmlsanitizer.NewHTMLSanitizer()
				s.AllowList = nil
			}

			var got string
			var err error
			if s != nil {
				got, err = s.SanitizeString(tc.in)
			} else {
				got, err = htmlsanitizer.SanitizeString(tc.in)
			}
			if err != nil {
				t.Fatalf("SanitizeString(%q) error: %v", tc.in, err)
			}
			if got != tc.out {
				t.Errorf("SanitizeString(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.out)
			}
		})
	}
}
