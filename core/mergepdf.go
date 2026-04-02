package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Num struct {
	Link  string
	Index string
}

func MergeHTMLtoPDF(htmlBytes [][]byte, final *[]byte,
	params *Params, cfg *config.Config,
	log *slog.Logger, sem *utils.Semaphore,
) error {
	sem.Acquire()
	defer sem.Release()

	var res []byte
	var toc []byte
	var err error

	log.Debug("merge html pages")
	html := bytes.Join(htmlBytes, []byte("\n<div style=\"page-break-before: always;\"></div>"))

	for i, pageURL := range params.Urls {
		href := []byte(fmt.Sprintf("href=\"%s\"", pageURL))
		html = bytes.ReplaceAll(html, href, []byte(fmt.Sprintf("href=\"#gopdf-page-header-%d\"", i)))
		hrefAnchor := []byte(fmt.Sprintf("href=\"%s#", pageURL))
		html = bytes.ReplaceAll(html, hrefAnchor, []byte("href=\"#"))
	}

	if cfg.Debug {
		// dump html-data
		err := os.WriteFile("./data.html", html, 0o666) //nolint
		if err != nil {
			return err
		}
		log.Debug("dump data.html")
	}

	timeout := make(chan bool, 1)
	done := make(chan bool, 1)

	go func() {
		time.Sleep(cfg.ChromeTimeout)
		timeout <- true
	}()

	if len(params.ToC) > 0 {
		if !params.EnableToCPage {
			PrepareToC(params)
		}

		toc, err = prepareNumeration(params.ToC)
		if err != nil {
			return err
		}
	}

	if err := chromedp.Run(*cfg.Browser.Ctx, PDFPrinter(params.Urls[0],
		html,
		params,
		cfg.MergeScript,
		&res, toc)); err != nil {
		return err
	}

	done <- true

	select {
	case <-done:
		close(done)
	case <-timeout:
		return fmt.Errorf("the timeout has expired while performing the merge")
	}

	*final = res

	return nil
}

func PDFPrinter(url string, allPages []byte, params *Params,
	customScript string, res *[]byte, toc []byte,
) chromedp.Tasks {
	/*
		Creates a chrome task to export a all pages to pdf.
		url - the URL that will be opened in chrome
		allPages - the html code that will be inserted on the page
		params - the structure of the request parameters
		customScript - a custom script is set via the config
		res - the variable in which the PDF will be written
	*/
	var insertPages string
	var addNumeration string

	if len(params.Selector) > 0 {
		insertPages = fmt.Sprintf(`
		%s
		class GoJSEncoder {
			decode(text) {
				%s
				return text.replace(regex, (match, base64) => {
					try {
						return this.decodeBase64(base64);
					} catch (e) {
						console.error('Decoding error:', e);
						return match;
					}
				});
			}
			decodeBase64(base64) {
				const binaryString = atob(base64);
				if (typeof TextDecoder !== 'undefined') {
					const bytes = new Uint8Array(binaryString.length);
					for (let i = 0; i < binaryString.length; i++) {
						bytes[i] = binaryString.charCodeAt(i);
					}
					return new TextDecoder('utf-8').decode(bytes);
				} else {
					return decodeURIComponent(escape(binaryString));
				}
			}
		}

		const encoder = new GoJSEncoder()
		function InsertPages() {
			document.querySelector("%s").innerHTML = encoder.decode(content)
		}
		InsertPages()`,
			"const content = `"+string(allPages)+"`",
			"const regex = new RegExp(`"+"GOB64\\{([^}]+)\\}"+"`, 'g')",
			params.Selector,
		)
		if len(toc) > 0 {
			addNumeration = `function AddNumeration(obj) {
				var nums = JSON.parse(obj)
				for(let i = 0; i < nums.length; i++) {
					let num = nums[i];
					document.querySelector("#" + num.Link ).innerText = num.Index + " " + document.querySelector("#" + num.Link ).innerText
				}
			};
			` + fmt.Sprintf(`AddNumeration("%s")`, strings.ReplaceAll(string(toc), `"`, `\"`))
		}
	}

	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		chromedp.Evaluate(insertPages, nil),
		chromedp.Evaluate(addNumeration, nil),
		chromedp.WaitVisible(params.Selector, chromedp.ByQuery),
		chromedp.Evaluate(customScript, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(1 * time.Second)
			width := 8.26771653543307         // 21 cm in inches
			height := 11.69291338582677       // 29,70 in inches
			buf, _, err := page.PrintToPDF(). // https://chromedevtools.github.io/devtools-protocol/tot/Page/#method-printToPDF
								WithLandscape(params.FinalLandscape).
								WithPrintBackground(true).
								WithScale(0.97).
								WithMarginLeft(1).
								WithMarginTop(0.4).
								WithMarginRight(0.6).
								WithMarginBottom(1.6).
								WithPreferCSSPageSize(true).
								WithPaperWidth(width).
								WithPaperHeight(height).
								WithDisplayHeaderFooter(true).
								WithHeaderTemplate(`<span></span>`).
								WithFooterTemplate(`<span style="
								font-size: 12px;
								margin-left: auto;
								padding-right: 28px" class="pageNumber"></span>`).
				WithGenerateDocumentOutline(true).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func prepareNumeration(tocJSON []Page) ([]byte, error) {
	var numbers []Num
	var obj []byte
	var err error

	for _, page := range tocJSON {
		for _, header := range page.Headers {
			numbers = append(numbers, Num{
				Link:  header.Link,
				Index: header.Index,
			})
		}
	}

	if len(numbers) > 0 {
		obj, err = json.Marshal(numbers)
		if err != nil {
			return nil, err
		}
	}

	return obj, nil
}
