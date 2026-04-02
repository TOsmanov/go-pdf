package response

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Body   string `json:"body,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func Write(body string) Response {
	return Response{
		Status: StatusOK,
		Body:   body,
	}
}

func ReturnError(log *slog.Logger,
	w http.ResponseWriter, r *http.Request,
	code int, msg string, op string, err error,
) {
	w.WriteHeader(code)
	log.Error(
		msg,
		op, err)
	render.JSON(w, r,
		Error(msg))
}
