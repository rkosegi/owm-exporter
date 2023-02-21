# OpenWeatherMap Prometheus exporter

This is prometheus exporter for Openweathermap current conditions at particular place.


### Building

#### Docker build

```shell
make build-docker
```

#### Local build


```shell
make build-local
```

### Running

First, obtain API token from [OpenWeatherMap service](https://home.openweathermap.org/users/sign_up).
Now you can define one or more targets for scraping.

Here is example `config.yaml`:

```yaml
---
apiKey: 12345678998765432345678
targets:
  - name: Vienna
    lat: 48.2082
    lon: 16.3738
    interval: 600
```

Start exporter

```shell
./owm-exporter
level=info ts=2021-11-08T11:59:35.176Z caller=main.go:89 msg="Starting owm_exporter" version="(version=, branch=, revision=)"
level=info ts=2021-11-08T11:59:35.176Z caller=main.go:90 msg="Build context" (gogo1.16.10,user,date)=(MISSING)
level=info ts=2021-11-08T11:59:35.176Z caller=main.go:91 msg="Loading configuration" config=config.yaml
level=info ts=2021-11-08T11:59:35.176Z caller=main.go:100 msg="Got 1 targets"
level=info ts=2021-11-08T11:59:35.176Z caller=main.go:117 msg="Listening on address" address=:9111
level=info ts=2021-11-08T11:59:35.176Z caller=tls_config.go:191 msg="TLS is disabled." http2=false
```

Example collector output from `curl --silent localhost:9111/metrics | grep ^owm_
`
```
owm_current_humidity{location="Vienna"} 62
owm_current_pressure{location="Vienna"} 1022
owm_current_temperature{location="Vienna"} 10.729999542236328
owm_current_temperature_feel{location="Vienna"} 9.479999542236328
owm_current_temperature_max{location="Vienna"} 12.279999732971191
owm_current_temperature_min{location="Vienna"} 9.350000381469727
owm_current_wind_direction{location="Vienna"} 333
owm_current_wind_speed{location="Vienna"} 1.340000033378601
owm_exporter_api_requests{location="Vienna"} 1
owm_exporter_build_info{branch="",goversion="go1.16.10",revision="",version=""} 1
owm_exporter_last_scrape_error 0
owm_exporter_scrapes_total 1

```
