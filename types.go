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

import "github.com/prometheus/client_golang/prometheus"

type ExporterMetrics struct {
	TotalScrapes prometheus.Counter
	ScrapeErrors *prometheus.CounterVec
	ApiRequests  *prometheus.CounterVec
	Error        prometheus.Gauge
}

type Target struct {
	Latitude  string `yaml:"lat"`
	Longitude string `yaml:"lon"`
	Interval  int    `yaml:"interval"`
	Name      string `yaml:"name"`
}

type Config struct {
	ApiKey  string   `yaml:"apiKey"`
	Targets []Target `yaml:"targets"`
}
