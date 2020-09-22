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

	prometheusMetrics map[string]prometheus.GaugeVec
)

func main() {

	// Create new parser object
	parser := argparse.NewParser("print", "Prints provided string to stdout")
	// Create string flag
	upsList = parser.StringList("u", "ups", &argparse.Options{Required: true, Help: "List of Ups name"})
	excludeKeyString = parser.StringList("e", "exclude", &argparse.Options{Required: false, Help: "Do not export Key"})
	upscPath = parser.String("b", "binary", &argparse.Options{Required: false, Help: "Path to the executable of upsc", Default: "upsc"})
	var excludeKeyStringRegex = parser.StringList("E", "exclude-regex", &argparse.Options{Required: false, Help: "Do not export Key using Regex"})
	var port = parser.Int("p", "port", &argparse.Options{Required: false, Help: "Port Number", Default: 8081})
	var host = parser.String("H", "host", &argparse.Options{Required: false, Help: "Define the host address for the web server", Default: "127.0.0.1"})
	var sampleInterval = parser.Int("i", "interval", &argparse.Options{Required: false, Help: "Sample interval in second", Default: 5})
	//var debugLevel = parser.FlagCounter("v", "verbose", &argparse.Options{Required: false, Help: "Enable debug", Default: 1})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
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

	prometheusMetrics = map[string]prometheus.GaugeVec{}

	c := make(chan string)
	go processGaugeCreation(c)

	for _, upsName := range *upsList {
		go sampleUps(c, upsName, *sampleInterval)
	}

	http.ListenAndServe(listenAddr, nil)

}

func sampleUps(c chan string, upsName string, sampleInterval int) {
	for {

		cmd := exec.Command(*upscPath, upsName)
		stdout, err := cmd.StdoutPipe()

		if err != nil {
			log.Error("Error while using upsc for ", upsName, ": error piping stdout. ", upsName, " will now be ignored ")
			return
		}
		if err := cmd.Start(); err != nil {
			log.Error("Error while using upsc for ", upsName, ": error executing command. ", upsName, " will now be ignored ")
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), ": ")

			if ignoredKey(parts[0]) {
				log.Debug("the key ", parts[0], " is ignored")
				continue
			}

			//Special case for the Status
			if parts[0] == "ups.status" {
				gauge := getGauge(parts[0])
				if gauge == nil {
					c <- parts[0]
					continue
				}
				switch parts[1] {
				case "CAL":
					gauge.WithLabelValues(upsName).Set(0)
				case "TRIM":
					gauge.WithLabelValues(upsName).Set(1)
				case "BOOST":
					gauge.WithLabelValues(upsName).Set(2)
				case "OL":
					gauge.WithLabelValues(upsName).Set(3)
				case "OB":
					gauge.WithLabelValues(upsName).Set(4)
				case "OVER":
					gauge.WithLabelValues(upsName).Set(5)
				case "LB":
					gauge.WithLabelValues(upsName).Set(6)
				case "RB":
					gauge.WithLabelValues(upsName).Set(7)
				case "BYPASS":
					gauge.WithLabelValues(upsName).Set(8)
				case "OFF":
					gauge.WithLabelValues(upsName).Set(9)
				case "CHRG":
					gauge.WithLabelValues(upsName).Set(10)
				case "DISCHRG":
					gauge.WithLabelValues(upsName).Set(11)
				}
			} else {
				//Try to convert everything to float. If not possible, drop the key...
				value, err := strconv.ParseFloat(parts[1], 64)
				if err == nil {
					gauge := getGauge(parts[0])
					if gauge == nil {
						c <- parts[0]
						continue
					}
					gauge.WithLabelValues(upsName).Set(value)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Info("Error while using upsc for ", upsName, "error scanning command output ", upsName, " will now be ignored ")
		}

		cmd.Wait() // needed to close the pipe and release the child process
		time.Sleep(time.Duration(sampleInterval) * time.Second)

	}
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

func CreateAndRegisterGauge(key string) prometheus.GaugeVec {

	key = toPrometheusKey(key)

	if val, ok := prometheusMetrics[key]; ok {
		return val
	} else {
		log.Info("Creating metric ", key)
		gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: key,
		}, []string{
			"ups",
		})

		prometheus.MustRegister(gauge)
		prometheusMetrics[key] = *gauge
		return *gauge
	}
}

func processGaugeCreation(c chan string) {
	for i := range c {
		_ = CreateAndRegisterGauge(i)
	}
}

func getGauge(key string) *prometheus.GaugeVec {
	if val, ok := prometheusMetrics[toPrometheusKey(key)]; ok {
		return &val
	}
	return nil
}
