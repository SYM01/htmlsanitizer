package htmlsanitizer

import (
	"fmt"
)

func ExampleAllowList_RemoveTag() {
	// sometimes we don't want user to pass HTML with <a> tag
	sanitizer := NewHTMLSanitizer()
	sanitizer.RemoveTag("a")

	data := `
<h1 ClaSs="h1">hello</h1>
<p>
	Hello, world<br>
	Welcome to use <a href="https://github.com/sym01/htmlsanitizer">htmlsanitizer</a>
</p>`
	output, _ := sanitizer.SanitizeString(data)
	fmt.Print(output)
	// Output:
	//
	// <h1 class="h1">hello</h1>
	// <p>
	// 	Hello, world<br>
	// 	Welcome to use htmlsanitizer
	// </p>
}
