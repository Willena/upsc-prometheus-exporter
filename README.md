# upsc-prometheus-exporter

Monitor your ups with prometheus using this exporter

## Features

- Support for multiple UPS exports (using prometheus labels)
- Support exclusion partern (raw string or regex) for some keys (drivers for example)
- Customisable interval between updates
- Customisable upsc binary path
- Customisable listen host and port
- Uses prometheus client library
- Dynamically generated metrics

## Build 

```console
go build
```

## Run

```
./ups-prometheus-exporter -u myups@localhost -p 8081 -H 127.0.0.1
```

## Docker

### Build
```
docker build -t gillena/ups-promtheus-exporter .
```

### Run

```
docker run --rm -p 8081:8081 gillena/ups-promtheus-exporter -u myups@localhost -p 8081
```

## Dahsboards

- https://grafana.com/grafana/dashboards/11712
- https://github.com/klippo/nut_exporter/blob/master/dashboards/grafana.json

## Credits 

This project is inspired from the following ones 

- https://github.com/p404/nut_exporter
- https://e1e0.net/upsc-prometheus-exporter.html
- https://github.com/mabunixda/nut_exporter
- https://github.com/klippo/nut_exporter