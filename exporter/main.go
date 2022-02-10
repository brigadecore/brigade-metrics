package main

import (
	"log"
	"net/http"

	libHTTP "github.com/brigadecore/brigade-foundations/http"
	"github.com/brigadecore/brigade-foundations/signals"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Printf(
		"Starting Brigade Metrics Exporter -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	ctx := signals.Context()

	{
		address, token, opts, err := apiClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		scrapeInterval, err := scrapeDuration()
		if err != nil {
			log.Fatal(err)
		}
		newMetricsExporter(
			sdk.NewAPIClient(address, token, &opts),
			scrapeInterval,
		).start(ctx)
	}

	var server libHTTP.Server
	{
		router := mux.NewRouter()
		router.StrictSlash(true)
		router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
		router.HandleFunc("/healthz", libHTTP.Healthz).Methods(http.MethodGet)
		serverConfig, err := serverConfig()
		if err != nil {
			log.Fatal(err)
		}
		server = libHTTP.NewServer(router, &serverConfig)
	}

	log.Println(
		server.ListenAndServe(signals.Context()),
	)
}
