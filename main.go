// Copyright 2021 Richard Kosegi
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/exporter-toolkit/web"
)

const (
	PROG_NAME = "owm_exporter"
)

var (
	webConfig = webflag.AddFlags(kingpin.CommandLine)

	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(":9111").String()

	metricPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()

	configFile = kingpin.Flag(
		"config.file",
		"Path to YAML file with configuration",
	).Default("config.yaml").String()
)

func init() {
	prometheus.MustRegister(version.NewCollector(PROG_NAME))
}

func newHandler(config *Config, logger log.Logger, exporterMetrics ExporterMetrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		registry := prometheus.NewRegistry()
		registry.MustRegister(NewExporter(ctx, config, logger, exporterMetrics))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(PROG_NAME))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	//nolint:errcheck
	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s", PROG_NAME),
		"version", version.Info(),
		"config", *configFile)

	config, err := LoadConfig(*configFile)

	if err != nil {
		//nolint:errcheck
		level.Error(logger).Log("msg", "Error reading configuration", "err", err)
		os.Exit(1)
	}

	//nolint:errcheck
	level.Info(logger).Log("msg", fmt.Sprintf("Got %d targets", len(config.Targets)))

	var landingPage = []byte(`<html>
<head><title>OWM exporter</title></head>
<body>
<h1>OWM exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

	handlerFunc := newHandler(config, logger, NewExporterMetrics())
	http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err = w.Write(landingPage); err != nil {
			//nolint:errcheck
			level.Error(logger).Log("msg", "Unable to write page content", "err", err)
		}
	})

	//nolint:errcheck
	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)

	srv := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		//nolint:errcheck
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
