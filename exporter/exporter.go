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

package exporter

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rkosegi/owm-exporter/client"
	"github.com/rkosegi/owm-exporter/types"
)

const (
	subsystem = "exporter"
	namespace = "owm"
)

var (
	currentTemperatureDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "temperature"),
		"The current temperature.",
		[]string{"location"}, nil)

	currentTemperatureMinDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "temperature_min"),
		"The minimal currently observed temperature.",
		[]string{"location"}, nil)

	currentTemperatureMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "temperature_max"),
		"The maximal currently observed temperature.",
		[]string{"location"}, nil)

	currentTemperatureFeelDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "temperature_feel"),
		"The current temperature feel like.",
		[]string{"location"}, nil)

	currentHumidityDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "humidity"),
		"The current humidity.",
		[]string{"location"}, nil)

	currentPressureDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "pressure"),
		"The current atmospheric pressure.",
		[]string{"location"}, nil)

	currentWindSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "wind_speed"),
		"The current wind speed.",
		[]string{"location"}, nil)

	currentWindDirectionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "current", "wind_direction"),
		"The current wind direction in degrees.",
		[]string{"location"}, nil)
)

type Exporter struct {
	ctx     context.Context
	logger  log.Logger
	config  *types.Config
	metrics types.ExporterMetrics
	client  client.OwmClient
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.TotalScrapes.Desc()
	ch <- e.metrics.Error.Desc()
	e.metrics.ScrapeErrors.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(e.ctx, ch)

	ch <- e.metrics.TotalScrapes
	ch <- e.metrics.Error

	e.metrics.ScrapeErrors.Collect(ch)
	e.metrics.ApiRequests.Collect(ch)
}

func (e *Exporter) scrape(ctx context.Context, ch chan<- prometheus.Metric) {
	e.metrics.TotalScrapes.Inc()
	e.metrics.Error.Set(0)

	for _, target := range e.config.Targets {
		level.Debug(e.logger).Log("msg", "Processing target", "target", target.Name)

		respone, err := e.client.Fetch(ctx, target, e.logger)
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching current conditions",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.current." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {
			ch <- prometheus.MustNewConstMetric(currentTemperatureDesc,
				prometheus.GaugeValue, float64(respone.Main.Temp), target.Name)
			ch <- prometheus.MustNewConstMetric(currentTemperatureMinDesc,
				prometheus.GaugeValue, float64(respone.Main.TempMin), target.Name)
			ch <- prometheus.MustNewConstMetric(currentTemperatureMaxDesc,
				prometheus.GaugeValue, float64(respone.Main.TempMax), target.Name)
			ch <- prometheus.MustNewConstMetric(currentTemperatureFeelDesc,
				prometheus.GaugeValue, float64(respone.Main.TempFeel), target.Name)
			ch <- prometheus.MustNewConstMetric(currentHumidityDesc,
				prometheus.GaugeValue, float64(respone.Main.Humidity), target.Name)
			ch <- prometheus.MustNewConstMetric(currentPressureDesc,
				prometheus.GaugeValue, float64(respone.Main.Pressure), target.Name)
			ch <- prometheus.MustNewConstMetric(currentWindSpeedDesc,
				prometheus.GaugeValue, float64(respone.Wind.Speed), target.Name)
			ch <- prometheus.MustNewConstMetric(currentWindDirectionDesc,
				prometheus.GaugeValue, float64(respone.Wind.Direction), target.Name)
		}
	}
}

func NewExporter(ctx context.Context, config *types.Config, logger log.Logger,
	exporterMetrics types.ExporterMetrics) *Exporter {
	return &Exporter{
		ctx:     ctx,
		logger:  logger,
		config:  config,
		metrics: exporterMetrics,
		client:  client.NewClient(config.ApiKey, exporterMetrics),
	}
}

func NewExporterMetrics() types.ExporterMetrics {
	return types.ExporterMetrics{
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrapes_total",
			Help:      "Total number of times OWM was scraped for metrics.",
		}),
		ApiRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "api_requests",
			Help:      "Total number of API requests for given location.",
		}, []string{"location"}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a OWM.",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from OWM resulted in an error (1 for error, 0 for success).",
		}),
	}
}
