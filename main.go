package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appPasswordEnvKey  = "PHOTOPRISM_APP_PASSWORD"
	appPasswordDocsUrl = "https://docs.photoprism.app/user-guide/users/client-credentials/#app-passwords"
)

type RequestBody struct {
	Path string `json:"path"`
	Move bool   `json:"move"`
}

type ResponseBody struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	debug := flag.Bool("debug", false, "Log filesystem events")
	delay := flag.Duration("delay", 10*time.Second, "How soon after the last filesystem event to trigger the import")
	move := flag.Bool("move", false, "Tell PhotoPrism to remove imported files")
	apiURL := flag.String("url", "http://127.0.0.1:2342/api/v1/", "PhotoPrism API URL")
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage: %s [options] PHOTOPRISM_IMPORT_PATH\n",
			os.Args[0],
		)
		flag.PrintDefaults()
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"\nThe %s environment variable must be set to an app-specific password.\nSee %s for details.\n",
			appPasswordEnvKey, appPasswordDocsUrl,
		)
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	path := flag.Arg(0)
	fs, err := os.Stat(path)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	if !fs.IsDir() {
		log.Fatal().Msgf("%s: not a directory", path)
	}
	*apiURL, err = url.JoinPath(*apiURL, "/import/")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	token := os.Getenv(appPasswordEnvKey)
	if token == "" {
		log.Fatal().Msgf("%s environment variable is required", appPasswordEnvKey)
	}

	doRequest := func() {
		log.Debug().Msg("Sending request")
		body, err := json.Marshal(RequestBody{"/", *move})
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}
		req, err := http.NewRequest("POST", *apiURL, bytes.NewBuffer(body))
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}
		var response ResponseBody
		if err = json.Unmarshal(body, &response); err != nil {
			log.Error().Msgf("Unexpected response: %d => %s", resp.StatusCode, string(body))
			return
		}
		if response.Code != 200 {
			log.Error().Msgf("Unexpected response: %d => %s", response.Code, response.Error)
			return
		}
		log.Info().Msgf("%d => %s", response.Code, response.Message)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	defer watcher.Close()

	go func() {
		var timer *time.Timer

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Debug().Str("op", event.Op.String()).Str("path", event.Name).Msg("")
				if event.Has(fsnotify.Remove) {
					continue
				}
				if timer == nil {
					timer = time.AfterFunc(*delay, doRequest)
				} else {
					timer.Reset(*delay)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Err(err).Msg("")
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Info().Msgf("Watching %s", path)
	<-make(chan struct{})
}
