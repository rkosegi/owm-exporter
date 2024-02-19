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
	owm "github.com/briandowns/openweathermap"
	"time"
)

type Target struct {
	Latitude  float64 `yaml:"lat"`
	Longitude float64 `yaml:"lon"`
	// in seconds
	Interval float64 `yaml:"interval"`
	Name     string  `yaml:"name"`
}

type ApiResponse struct {
	Main       owm.Main   `json:"main"`
	Visibility int        `json:"visibility"`
	Wind       owm.Wind   `json:"wind"`
	Rain       owm.Rain   `json:"rain"`
	Snow       owm.Snow   `json:"snow"`
	Clouds     owm.Clouds `json:"clouds"`
	Name       string     `json:"name"`
}

type CacheEntry struct {
	lastResponse *ApiResponse
	lastUpdated  time.Time
}

type Config struct {
	ApiKey  string   `yaml:"apiKey"`
	Targets []Target `yaml:"targets"`
}
