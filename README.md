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

## License

MIT Copyright 2020 Guillaume VILLENA <guillaume@villena.me> (Willena)

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
