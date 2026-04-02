package core

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

const Selector = "body"

func TestMergePDF(t *testing.T) {
	var cfg config.Config
	var final []byte

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

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

	allPages := [][]byte{
		[]byte("<div>Text 1</div>"),
		[]byte("<div>Text 2</div>"),
		[]byte("<div>Text 3</div>"),
	}

	cfg.GrabScript = ""
	cfg.ChromeTimeout = 30 * time.Second
	cfg.Limit = 1

	cfg.CreateBrowser()

	sem := utils.Semaphore{
		C: make(chan struct{}, cfg.Limit),
	}

	err := MergeHTMLtoPDF(allPages, &final, &params, &cfg, log, &sem)
	assert.Nil(t, err)
	assert.LessOrEqual(t, 10, len(final)/1000)
	err = utils.SaveFile("../tests/merge-pdf-test1.pdf", final)
	assert.Nil(t, err)
}

func TestPDFPrinter(t *testing.T) {
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

	htmlBytes := [][]byte{
		[]byte(`
	<div>Text 1</div>
`),
		[]byte(`
<div>Text 2</div>
`),
		[]byte(`
<div>Text 3</div>
`),
	}

	allPages := bytes.Join(htmlBytes, []byte("\n<div style=\"page-break-before: always;\"></div>"))

	var res []byte
	err := chromedp.Run(taskCtx, PDFPrinter(ts.URL, allPages, &params, "", &res, nil))
	assert.Nil(t, err)
	err = utils.SaveFile("../tests/PDFPrinter-test1.pdf", res)
	assert.Nil(t, err)
	assert.LessOrEqual(t, 10, len(res)/1000)
	assert.Nil(t, err)
}
