package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/stretchr/testify/assert"
)

const (
	logoPath              = "../../test_assets/logo.jpg"
	templateTitlePagePath = "../../templates/title-page.html"
	templateTocPagePath   = "../../templates/toc-page.html"
	templateFacePagePath  = "../../face/index.html"
)

func TestPDFHandler(t *testing.T) {
	ts1 := httptest.NewServer(utils.WriteHTML(`
	<head>
		<title>Test PDFhandler 1</title>
	</head>
	<body>
		<h1>Test PDFhandler 1</h1>
		<p>Text 1</p>
	</body>
	`))
	defer ts1.Close()

	ts2 := httptest.NewServer(utils.WriteHTML(`
	<head>
		<title>Test PDFhandler 2</title>
	</head>
	<body>
		<h1>Test PDFhandler 2</h1>
		<p>Text 2</p>
		<h2>Header 1</h2>
		<h3 id="header-1-3">Header 1 level 3</h3>
		<h4 id="header-1-4">Header 2 level 4</h4>
		<h4>Header 3 level 4</h4>
		<h4>Header 4 level 4</h4>
		<h5>Header 5 level 5</h5>
		<h3>Header 6 level 3</h3>
		<h2>Header 2</h2>
	</body>
	`))
	defer ts2.Close()

	ts3 := httptest.NewServer(utils.WriteHTML(`
	<head>
		<title>Test PDFhandler 3</title>
	</head>
	<body>
		<h1>Test PDFhandler 3</h1>
		<p>Text 3</p>
	</body>
	`))
	defer ts3.Close()

	image, err := os.ReadFile(logoPath)
	assert.Nil(t, err)
	logo := httptest.NewServer(utils.WriteImage(image))
	defer ts3.Close()

	data := []byte(
		fmt.Sprintf(
			`{"title": "test PDFhandler","sub-title": "test sub-title", "info":"test info",
			"urls":["%s", "%s", "%s"], "landscape": "0", "selector": "body", "title-selector": "title",
			"title-page": true, "toc-page": true, "use-cache": false}`,
			ts1.URL, ts2.URL, ts3.URL))
	r := httptest.NewRequest(http.MethodGet, "/pdf", bytes.NewBuffer(data))
	w := httptest.NewRecorder()
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	cfg := config.Config{
		Env: "local",
	}
	cfg.LogoPath = logo.URL
	cfg.TocPage = templateTocPagePath
	cfg.TitlePage = templateTitlePagePath
	cfg.Face.FacePage = templateFacePagePath
	cfg.UrlsLimit = 30
	cfg.Limit = 1
	cfg.ChromeTimeout = 30 * time.Second
	cfg.CreateBrowser()

	err = cfg.InitTemplates()
	assert.Nil(t, err)

	sem := utils.Semaphore{
		C: make(chan struct{}, cfg.Limit),
	}

	handler := PDFHandler(log, &cfg, &sem)
	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	b := w.Body.Bytes()
	err = utils.SaveFile("../../tests/test-pdf-handler1.pdf", b)
	assert.Nil(t, err)

	assert.LessOrEqual(t, 4, len(w.Body.String())/10000)
}

func TestPDFHandlerRu(t *testing.T) {
	var err error

	ts1 := httptest.NewServer(utils.WriteHTML(`
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Тест PDFhandlerRu</title>
	</head>
	<body>
		<h1>Тест PDFhandler</h1>
		<p>Текст</p>
	</body>
	</html>
	`))
	defer ts1.Close()

	ts2 := httptest.NewServer(utils.WriteHTML(`
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Тест PDFhandler 2</title>
	</head>
	<body>
		<h1>Тест PDFhandler 2</h1>
		<p>Текст 2</p>
		<h2>Заголовок 1</h2>
		<h3 id="header-1-3">Заголовок 1 уровень 3</h3>
		<h4 id="header-1-4">Заголовок 2 уровень 4</h4>
		<h4>Заголовок 3 уровень 4</h4>
		<h4>Заголовок 4 уровень 4</h4>
		<h5>Заголовок 5 уровень 5</h5>
		<h3>Заголовок 6 уровень 3</h3>
		<h2>Заголовок 2</h2>
	</body>
	</html>
	`))
	defer ts2.Close()

	ts3 := httptest.NewServer(utils.WriteHTML(`
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Тест PDFhandler 3</title>
	</head>
	<body>
		<h1>Тест PDFhandler 3</h1>
		<p>Текст 3</p>
	</body>
	</html>
	`))
	defer ts3.Close()

	image, err := os.ReadFile(logoPath)
	assert.Nil(t, err)
	logo := httptest.NewServer(utils.WriteImage(image))
	defer ts3.Close()

	data := []byte(
		fmt.Sprintf(
			`{"title": "test PDFhandler","sub-title": "test sub-title", "info":"test info",
			"urls":["%s", "%s", "%s"], "landscape": "0", "selector": "body", "title-selector": "title",
			"title-page": true, "toc-page": true, "use-cache": false}`,
			ts1.URL, ts2.URL, ts3.URL))
	r := httptest.NewRequest(http.MethodGet, "/pdf", bytes.NewBuffer(data))
	w := httptest.NewRecorder()
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	cfg := config.Config{
		Env: "local",
	}
	cfg.LogoPath = logo.URL
	cfg.TocPage = templateTocPagePath
	cfg.TitlePage = templateTitlePagePath
	cfg.Face.FacePage = templateFacePagePath
	cfg.UrlsLimit = 30
	cfg.Limit = 1
	cfg.ChromeTimeout = 30 * time.Second
	cfg.CreateBrowser()

	err = cfg.InitTemplates()
	assert.Nil(t, err)

	sem := utils.Semaphore{
		C: make(chan struct{}, cfg.Limit),
	}

	handler := PDFHandler(log, &cfg, &sem)
	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	b := w.Body.Bytes()
	err = utils.SaveFile("../../tests/test-pdf-handler-ru.pdf", b)
	assert.Nil(t, err)

	assert.LessOrEqual(t, 4, len(w.Body.String())/10000)
}

func TestDOCXHandler(t *testing.T) {
	ts1 := httptest.NewServer(utils.WriteHTML(`
	<head>
		<title>Test DOCXhandler 1</title>
	</head>
	<body>
		<div>Text 1</div>
	</body>
	`))
	defer ts1.Close()
	ts2 := httptest.NewServer(utils.WriteHTML(`
<head>
	<title>Test DOCXhandler 2</title>
</head>
<body>
	<div>Text 2</div>
</body>
	`))
	defer ts2.Close()
	ts3 := httptest.NewServer(utils.WriteHTML(`
<head>
	<title>Test DOCXhandler 3</title>
</head>
<body>
	<div>Text 3</div>
</body>
	`))
	defer ts3.Close()

	data := []byte(fmt.Sprintf(`{"urls":["%s", "%s", "%s"], "landscape": "0"}`, ts1.URL, ts2.URL, ts3.URL))
	r := httptest.NewRequest(http.MethodGet, "/pdf", bytes.NewBuffer(data))
	w := httptest.NewRecorder()
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{
		Env:       "local",
		Debug:     true,
		HTTPSOnly: false,
		UrlsLimit: 30,
	}
	handler := DOCXHandler(log, &cfg)
	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	b := w.Body.Bytes()
	err := utils.SaveFile("../../tests/test-docx-handler1.docx", b)
	assert.Nil(t, err)

	assert.Equal(t, 85, len(w.Body.String())/100)
}

func TestFaceHandler(t *testing.T) {
	validData := []byte("{}")
	r := httptest.NewRequest(http.MethodGet, "/face", bytes.NewBuffer(validData))
	w := httptest.NewRecorder()
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{
		Env: "local",
	}
	cfg.TocPage = templateTitlePagePath
	cfg.TitlePage = templateTitlePagePath
	cfg.Face.FacePage = templateFacePagePath

	err := cfg.InitTemplates()
	assert.Nil(t, err)

	handler := FaceHandler(log, &cfg)
	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var b bytes.Buffer
	tmpl, err := template.ParseFiles(templateFacePagePath)
	assert.Nil(t, err)
	err = tmpl.Execute(&b, cfg.Face.Prefix)
	assert.Nil(t, err)

	expected := b.String()

	assert.Equal(t, expected, w.Body.String())
}
