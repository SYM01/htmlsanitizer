package htmlsanitizer

import (
	"bytes"
	"html"
	"io"
	"strings"
	"unicode"
)

type state int

const (
	sNORMAL state = iota
	sLTSIGN
	sTAGNAME
	sTAGEND
	sATTRGAP
	sATTRNAME
	sEQUALSIGN
	sATTRSPACE
	sATTRVAL
	sVALSPACE
	sATTRQVAL
	sETAGSTART
	sETAGNAME
	sERRTAG
	sETAGATTR
	sETAGEND
)

func legalKeywordByte(b byte) bool {
	switch {
	case '0' <= b && b <= '9':
	case 'a' <= b && b <= 'z':
	case 'A' <= b && b <= 'Z':
	case b == '-':

	default:
		return false
	}

	return true
}

type writer struct {
	*HTMLSanitizer
	w     io.Writer
	state state

	// input data
	data []byte
	off  int

	// tmp data
	tagName    []byte
	tag        *Tag
	nonHTMLTag *Tag
	attr       []byte
	val        []byte
	quote      byte
	lastByte   byte // last byte for ATTRGAP

	// buf for write
	buf []byte
}

func (w *writer) flush() (n int, err error) {
	if len(w.buf) == 0 {
		return
	}

	n, err = w.w.Write(w.buf)

	// reset buf
	w.buf = w.buf[:0]
	return
}

func (w *writer) safeAppend(p []byte) {
	danger := false
	for _, b := range p {
		switch b {
		case '\'', '<', '>', '"':
			danger = true
		}
	}

	// fast path
	if !danger {
		w.buf = append(w.buf, p...)
		return
	}

	for _, b := range p {
		switch b {
		case '\'':
			w.buf = append(w.buf, `&#39;`...)
		case '<':
			w.buf = append(w.buf, `&lt;`...)
		case '>':
			w.buf = append(w.buf, `&gt;`...)
		case '"':
			w.buf = append(w.buf, `&#34;`...)
		default:
			w.buf = append(w.buf, b)
		}
	}
}

// write tag attribute and its sanitzed value if legal.
func (w *writer) safeAppendAttr() {
	if w.tag == nil {
		return
	}

	attrName := bytes.ToLower(w.attr)
	ok, urlAttr := w.tag.attrExists(attrName)
	if !ok && !w.attrExists(attrName) {
		return
	}

	attrVal := w.val
	if urlAttr {
		// unescape first
		rawURL := html.UnescapeString(string(w.val))
		newURL, ok := w.urlSanitizer(rawURL)
		if !ok {
			return
		}
		attrVal = []byte(newURL)
	}

	w.buf = append(w.buf, ' ')
	w.buf = append(w.buf, attrName...)
	w.buf = append(w.buf, `="`...)
	w.safeAppend(attrVal)
	w.buf = append(w.buf, '"')
}

// not an allowed tag, but it's a non-html tag.
// then we should ignore the contents inside
func (w *writer) shouldSwallowContent() bool {
	return w.nonHTMLTag != nil && w.tag == nil
}

func (w *writer) shouldKeepNonHTMLContent() bool {
	return w.nonHTMLTag != nil && w.tag != nil && w.nonHTMLTag.Name == w.tag.Name
}

func (w *writer) isEndTagOfNonHTMLElement(p []byte) bool {
	if w.nonHTMLTag == nil {
		return false
	}
	return w.nonHTMLTag.Name == strings.ToLower(string(p))
}

func (w *writer) Write(p []byte) (n int, err error) {
	// reset data
	w.data = p
	w.off = 0

	for err == nil && w.off < len(p) {
		switch w.state {
		case sNORMAL:
			err = w.sNORMAL()
		case sLTSIGN:
			err = w.sLTSIGN()
		case sTAGNAME:
			err = w.sTAGNAME()
		case sTAGEND:
			err = w.sTAGEND()
		case sATTRGAP:
			err = w.sATTRGAP()
		case sATTRNAME:
			err = w.sATTRNAME()
		case sEQUALSIGN:
			err = w.sEQUALSIGN()
		case sATTRSPACE:
			err = w.sATTRSPACE()
		case sATTRVAL:
			err = w.sATTRVAL()
		case sVALSPACE:
			err = w.sVALSPACE()
		case sATTRQVAL:
			err = w.sATTRQVAL()
		case sETAGSTART:
			err = w.sETAGSTART()
		case sETAGNAME:
			err = w.sETAGNAME()
		case sERRTAG:
			err = w.sERRTAG()
		case sETAGATTR:
			err = w.sETAGATTR()
		case sETAGEND:
			err = w.sETAGEND()
		default:
			panic("unknown state")
		}
	}

	n = w.off
	return
}

// done
func (w *writer) sNORMAL() (err error) {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; b {
		case '<':
			w.state = sLTSIGN
			w.off++

			_, err = w.flush()
			return
		case '>':
			if w.shouldSwallowContent() {
				continue
			}
			w.buf = append(w.buf, `&gt;`...)
		default:
			if w.shouldSwallowContent() {
				continue
			}
			w.buf = append(w.buf, b)
		}
	}

	_, err = w.flush()
	return err
}

// done
func (w *writer) sLTSIGN() error {
	switch b := w.data[w.off]; {
	case b == '/':
		w.off++

		w.state = sETAGSTART
		w.tagName = nil
		w.tag = nil

	case w.shouldKeepNonHTMLContent():
		w.buf = append(w.buf, `&lt;`...)
		w.state = sNORMAL

	case w.shouldSwallowContent():
		w.state = sNORMAL
		w.off++

	default:
		w.off++

		if legalKeywordByte(b) {
			w.state = sTAGNAME
			if len(w.tagName) > 0 {
				w.tagName = w.tagName[:0]
			}
			w.tagName = append(w.tagName, b)
			w.tag = nil
			return nil
		}

		w.state = sERRTAG
	}
	return nil
}

// done
func (w *writer) sTAGNAME() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; b {
		case '>':
			w.tag = w.FindTag(w.tagName)
			w.nonHTMLTag = w.checkNonHTMLTag(w.tagName)
			// no w.off++
			w.state = sTAGEND

			if w.tag == nil {
				return nil
			}

			w.buf = append(w.buf, '<')
			w.buf = append(w.buf, w.tag.Name...)
			return nil

		default:
			if legalKeywordByte(b) {
				w.tagName = append(w.tagName, b)
				continue
			}
			w.off++
			w.state = sATTRGAP
			w.lastByte = b

			w.tag = w.FindTag(w.tagName)
			w.nonHTMLTag = w.checkNonHTMLTag(w.tagName)
			if w.tag == nil {
				return nil
			}

			w.buf = append(w.buf, '<')
			w.buf = append(w.buf, w.tag.Name...)
			return nil
		}
	}

	return nil
}

// done
func (w *writer) sTAGEND() error {
	// eat '>'
	w.off++
	w.state = sNORMAL
	if w.tag == nil {
		// illegal tag, just reset the buf
		if len(w.buf) > 0 {
			w.buf = w.buf[:0]
		}
		w.lastByte = 0
		return nil
	}

	if w.lastByte == '/' {
		w.buf = append(w.buf, ` /`...)
		w.lastByte = 0
	}
	w.buf = append(w.buf, '>')
	_, err := w.flush()
	return err
}

// done
func (w *writer) sATTRGAP() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '>':
			// no w.off++
			w.state = sTAGEND
			return nil

		case legalKeywordByte(b):
			w.off++

			// reset
			if len(w.attr) > 0 {
				w.attr = w.attr[:0]
			}

			w.attr = append(w.attr, b)
			w.state = sATTRNAME
			return nil

		default:
			w.lastByte = b
		}
	}
	return nil
}

func (w *writer) sATTRNAME() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '=':
			w.off++
			w.state = sEQUALSIGN
			return nil
		case unicode.IsSpace(rune(b)):
			w.off++
			w.state = sATTRSPACE
			return nil
		case legalKeywordByte(b):
			w.attr = append(w.attr, b)
			continue
		default:
			// no w.off++
			w.lastByte = 0
			w.state = sATTRGAP

			if w.tag == nil {
				return nil
			}

			// name only attribute for HTML5
			attrName := bytes.ToLower(w.attr)
			if ok, _ := w.tag.attrExists(attrName); ok || w.attrExists(attrName) {
				w.buf = append(w.buf, ' ')
				w.buf = append(w.buf, attrName...)
			}
			return nil
		}
	}
	return nil
}

// done
func (w *writer) sEQUALSIGN() error {
	// reset
	if len(w.val) > 0 {
		w.val = w.val[:0]
	}

	switch b := w.data[w.off]; {
	case b == '>':
		// no w.off++
		w.safeAppendAttr()
		w.state = sTAGEND
	case b == '\'' || b == '"':
		w.off++
		w.quote = b
		w.state = sATTRQVAL
	case unicode.IsSpace(rune(b)):
		w.off++
		w.state = sVALSPACE
	default:
		w.off++
		w.state = sATTRVAL
		w.val = append(w.val, b)
	}

	return nil
}

// done
func (w *writer) sATTRSPACE() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '=':
			w.off++
			w.state = sEQUALSIGN
			return nil
		case unicode.IsSpace(rune(b)):
			continue
		case legalKeywordByte(b), b == '>':
			if w.tag != nil {
				attrName := bytes.ToLower(w.attr)
				ok, urlAttr := w.tag.attrExists(attrName)
				if (ok && !urlAttr) || w.attrExists(attrName) {
					w.buf = append(w.buf, ' ')
					w.buf = append(w.buf, attrName...)
				}
			}

			if b == '>' {
				// no w.off++
				w.state = sTAGEND
				return nil
			}

			w.off++
			if len(w.attr) > 0 {
				w.attr = w.attr[:0]
			}
			w.attr = append(w.attr, b)
			w.state = sATTRNAME
			return nil
		default:
			if w.tag != nil {
				attrName := bytes.ToLower(w.attr)
				ok, urlAttr := w.tag.attrExists(attrName)
				if (ok && !urlAttr) || w.attrExists(attrName) {
					w.buf = append(w.buf, ' ')
					w.buf = append(w.buf, attrName...)
				}
			}

			w.off++
			w.lastByte = b
			w.state = sATTRGAP
			return nil
		}
	}

	return nil
}

// done
func (w *writer) sATTRVAL() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '>':
			// no w.off++
			w.state = sTAGEND
			w.safeAppendAttr()
			return nil
		case unicode.IsSpace(rune(b)):
			w.off++
			w.lastByte = 0
			w.state = sATTRGAP
			w.safeAppendAttr()
			return nil
		default:
			w.val = append(w.val, b)
		}
	}
	return nil
}

// done
func (w *writer) sVALSPACE() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '>':
			// no w.off++
			w.safeAppendAttr()
			w.state = sTAGEND
			return nil
		case b == '\'' || b == '"':
			w.off++
			w.quote = b
			w.state = sATTRQVAL
			return nil
		case unicode.IsSpace(rune(b)):
			continue
		default:
			w.off++
			w.state = sATTRVAL
			// reset
			if len(w.val) > 0 {
				w.val = w.val[:0]
			}
			w.val = append(w.val, b)
			return nil
		}
	}
	return nil
}

// done
func (w *writer) sATTRQVAL() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; b {
		case w.quote:
			w.off++
			w.safeAppendAttr()
			w.lastByte = 0
			w.state = sATTRGAP
			return nil
		default:
			w.val = append(w.val, b)
		}
	}
	return nil
}

// done
func (w *writer) sETAGSTART() error {
	switch b := w.data[w.off]; {
	case legalKeywordByte(b):
		w.off++
		// reset
		if len(w.tagName) > 0 {
			w.tagName = w.tagName[:0]
		}
		w.tagName = append(w.tagName, b)
		w.state = sETAGNAME
	default:
		// no w.off++
		w.state = sERRTAG
	}

	return nil
}

// done
func (w *writer) sETAGNAME() error {
	for ; w.off < len(w.data); w.off++ {
		switch b := w.data[w.off]; {
		case b == '>':
			if w.isEndTagOfNonHTMLElement(w.tagName) {
				w.nonHTMLTag = nil
			}

			w.tag = w.FindTag(w.tagName)
			if w.tag == nil {
				// no w.off++
				w.state = sERRTAG
				return nil
			}

			// no w.off++
			w.state = sETAGEND
			w.buf = append(w.buf, `</`...)
			w.buf = append(w.buf, w.tag.Name...)
			return nil

		case legalKeywordByte(b):
			w.tagName = append(w.tagName, b)
			continue

		default:
			if w.isEndTagOfNonHTMLElement(w.tagName) {
				w.nonHTMLTag = nil
			}

			w.tag = w.FindTag(w.tagName)
			if w.tag == nil {
				// no w.off++
				w.state = sERRTAG
				return nil
			}

			w.off++
			w.state = sETAGATTR
			w.buf = append(w.buf, `</`...)
			w.buf = append(w.buf, w.tag.Name...)
			return nil
		}
	}

	return nil
}

// done
func (w *writer) sERRTAG() error {
	for ; w.off < len(w.data); w.off++ {
		if w.data[w.off] == '>' {
			w.off++
			w.state = sNORMAL
			// reset
			if len(w.buf) > 0 {
				w.buf = w.buf[:0]
			}
			return nil
		}
	}

	return nil
}

// done
func (w *writer) sETAGATTR() error {
	for ; w.off < len(w.data); w.off++ {
		if w.data[w.off] == '>' {
			// no w.off++
			w.state = sETAGEND
			return nil
		}
	}

	return nil
}

// done
func (w *writer) sETAGEND() error {
	// eat '>'
	w.off++

	w.state = sNORMAL
	w.buf = append(w.buf, '>')
	_, err := w.flush()
	return err
}
