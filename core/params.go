package core

import (
	"fmt"
	"log/slog"
	"net/http"
	u "net/url"

	"github.com/TOsmanov/go-pdf/internal/config"
)

type Params struct {
	Title           string   `json:"title,omitempty"`
	SubTitle        string   `json:"sub-title,omitempty" default:""`
	FooterTitle     string   `json:"info,omitempty" default:""`
	Logo            string   `json:"logo-path,omitempty"`
	Urls            []string `json:"urls"`
	Landscape       string   `json:"landscape,omitempty"`
	Selector        string   `json:"selector,omitempty" default:"body"`
	TitleSelector   string   `json:"title-selector,omitempty" default:"title"`
	EnableToCPage   bool     `json:"toc-page,omitempty" default:"false"`
	EnableTitlePage bool     `json:"title-page,omitempty" default:"false"`
	UseCache        bool     `json:"use-cache,omitempty"`
	ToC             []Page   // the structure contains parameters for creating a table of contents page
	FinalLandscape  bool
}

type Page struct {
	Index   int      `json:"index"`
	Title   string   `json:"title"`
	Anchor  string   `json:"anchor"`
	Headers []Header `json:"headers"`
}

type Header struct {
	Lvl    int    `json:"lvl"`    // the header level, for example: 3
	Text   string `json:"text"`   // title text, for example: "Built-in data types"
	Link   string `json:"link"`   // generated anchor link, for example: "gopdf-header-6h3"
	Index  string `json:"index"`  // sequence number, for example: "7.3.1"
	Class  string `json:"class"`  // class of header, for example: "lvl-3"
	Indent int    `json:"indent"` // margin of header, for example: -1
}

func (params *Params) Validation(log *slog.Logger, cfg *config.Config, mode string) error {
	// TODO: return error if the file is too large
	var validURLs []string
	statusThreshold := 400

	if cfg.Env != "prod" {
		statusThreshold = 403 // allow redirects if not prod env
	}

	if params.Landscape == "1" {
		params.FinalLandscape = true
	}
	if len(params.Selector) == 0 {
		if mode == "pdf" {
			params.Selector = cfg.PdfSelector
		} else if mode == "docx" {
			params.Selector = cfg.DocxSelector
		}
	}
	if len(params.TitleSelector) == 0 {
		params.TitleSelector = cfg.PdfTitleSelector
	}
	for _, url := range params.Urls {
		newURL, err := u.Parse(url)
		if err != nil || newURL.Scheme == "" || newURL.Host == "" {
			log.Debug("the invalid URL has been removed from the URL list",
				slog.Any("url", newURL))
		} else {
			url, msg, badURL := validationURL(newURL, cfg.TrustedHosts, cfg.HTTPSOnly)
			if msg != "" {
				log.Debug(msg,
					slog.Any("deleted url", badURL))
			}
			if url != "" {
				validURLs = append(validURLs, url)
			}
		}
	}
	params.Urls = validURLs
	if len(params.Urls) == 0 {
		return fmt.Errorf("URLs is missing")
	}
	if len(params.Urls) > cfg.UrlsLimit {
		return fmt.Errorf("the number of URLs is more than %d", cfg.UrlsLimit)
	}

	for _, url := range params.Urls {
		client := &http.Client{}
		testResp, err := client.Head(url) //nolint
		if err != nil {
			return fmt.Errorf("url verification error: %w", err)
		}
		defer testResp.Body.Close()

		log.Debug("check url status code",
			slog.String("URL", url),
			slog.Int("status code", testResp.StatusCode))
		if testResp.StatusCode > statusThreshold {
			return fmt.Errorf("url verification error: status code %d", testResp.StatusCode)
		}
	}
	if len(params.Logo) == 0 {
		params.Logo = cfg.LogoPath
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func validationURL(newURL *u.URL, trustedHosts []string, replaceProtocol bool) (string, string, string) {
	// check protocol
	if newURL.Scheme != "https" && replaceProtocol {
		newURL.Scheme = "https"
		return newURL.String(), "invalid URL scheme has been replaced with https", newURL.String()
	}

	// check host
	if len(trustedHosts) > 0 {
		if contains(trustedHosts, newURL.Host) {
			return newURL.String(), "", ""
		}
		return "", "the host is not in the trusted list, URL removed", newURL.String()
	}
	return newURL.String(), "", ""
}
