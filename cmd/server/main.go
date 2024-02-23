package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/mxcd/go-config/config"
	"github.com/mxcd/outline-pdf-converter/internal/converter"
	"github.com/mxcd/outline-pdf-converter/internal/util"
	"github.com/mxcd/outline-pdf-converter/internal/web"
)

func main() {
	if err := util.InitConfig(); err != nil {
		log.Panic().Err(err).Msg("error initializing config")
	}

	if err := util.InitLogger(); err != nil {
		log.Panic().Err(err).Msg("error initializing logger")
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/file-upload", converter.GetHandler())
	http.HandleFunc("/status", converter.GetStatusHandler())
	http.HandleFunc("/download", converter.GetDownloadHandler())

	webFileHandler := web.GetHandleFunc()
	if !config.Get().Bool("DEV") {
		webFileHandler = web.CacheControlMiddleware(webFileHandler)
	}

	http.Handle("/", webFileHandler)

	log.Info().Msg("starting server")
	portString := fmt.Sprintf(":%d", config.Get().Int("PORT"))
	http.ListenAndServe(portString, nil)
}
