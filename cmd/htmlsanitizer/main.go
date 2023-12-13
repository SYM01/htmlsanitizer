package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sym01/htmlsanitizer"
)

var (
	srcFilePath = flag.String("src", "", "could be either source file path, or the source URL")
)

func main() {
	flag.Parse()

	if len(*srcFilePath) == 0 {
		flag.CommandLine.Usage()
		return
	}

	var src io.ReadCloser
	switch {
	case strings.HasPrefix(*srcFilePath, "http://"), strings.HasPrefix(*srcFilePath, "https://"):
		resp, err := http.Get(*srcFilePath)
		if err != nil {
			log.Fatalf("unable to fetch remote content: %s", err)
		}
		src = resp.Body
	default:
		file, err := os.OpenFile(*srcFilePath, os.O_RDONLY, 0755)
		if err != nil {
			log.Fatalf("unable to open src file: %s", err)
		}
		src = file
	}

	defer src.Close()

	san := htmlsanitizer.NewHTMLSanitizer()
	writer := san.NewWriter(os.Stdout)
	if _, err := io.Copy(writer, src); err != nil {
		log.Printf("unable to sanitize HTML content: %s", err)
	}
}
