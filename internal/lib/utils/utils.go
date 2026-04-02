package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SumSha256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func FileSumSha256(f *os.File) (string, error) {
	file1Sum := sha256.New()
	if _, err := io.Copy(file1Sum, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", file1Sum.Sum(nil)), nil
}

func WriteHTML(content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(content))
	})
}

func WriteImage(b []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(b)
	})
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func SaveFile(path string, data []byte) error {
	if GetEnv("SAVE_FILES", "false") == "true" {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm) //nolint:ireturn,nolintlint
		if err != nil {
			return err
		}
		err = os.WriteFile(path, data, 0o600)
		if err != nil {
			return err
		}
	}
	return nil
}

type Semaphore struct {
	C chan struct{}
}

func (s *Semaphore) Acquire() {
	s.C <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.C
}
