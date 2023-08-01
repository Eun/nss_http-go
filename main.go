package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type flushWriter struct {
	file string
}

func (w *flushWriter) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(w.file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	return f.Write(p)
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
	if os.Getenv("NSS_HTTP_DEBUG") != "" {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}
	log.Logger = log.Logger.Output(&flushWriter{
		file: "/var/log/nss_http.log",
	})
}

func main() {}
