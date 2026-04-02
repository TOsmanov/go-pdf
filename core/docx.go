package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	u "net/url"
	"strings"

	"github.com/fumiama/go-docx"
	"golang.org/x/net/html"
)

func DOCXGrabber(params Params, res *[]byte, selector string) error {
	f := docx.New().WithDefaultTheme()
	for _, url := range params.Urls {
		b, err := GetHTML(url)
		if err != nil {
			return err
		}
		r := bytes.NewReader(b)
		parseHTML(r, f, selector)
	}
	var buf bytes.Buffer
	r := io.Writer(&buf)
	_, err := f.WriteTo(r)
	if err != nil {
		return err
	}
	*res = buf.Bytes()
	return nil
}

func GetHTML(url string) ([]byte, error) {
	r, err := http.Get(url) //nolint
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func parseHTML(r io.Reader, f *docx.Docx, selector string) {
	z := html.NewTokenizer(r)
	para := f.AddParagraph()
	href := ""
	skip := false
	start := false
	var text string
	for {
		tokenType := z.Next()
		token := z.Token()
		switch tokenType { //nolint
		case html.StartTagToken:
			if start {
				switch token.Data {
				case "a":
					href = findHref(token)
				case "p", "h1", "h2", "h3", "h4", "h5", "h6":
					para.AddText("\n")
				case "script":
					skip = true
				}
			} else {
				start = findSelector(token, selector)
			}
		case html.TextToken:
			text = fmt.Sprintf("%s ", strings.TrimSpace(token.Data))
			switch skip {
			case false && len(text) > 1:
				switch len(href) {
				default:
					_, err := u.Parse(href)
					if err == nil {
						para.AddText(text)
						para.AddText(href)
					}
					href = ""
				case 0:
					para.AddText(text)
				}
			case true:
				skip = false
			}
		}
		if tokenType == html.ErrorToken {
			break
		}
	}
}

func findHref(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

func findSelector(token html.Token, selector string) bool {
	for _, attr := range token.Attr {
		if attr.Key == "class" {
			for _, class := range strings.Split(attr.Val, " ") {
				if class == selector {
					return true
				}
			}
		}
	}
	return false
}
