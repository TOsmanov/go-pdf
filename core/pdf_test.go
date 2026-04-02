package core

import (
	"context"
	"log"
	"net/http/httptest"
	"testing"

	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestPageGrabber(t *testing.T) {
	taskCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ts := httptest.NewServer(utils.WriteHTML(`
<head>
	<title>Test PDF Grabber</title>
</head>
<body>
	<div>Text</div>
</body>
	`))
	defer ts.Close()

	params := Params{
		Urls:      []string{ts.URL},
		Landscape: "0",
	}
	params.Selector = Selector
	var res string
	err := chromedp.Run(taskCtx, PageGrabber(0, ts.URL, &params, "", &res, nil))
	assert.Nil(t, err)
	expected := `
	<div>Text</div>
`
	assert.Equal(t, expected, res)
	assert.Nil(t, err)
}

func TestPDFGrabber(t *testing.T) {
	taskCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ts := httptest.NewServer(utils.WriteHTML(`
<head>
	<title>Test PDF Grabber</title>
</head>
<body>
	<div>Text</div>
</body>
	`))
	defer ts.Close()

	params := Params{
		Urls:      []string{ts.URL},
		Landscape: "0",
	}
	var res []byte
	_ = chromedp.Run(taskCtx, PDFGrabber(0, ts.URL, &params, "", &res, nil))
	// assert.Nil(t, err)
	// assert.Equal(t, 7, len(res)/1000)
	err := utils.SaveFile("../tests/pdf-grabber-test.pdf", res)
	assert.Nil(t, err)
}
