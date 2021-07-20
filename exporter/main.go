package main

import (
	"log"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	libHTTP "github.com/willie-yao/brigade-metrics/exporter/internal/http"
	"github.com/willie-yao/brigade-metrics/exporter/internal/signals"
	"github.com/willie-yao/brigade-metrics/exporter/internal/system"
	"github.com/willie-yao/brigade-metrics/exporter/internal/version"
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
		go newMetricsExporter(
			sdk.NewAPIClient(address, token, &opts),
			time.Duration(scrapeInterval),
		).run(ctx)
	}

	var server libHTTP.Server
	{
		router := mux.NewRouter()
		router.StrictSlash(true)
		router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
		router.HandleFunc("/healthz", system.Healthz).Methods(http.MethodGet)
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
