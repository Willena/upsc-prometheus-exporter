package main

import (
	"bufio"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	upsList          *[]string
	excludeKeyRegex  []*regexp.Regexp
	excludeKeyString *[]string
	upscPath         *string

	prometheusMetrics map[string]map[string]prometheus.Gauge
)

func main() {

	// Create new parser object
	parser := argparse.NewParser("print", "Prints provided string to stdout")
	// Create string flag
	upsList = parser.StringList("u", "ups", &argparse.Options{Required: true, Help: "List of Ups name"})
	excludeKeyString = parser.StringList("e", "exclude", &argparse.Options{Required: false, Help: "Do not export Key"})
	upscPath = parser.String("e", "executable", &argparse.Options{Required: false, Help: "Path to the executable of upsc", Default: "upsc"})
	var excludeKeyStringRegex = parser.StringList("E", "exclude-regex", &argparse.Options{Required: false, Help: "Do not export Key using Regex"})
	var port = parser.Int("p", "port", &argparse.Options{Required: false, Help: "Port Number", Default: 8081})
	var host = parser.String("H", "host", &argparse.Options{Required: false, Help: "Define the host address for the web server", Default: "127.0.0.1"})
	var sampleInterval = parser.Int("i", "interval", &argparse.Options{Required: false, Help: "Sample interval in second", Default: 5})
	//var debugLevel = parser.FlagCounter("v", "verbose", &argparse.Options{Required: false, Help: "Enable debug", Default: 1})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	//Compile regex given at input
	for _, regexstring := range *excludeKeyStringRegex {
		excludeKeyRegex = append(excludeKeyRegex, regexp.MustCompile(regexstring))
	}

	var listenAddr = fmt.Sprintf("%s:%d", *host, *port)
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Started webserver on ", listenAddr)
	log.Info("Giving metrics for ", *upsList)
	log.Info("The folowing keys are ignored: ", *excludeKeyString)
	log.Info("The folling regex will exclude keys: ", *excludeKeyStringRegex)

	for _, upsName := range *upsList {
		sampleUps(&upsName, *sampleInterval)
	}

	http.ListenAndServe(listenAddr, nil)

}

func sampleUps(upsName *string, sampleInterval int) {
	go func() {

		for {

			cmd := exec.Command(*upscPath, *upsName)
			stdout, err := cmd.StdoutPipe()

			if err != nil {
				log.Error("Error while using upsc for ", *upsName, ": error piping stdout. ", *upsName, " will now be ignored ")
				return
			}
			if err := cmd.Start(); err != nil {
				log.Error("Error while using upsc for ", *upsName, ": error executing command. ", *upsName, " will now be ignored ")
				return
			}

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				parts := strings.Split(scanner.Text(), ": ")

				if ignoredKey(parts[0]) {
					log.Debug("the key ", parts[0], " is ignored")
					continue
				}

				// if we can convert to float assign a value
				// if not, just check for ups status (OL || OB)

				if parts[0] == "ups.status" {
					gauge := CreateAndRegisterGauge(*upsName, parts[0], prometheusMetrics[*upsName])
					//TODO: Check and apply each status number
					switch parts[1] {
					case "CAL":
						gauge.Set(0)
					case "TRIM":
						gauge.Set(1)
					case "BOOST":
						gauge.Set(2)
					case "OL":
						gauge.Set(3)
					case "OB":
						gauge.Set(4)
					case "OVER":
						gauge.Set(5)
					case "LB":
						gauge.Set(6)
					case "RB":
						gauge.Set(7)
					case "BYPASS":
						gauge.Set(8)
					case "OFF":
						gauge.Set(9)
					case "CHRG":
						gauge.Set(10)
					case "DISCHRG":
						gauge.Set(11)
					}
				} else {
					value, err := strconv.ParseFloat(parts[1], 64)
					if err == nil {
						gauge := CreateAndRegisterGauge(*upsName, parts[0], prometheusMetrics[*upsName])
						gauge.Set(value)
					}
				}

				if err := scanner.Err(); err != nil {
					log.Info("Error while using upsc for ", *upsName, "error scanning command output ", *upsName, " will now be ignored ")
				}

				cmd.Wait() // needed to close the pipe and release the child process
				time.Sleep(time.Duration(sampleInterval) * time.Second)
			}
		}
	}()

}

func ignoredKey(key string) bool {
	for _, excludeString := range *excludeKeyString {
		if key == excludeString {
			return true
		}
	}

	for _, excludeRegex := range excludeKeyRegex {
		if excludeRegex.MatchString(key) {
			return true
		}
	}

	return false
}

func toPrometheusKey(s string) string {
	return "upsc_" + strings.ReplaceAll(s, ".", "_")
}

func CreateAndRegisterGauge(upsName string, key string, m map[string]prometheus.Gauge) prometheus.Gauge {

	key = toPrometheusKey(key)

	if val, ok := m[key]; ok {
		return val
	} else {
		log.Info("Creating key " + key)
		constLabels := map[string]string{"ups": upsName}
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        key,
			ConstLabels: constLabels,
		})

		prometheus.MustRegister(gauge)
		m[key] = gauge
		return gauge
	}
}
