/*
 * Copyright 2021 Richard Kosegi
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/rkosegi/owm-exporter/internal"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	pv "github.com/prometheus/common/version"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus/exporter-toolkit/web"
)

const (
	progName = "owm_exporter"
)

var (
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9111")

	metricPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()

	disableDefaultMetrics = kingpin.Flag(
		"disable-default-metrics",
		"Exclude default metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Bool()

	configFile = kingpin.Flag(
		"config.file",
		"Path to YAML file with configuration",
	).Default("config.yaml").String()
)

func init() {
	prometheus.MustRegister(version.NewCollector(progName))
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)

	kingpin.Version(pv.Print(progName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting "+progName, "version", pv.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", pv.BuildContext())
	level.Info(logger).Log("msg", "Loading configuration from from file", "file", configFile)

	config, err := internal.LoadConfig(*configFile)

	if err != nil {
		level.Error(logger).Log("msg", "Error reading configuration", "err", err)
		os.Exit(1)
	}
	level.Info(logger).Log("msg", fmt.Sprintf("Got %d targets", len(config.Targets)))

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector(progName))

	if err := r.Register(internal.NewExporter(config, logger)); err != nil {
		level.Error(logger).Log("msg", "Couldn't register "+progName, "err", err)
		os.Exit(1)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{r},
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	if !*disableDefaultMetrics {
		r.MustRegister(collectors.NewGoCollector())
		r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		handler = promhttp.InstrumentMetricHandler(
			r, handler,
		)
	}
	landingPage, err := web.NewLandingPage(web.LandingConfig{
		Name:        strings.ReplaceAll(progName, "_", " "),
		Description: "Prometheus exporter for www.openweatherman.org",
		Version:     pv.Info(),
		Links: []web.LandingLinks{
			{
				Address: *metricPath,
				Text:    "Metrics",
			},
			{
				Address: "/health",
				Text:    "Health",
			},
		},
	})
	if err != nil {
		level.Error(logger).Log("msg", "Couldn't create landing page", "err", err)
		os.Exit(1)
	}

	http.Handle("/", landingPage)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	http.Handle(*metricPath, handler)

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting server", "err", err)
		os.Exit(1)
	}
}
