package main

import "C"
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Eun/nss_http/types"
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

func getLogFilePath() string {
	logFile := os.Getenv("NSS_HTTP_LOG_FILE")
	switch logFile {
	case "disable", "disabled", "off":
		return ""
	}
	if logFile != "" {
		return logFile
	}
	return "/var/log/nss_http.log"
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
	if os.Getenv("NSS_HTTP_DEBUG") != "" {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}
	logWriter := io.Discard
	if logFile := getLogFilePath(); logFile != "" {
		logWriter = &flushWriter{
			file: logFile,
		}
	}
	log.Logger = log.Logger.Output(logWriter)
}

func usage() {
	fmt.Println("nss_http sshkey <username>")
}

func main() {
	if len(os.Args) == 0 {
		fmt.Println("critical error")
		os.Exit(1)
		return
	}

	args := os.Args
	switch strings.ToLower(filepath.Base(args[0])) {
	case "nss_http_sshkey":
		args = append([]string{"sshkey"}, args[1:]...)
	default:
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Println("invalid or missing arguments")
		usage()
		os.Exit(1)
		return
	}
	switch args[0] {
	case "sshkey":
		if len(args) != 2 {
			fmt.Println("invalid or missing arguments")
			usage()
			os.Exit(1)
			return
		}
		user, err := getUser(types.NameIdentifier(args[1]))
		if err != nil {
			log.Err(err).Str("name", args[1]).Msg("unable to get user by name")
			os.Exit(1)
			return
		}
		if user == nil {
			log.Debug().Str("name", args[1]).Msg("user not found")
			os.Exit(1)
			return
		}

		log.Debug().Any("user", user).Msg("user found")
		for _, key := range user.AuthKeys {
			fmt.Println(key)
		}
		return
	}
	fmt.Println("invalid or missing arguments")
	usage()
	os.Exit(1)
}
