package core

import (
	"log/slog"
	"os"
	"testing"

	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidationPdf(t *testing.T) {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{}
	cfg.PdfTitleSelector = "h1"
	cfg.UrlsLimit = 30
	cfg.TrustedHosts = []string{
		"ya.ru",
	}

	params := Params{
		Urls: []string{
			"not-valid url",
			"https://ya.ru",
			"",
			"tewfger",
		},
		Selector: Selector,
	}
	err := params.Validation(log, &cfg, "pdf")
	assert.Nil(t, err)

	expect := Params{
		Urls: []string{
			"https://ya.ru",
		},
		Selector:      Selector,
		TitleSelector: "h1",
	}
	assert.Equal(t, expect, params)
}

func TestValidationHost(t *testing.T) {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{}
	cfg.PdfTitleSelector = "h1"
	cfg.UrlsLimit = 30
	cfg.TrustedHosts = []string{
		"ya.ru",
	}

	params := Params{
		Urls: []string{
			"not-valid url",
			"https://ya.ru",
			"https://habr.com/badurl",
			"",
			"tewfger",
		},
		Selector: Selector,
	}
	err := params.Validation(log, &cfg, "pdf")
	assert.Nil(t, err)

	expect := Params{
		Urls: []string{
			"https://ya.ru",
		},
		Selector:      Selector,
		TitleSelector: "h1",
	}
	assert.Equal(t, expect, params)
}

func TestValidationProtocol(t *testing.T) {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{}
	cfg.PdfTitleSelector = "h1"
	cfg.UrlsLimit = 30
	cfg.HTTPSOnly = true

	params := Params{
		Urls: []string{
			"not-valid url",
			"http://ya.ru",
			"http://habr.com",
			"",
			"tewfger",
		},
		Selector: Selector,
	}
	err := params.Validation(log, &cfg, "pdf")
	assert.Nil(t, err)

	expect := Params{
		Urls: []string{
			"https://ya.ru",
			"https://habr.com",
		},
		Selector:      Selector,
		TitleSelector: "h1",
	}
	assert.Equal(t, expect, params)
}

func TestValidationDocx(t *testing.T) {
	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	cfg := config.Config{}
	cfg.PdfTitleSelector = "h1"
	cfg.UrlsLimit = 30
	cfg.TrustedHosts = []string{
		"ya.ru",
	}

	params := Params{
		Urls: []string{
			"not-valid url",
			"https://ya.ru",
			"",
			"tewfger",
		},
		Selector: Selector,
	}
	err := params.Validation(log, &cfg, "docx")
	assert.Nil(t, err)

	expect := Params{
		Urls: []string{
			"https://ya.ru",
		},
		Selector:      Selector,
		TitleSelector: "h1",
	}
	assert.Equal(t, expect, params)
}
