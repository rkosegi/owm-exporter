/*
 * Copyright 2024 Richard Kosegi
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

package internal

import (
	"context"
	"log/slog"
	"sync"
	"time"

	owm "github.com/briandowns/openweathermap"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/rkosegi/owm-exporter/types"
)

const (
	subsystem = "exporter"
	namespace = "owm"
)

var (
	currentTemperatureDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "temperature"),
		"The current temperature.",
		[]string{"location"}, nil)

	currentTemperatureMinDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "temperature_min"),
		"The minimal currently observed temperature.",
		[]string{"location"}, nil)

	currentTemperatureMaxDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "temperature_max"),
		"The maximal currently observed temperature.",
		[]string{"location"}, nil)

	currentTemperatureFeelDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "temperature_feel"),
		"The current temperature feel like.",
		[]string{"location"}, nil)

	currentHumidityDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "humidity"),
		"The current humidity.",
		[]string{"location"}, nil)

	currentPressureDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "pressure"),
		"The current atmospheric pressure.",
		[]string{"location"}, nil)

	currentWindSpeedDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "wind_speed"),
		"The current wind speed.",
		[]string{"location"}, nil)

	currentWindDirectionDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "wind_direction"),
		"The current wind direction in degrees.",
		[]string{"location"}, nil)

	currentCloudsDirectionDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "clouds"),
		"The current cloudiness in percent.",
		[]string{"location"}, nil)

	currentRain1hVolumeDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "rain_1h"),
		"Rain volume for the last 1 hour, in millimeters.",
		[]string{"location"}, nil)

	currentRain3hVolumeDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "rain_3h"),
		"Rain volume for the last 3 hours, in millimeters.",
		[]string{"location"}, nil)

	currentSnow1hVolumeDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "snow_1h"),
		"Snow volume for the last 1 hour, in millimeters.",
		[]string{"location"}, nil)

	currentSnow3hVolumeDesc = prom.NewDesc(
		prom.BuildFQName(namespace, "current", "snow_3h"),
		"Snow volume for the last 3 hours, in millimeters.",
		[]string{"location"}, nil)
)

type exporter struct {
	lock         sync.Mutex
	cache        map[string]types.CacheEntry
	ctx          context.Context
	logger       *slog.Logger
	config       *types.Config
	totalScrapes prom.Summary
	apiRequests  *prom.CounterVec
	scrapeErrors *prom.CounterVec
	error        prom.Gauge
	cacheHit     *prom.CounterVec
}

func (e *exporter) Describe(ch chan<- *prom.Desc) {
	ch <- e.totalScrapes.Desc()
	ch <- e.error.Desc()
	e.apiRequests.Describe(ch)
	e.cacheHit.Describe(ch)
	e.scrapeErrors.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prom.Metric) {
	e.scrape(ch)

	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
	e.apiRequests.Collect(ch)
	e.cacheHit.Collect(ch)
}

func (e *exporter) scrape(ch chan<- prom.Metric) {
	start := time.Now()
	defer e.totalScrapes.Observe(time.Since(start).Seconds())
	e.error.Set(0)

	for _, target := range e.config.Targets {
		e.logger.Debug("Processing target", "target", target.Name)

		resp, err := e.FetchTarget(target)
		if err != nil {
			e.logger.Error("Error while fetching current conditions", "name", target.Name, "err", err)
			e.scrapeErrors.WithLabelValues("collect.current." + target.Name).Inc()
			e.error.Set(1)
		} else {
			ch <- prom.MustNewConstMetric(currentTemperatureDesc, prom.GaugeValue, resp.Main.Temp, target.Name)
			ch <- prom.MustNewConstMetric(currentTemperatureMinDesc, prom.GaugeValue, resp.Main.TempMin, target.Name)
			ch <- prom.MustNewConstMetric(currentTemperatureMaxDesc, prom.GaugeValue, resp.Main.TempMax, target.Name)
			ch <- prom.MustNewConstMetric(currentTemperatureFeelDesc, prom.GaugeValue, resp.Main.FeelsLike, target.Name)
			ch <- prom.MustNewConstMetric(currentHumidityDesc, prom.GaugeValue, float64(resp.Main.Humidity), target.Name)
			ch <- prom.MustNewConstMetric(currentPressureDesc, prom.GaugeValue, resp.Main.Pressure, target.Name)
			ch <- prom.MustNewConstMetric(currentWindSpeedDesc, prom.GaugeValue, resp.Wind.Speed, target.Name)
			ch <- prom.MustNewConstMetric(currentWindDirectionDesc, prom.GaugeValue, resp.Wind.Deg, target.Name)
			if resp.Clouds.All != 0 {
				ch <- prom.MustNewConstMetric(currentCloudsDirectionDesc, prom.GaugeValue, float64(resp.Clouds.All), target.Name)
			}
			if resp.Snow.OneH != 0 {
				ch <- prom.MustNewConstMetric(currentSnow1hVolumeDesc, prom.GaugeValue, resp.Snow.OneH, target.Name)
			}
			if resp.Snow.ThreeH != 0 {
				ch <- prom.MustNewConstMetric(currentSnow3hVolumeDesc, prom.GaugeValue, resp.Snow.ThreeH, target.Name)
			}
			if resp.Rain.OneH != 0 {
				ch <- prom.MustNewConstMetric(currentRain1hVolumeDesc, prom.GaugeValue, resp.Rain.OneH, target.Name)
			}
			if resp.Rain.ThreeH != 0 {
				ch <- prom.MustNewConstMetric(currentRain3hVolumeDesc, prom.GaugeValue, resp.Rain.ThreeH, target.Name)
			}
		}
	}
}

func (e *exporter) responseFromCache(tgt types.Target) *types.ApiResponse {
	e.lock.Lock()
	defer e.lock.Unlock()
	entry, present := e.cache[tgt.Name]
	if present && time.Now().Unix() < int64(tgt.Interval)+entry.LastUpdated.Unix() {
		e.cacheHit.WithLabelValues(tgt.Name).Inc()
		return entry.LastResponse
	}
	return nil
}

func (e *exporter) FetchTarget(target types.Target) (*types.ApiResponse, error) {
	if fromCache := e.responseFromCache(target); fromCache != nil {
		return fromCache, nil
	}
	oc, err := owm.NewCurrent("C", "EN", e.config.ApiKey)
	if err != nil {
		return nil, err
	}
	err = oc.CurrentByCoordinates(&owm.Coordinates{
		Longitude: target.Longitude,
		Latitude:  target.Latitude,
	})
	e.apiRequests.WithLabelValues(target.Name).Inc()
	if err != nil {
		return nil, err
	}
	resp := &types.ApiResponse{
		Main:       oc.Main,
		Wind:       oc.Wind,
		Visibility: oc.Visibility,
		Rain:       oc.Rain,
		Snow:       oc.Snow,
		Clouds:     oc.Clouds,
		Name:       oc.Name,
	}
	e.lock.Lock()
	e.cache[target.Name] = types.CacheEntry{
		LastResponse: resp,
		LastUpdated:  time.Now(),
	}
	e.lock.Unlock()
	return resp, nil
}

func (e *exporter) setup() {
	e.totalScrapes = prom.NewSummary(prom.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "scrapes_total",
		Help:      "Total number of times OWM was scraped for metrics.",
	})
	e.apiRequests = prom.NewCounterVec(prom.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "api_requests",
		Help:      "Total number of API requests for given location.",
	}, []string{"location"})
	e.cacheHit = prom.NewCounterVec(prom.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cache_hit",
		Help:      "Total number of cache hits for given location.",
	}, []string{"location"})

	e.scrapeErrors = prom.NewCounterVec(prom.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "scrape_errors_total",
		Help:      "Total number of times an error occurred scraping a OWM.",
	}, []string{"collector"})

	e.error = prom.NewGauge(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "last_scrape_error",
		Help:      "Whether the last scrape of metrics from OWM resulted in an error (1 for error, 0 for success).",
	})
}

func NewExporter(config *types.Config, logger *slog.Logger) prom.Collector {
	e := &exporter{
		logger: logger,
		config: config,
		cache:  map[string]types.CacheEntry{},
	}
	e.setup()
	return e
}
