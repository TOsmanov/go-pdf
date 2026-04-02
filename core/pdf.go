package core

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const GetPageOrient = `function PageOrient() {
    let result = false;
    const tables = document.querySelectorAll('table');

    tables.forEach(table => {
        const tbody = table.querySelector('tbody');
        if (tbody && tbody.scrollWidth > 1080) {
            result = true;
        }
    });
    return result;
}

PageOrient();`

func PageGrabber(i int, url string, params *Params,
	customScript string, res *string, pageToC *string,
) chromedp.Tasks {
	/*
		Creates a chrome task to export the html page of the selected selector to a string.
		Add data into pageToC
	*/
	var landscape bool
	var parsePage string
	var parseContent string

	if len(params.Selector) > 0 && len(params.TitleSelector) > 0 {
		parsePage = fmt.Sprintf(`function ParsePage() {
			page = []
			headers = []
			sP = "%s"
			sHeaders = sP + " h2," + sP + " h3," + sP + " h4," + sP + " h5," + sP + " h6"
			section = document.querySelector(sP)
			pageI = %d
			pageId = "%05d"
			if (section) {
				document.querySelectorAll(sHeaders).forEach(function(item, i) {
					if (item.innerText.length > 0 ) {
						item.id = "gopdf-header-" + pageId + "h" + i
						headers.push( {
							lvl: Number(item.tagName.slice(-1)),
							text: item.innerText,
							link: item.id,
						})
					}
				})
			}
			selectorTitle = document.querySelector("%s")
			if (selectorTitle) {
				selectorTitle.id = "gopdf-page-header-" + pageId
				page = {
					title: selectorTitle.innerText,
					anchor: selectorTitle.id,
					headers: headers
				}
				selectorTitle.innerText = pageI + 1 + '. ' + selectorTitle.innerText
			}
			return JSON.stringify(page)
		}
		ParsePage()`, params.Selector, i, i, params.TitleSelector)
	}

	replaceRelToAbs := `let allLinks = document.querySelectorAll('a')
	for (let el of allLinks) {
		el.href = new URL(el.href, document.baseURI).href
	}`

	if len(params.Selector) > 0 {
		parseContent = fmt.Sprintf(`
		function ParseContent() {
			return document.querySelector("%s").innerHTML
		}
		ParseContent()`, params.Selector)
	}

	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		// chromedp.WaitVisible(params.Selector, chromedp.ByQuery),
		chromedp.Evaluate(customScript, nil),
		chromedp.Evaluate(GetPageOrient, &landscape),
		chromedp.Evaluate(parsePage, &pageToC),
		chromedp.Evaluate(replaceRelToAbs, nil),
		chromedp.Evaluate(parseContent, &res),
		chromedp.ActionFunc(func(ctx context.Context) error { //nolint
			if !params.FinalLandscape {
				if params.Landscape == "auto" && landscape {
					params.FinalLandscape = true
				}
				if params.Landscape == "1" {
					params.FinalLandscape = true
				}
			}
			return nil
		}),
	}
}

func PDFGrabber(i int, url string, params *Params,
	customScript string, res *[]byte, pageToC *string,
) chromedp.Tasks {
	/*
		USED ONLY FOR CLI
		Creates a chrome task to export a single URL to pdf.
		i - the ID required to save the sequence of pages when calling the function via goroutine
		url - the URL that will be opened in chrome
		params - the structure of the request parameters
		customScript - a custom script is set via the config
		res - the variable in which the PDF will be written
		pageToC - the structure required for generating service pages
	*/
	var landscape *bool
	var parsePage string

	getPageOrient := `function PageOrient() {
		result = false
		if (window.jQuery) {
			$('table').each(function () {
				if ($(this).find('tbody')[0].scrollWidth > 1080) {
					result = true
				}
			})
		}
		return result
	};
	PageOrient()`

	if len(params.Selector) > 0 && len(params.TitleSelector) > 0 {
		parsePage = fmt.Sprintf(`function ParsePage() {
			page = []
			headers = []
			selectorHeader = document.querySelector("%s")
			if (selectorHeader) {
				selectorHeader.querySelectorAll("h2, h3, h4, h5, h6").forEach(function(item, i) {
					headers.push( {
						lvl: Number(item.tagName.slice(-1)),
						text: item.innerText,
						link: item.id,
						y: item.offsetTop
					})
				})
			}
			selectorTitle = document.querySelector("%s")
			if (selectorTitle) {
				selectorTitle.id = "gopdf-page-header-%05d"
				page = {
					title: selectorTitle.innerText,
					anchor: selectorTitle.id,
					headers: headers
				}
			}
			return JSON.stringify(page)
		}
		window.print()
		ParsePage()`, params.Selector, params.TitleSelector, i)
	}
	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		// chromedp.WaitVisible(params.Selector, chromedp.ByQueryAll), // This feature sometimes takes a very long time to load a page.
		chromedp.Evaluate(getPageOrient, &landscape),
		chromedp.Evaluate(parsePage, &pageToC),
		chromedp.Evaluate(customScript, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			width := 8.26771653543307   // 21 cm in inches
			height := 11.69291338582677 // 29,70 in inches
			if params.Landscape == "auto" && *landscape {
				params.FinalLandscape = true
			}
			if params.Landscape == "1" {
				params.FinalLandscape = true
			}
			buf, _, err := page.PrintToPDF().
				WithLandscape(params.FinalLandscape).
				WithPrintBackground(true).
				WithMarginLeft(1).
				WithMarginTop(0.4).
				WithMarginRight(0.6).
				WithMarginBottom(1.6).
				WithPreferCSSPageSize(true).
				WithPaperWidth(width).
				WithPaperHeight(height).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
