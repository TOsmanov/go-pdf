package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/TOsmanov/go-pdf/core"
	"github.com/TOsmanov/go-pdf/internal/config"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
	"github.com/chromedp/chromedp"
)

var (
	wg sync.WaitGroup
	m  sync.Mutex
)

func prepareToCData(m map[int]core.Page, count int) []core.Page {
	var arr []core.Page
	for i := 0; i <= count; i++ {
		if len(m[i].Title) > 0 {
			arr = append(arr, m[i])
		}
	}
	return arr
}

func MakeFinalPDF(log *slog.Logger,
	cfg *config.Config,
	params *core.Params,
	sem *utils.Semaphore,
) ([][]byte, string, error) {
	type msgError struct {
		msg string
		err error
	}

	size := len(params.Urls)
	output := make(map[int][]byte, size)
	tocData := make(map[int]core.Page, size)
	errChan := make(chan msgError, 1)
	errChan <- msgError{
		msg: "",
		err: nil,
	}

	for i, url := range params.Urls {
		sem.Acquire()
		wg.Add(1)
		log.Debug("start chrome for grab",
			slog.Int("index", i),
			slog.Any("URL", url))

		timeout := make(chan bool, 1)
		done := make(chan bool, 1)

		go func() {
			time.Sleep(cfg.ChromeTimeout)
			timeout <- true
		}()

		go func(i int, url string) {
			defer wg.Done()
			defer sem.Release()

			var res string
			var pageToC string
			var page core.Page

			if err := chromedp.Run(*cfg.Browser.Ctx,
				core.PageGrabber(i, url,
					params, cfg.GrabScript,
					&res, &pageToC)); err != nil {
				errChan <- msgError{
					msg: "failed page grab",
					err: err,
				}
			}

			done <- true

			if len(pageToC) > 0 {
				err := json.Unmarshal([]byte(pageToC), &page)
				if err != nil {
					errChan <- msgError{
						msg: "failed to unmarshal table of content",
						err: err,
					}
				}
				m.Lock()
				tocData[i] = page
				m.Unlock()
			}

			m.Lock()
			output[i] = []byte(res)
			m.Unlock()

			log.Debug("page grab is done",
				slog.Int("index", i),
				slog.String("URL", url))
		}(i, url)

		select {
		case <-done:
			close(done)
		case <-timeout:
			return nil, "the timeout has expired", fmt.Errorf("index: %d, url: %s", i, url)
		}
	}

	wg.Wait()

	errMsg := <-errChan
	defer close(errChan)
	if errMsg.err != nil {
		return nil, errMsg.msg, errMsg.err
	}
	log.Debug("grab of all pages has been done successfully")

	log.Debug("prepare HTML data")
	data := prepareHTMLData(output, size)
	log.Debug("prepare ToC data")
	params.ToC = prepareToCData(tocData, size)

	if params.EnableToCPage {
		var ToCPage string
		err := core.ToCPageHTML(params, &ToCPage, cfg)
		if err != nil {
			return nil, "failed to prepare ToC pages", err
		}
		log.Debug("prepare the ToC pages has been done successfully")
		data = append([][]byte{[]byte(ToCPage)}, data...)
	}

	if params.EnableTitlePage {
		var TitlePage string
		err := core.TitlePageHTML(params, &TitlePage, cfg)
		if err != nil {
			return nil, "failed to prepare title page", err
		}
		log.Debug("prepare title page has been done successfully")
		data = append([][]byte{[]byte(TitlePage)}, data...)
	}

	return data, "", nil
}

func prepareHTMLData(m map[int][]byte, count int) [][]byte {
	var arr [][]byte
	for i := 0; i <= count; i++ {
		if len(m[i]) > 0 {
			encoded := make([]byte, base64.StdEncoding.EncodedLen(len(m[i])))
			base64.StdEncoding.Encode(encoded, m[i])

			result := make([]byte, 0, len(encoded)+6) // 6 = "GOB64{" + "}"
			result = append(result, "GOB64{"...)
			result = append(result, encoded...)
			result = append(result, '}')

			arr = append(arr, result)
		}
	}
	return arr
}
