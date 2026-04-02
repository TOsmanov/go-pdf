package core

import (
	"net/http/httptest"
	"testing"

	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/stretchr/testify/assert"
)

func TestDOCXGrabber1(t *testing.T) {
	htmlSource := `<!DOCTYPE html>
	<head>
		<title>Text for test title</title>
	</head>
	<body class="main">
		<p>
			<b>Test text</b>
			<a href="https://ya.ru">Test link</a>
		</p>
		<p>
			<b>Test text 2</b>
		</p>
	</body>
</html>`

	ts := httptest.NewServer(utils.WriteHTML(htmlSource))
	defer ts.Close()

	params := Params{
		Urls:      []string{ts.URL},
		Landscape: "0",
	}
	var res []byte
	err := DOCXGrabber(params, &res, "main")
	assert.Nil(t, err)
	assert.LessOrEqual(t, 80, len(res)/100)
	err = utils.SaveFile("../tests/docx-test1.docx", res)
	assert.Nil(t, err)
}

func TestDOCXGrabber2(t *testing.T) {
	htmlSource := `<!DOCTYPE html>
	<head>
		<title>Text for test title 2</title>
	</head>
	<body class="main">
		<p>
			<b>Test text 2</b>
			<a href="https://ya.ru">Test link</a>
		</p>
		<p>
			<b>Test text 2</b>
		</p>
	</body>
</html>`
	htmlSource1 := `<!DOCTYPE html>
	<head>
		<title>Text for test title 2</title>
	</head>
	<body class="main">
		<p>
			<a href="https://ya.ru">Test link</a>
		</p>
		<p>
			<b>Test text 2</b>
		</p>
	</body>
</html>`

	ts := httptest.NewServer(utils.WriteHTML(htmlSource))
	defer ts.Close()

	ts1 := httptest.NewServer(utils.WriteHTML(htmlSource1))
	defer ts1.Close()

	params := Params{
		Urls:      []string{ts.URL, ts1.URL},
		Landscape: "0",
	}
	var res []byte
	err := DOCXGrabber(params, &res, "main")
	assert.Nil(t, err)
	assert.LessOrEqual(t, 80, len(res)/100)
	err = utils.SaveFile("../tests/docx-test2.docx", res)
	assert.Nil(t, err)
}
