package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/TOsmanov/go-pdf/core"
	"github.com/chromedp/chromedp"
)

var (
	params    core.Params
	url       string
	format    string
	output    string
	selector  string
	landscape bool
)

func init() {
	if flag.Lookup("url") == nil {
		flag.StringVar(&url, "url", "https://www.ya.ru", "The url to the page")
	}
	if flag.Lookup("format") == nil {
		flag.StringVar(&format, "format", "pdf", "Format of the file: pdf or docx")
	}
	if flag.Lookup("selector") == nil {
		flag.StringVar(&selector, "selector", "body", "Selector")
	}
	if flag.Lookup("output") == nil {
		flag.StringVar(&output, "output", "result", "Output path without extension")
	}
	if flag.Lookup("landscape") == nil {
		flag.BoolVar(&landscape, "landscape", false, "Landscape of the page")
	}
}

func main() {
	flag.Parse()
	start := time.Now()
	var buffer []byte
	outputFilename := fmt.Sprintf("%s.%s", output, strings.ToLower(format))

	if landscape {
		params.Landscape = "1"
	} else {
		params.Landscape = "0"
	}

	if format == "pdf" {
		taskCtx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		defer cancel()
		if err := chromedp.Run(
			taskCtx,
			core.PDFGrabber(0, url, &params, "", &buffer, nil)); err != nil {
			log.Panic(err)
		}
		defer func() {
			chromedp.Cancel(taskCtx)
		}()
	} else if format == "docx" {
		params.Urls = []string{url}
		err := core.DOCXGrabber(params, &buffer, "")
		if err != nil {
			log.Panic(err)
		}
	}
	if err := os.WriteFile(outputFilename, buffer, 0o600); err != nil {
		log.Panic(err)
	}

	fmt.Printf("\nTook: %.2f secs\n", time.Since(start).Seconds())
}
