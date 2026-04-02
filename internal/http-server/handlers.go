package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/TOsmanov/go-pdf/core"
	"github.com/TOsmanov/go-pdf/internal/config"
	response "github.com/TOsmanov/go-pdf/internal/lib/api"
	"github.com/TOsmanov/go-pdf/internal/lib/utils"
)

func PDFHandler(log *slog.Logger,
	cfg *config.Config,
	sem *utils.Semaphore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.PDFHandler"
		log.Debug(op)
		var params core.Params
		var msg string
		var data [][]byte
		var final []byte

		b, err := io.ReadAll(r.Body)
		if err != nil {
			response.ReturnError(log, w, r, 400, "failed to read request", op, err)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(b, &params)
		if err != nil {
			response.ReturnError(
				log, w, r, 400,
				"failed to unmarshal request body", op, err)
			return
		}

		err = params.Validation(log, cfg, "pdf")
		if err != nil {
			response.ReturnError(
				log, w, r, 400,
				fmt.Sprintf(
					"failed validation: %v", err,
				), op, err)
			return
		}

		// TODO: add using cache
		log.Info("the request passed parameters",
			slog.Any("parameters", params))

		data, msg, err = MakeFinalPDF(log, cfg, &params, sem)
		if err != nil {
			response.ReturnError(
				log, w, r, 400,
				msg, op, err)
			cfg.Reload <- err
			return
		}

		log.Debug("prepare data is done",
			slog.Int("data count", len(data)))
		err = core.MergeHTMLtoPDF(data, &final, &params, cfg, log, sem)
		if err != nil {
			log.Error("failed to merge HTML to PDF",
				slog.Any("error: ", err))
			response.ReturnError(
				log, w, r, 500,
				msg, op, err)
			cfg.Reload <- err
			return
		}
		log.Debug("merge is done",
			slog.String("file size", fmt.Sprintf("%.2fMb", float64(len(final))/(1<<20))))
		w.Write(final)
		sem.Acquire()
		defer sem.Release()

		cfg.CreateBrowser()
	}
}

func DOCXHandler(log *slog.Logger,
	cfg *config.Config,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.DOCXHandler"
		log.Debug(op)
		var res []byte
		var params core.Params

		b, err := io.ReadAll(r.Body)
		if err != nil {
			response.ReturnError(log, w, r, 400, "failed to read request", op, err)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(b, &params)
		if err != nil {
			response.ReturnError(
				log, w, r, 400,
				"failed to unmarshal request body", op, err)
			return
		}

		err = params.Validation(log, cfg, "docx")
		if err != nil {
			response.ReturnError(
				log, w, r, 400,
				fmt.Sprintf(
					"failed validation: %v", err,
				), op, err)
			return
		}

		if len(params.Selector) == 0 {
			params.Selector = cfg.DocxSelector
		}

		core.DOCXGrabber(params, &res, cfg.DocxSelector)
		w.Write(res)
	}
}

func FaceHandler(log *slog.Logger,
	cfg *config.Config,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.FaceHandler"
		log.Debug(op)
		err := cfg.FaceTemplate.Execute(w, cfg.Face.Prefix)
		if err != nil {
			response.ReturnError(log, w, r, 500, "failed to execute template", op, err)
			return
		}
		log.Debug("the page has been sent successfully")
	}
}

func ReloadHandler(cfg *config.Config) http.HandlerFunc {
	return func(_ http.ResponseWriter, _ *http.Request) {
		cfg.Reload <- fmt.Errorf("user-initiated stopping")
	}
}
