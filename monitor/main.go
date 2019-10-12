package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/shyamz-22/monitor/internal/exceptions"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const separator = ","
const configFilePath = "endpoints.txt"
const monitorInterval = 60 * time.Second

var configFile string
var interval time.Duration
var token string

func init() {
	flag.StringVar(&configFile, "f", configFilePath, "config file path")
	flag.DurationVar(&interval, "i", monitorInterval, "time interval to monitor endpoints")
	flag.StringVar(&token, "token", "", "Bearer token value for metrics endpoint authentication")
}

type endpoint struct {
	name  string
	url   string
	valid bool
}

func main() {
	var (
		endpoints []endpoint
		err       error
	)

	flag.Parse()

	logger := log.New(os.Stdout, "SLI monitor: ", log.LstdFlags|log.Lmicroseconds|log.Llongfile)

	if len(token) < 32 {
		logger.Fatalf("please configure a valid token of atleast length 32, ./monitor -token <<token value>>")
	}

	if endpoints, err = readEndPoints(logger, configFile); err != nil {
		log.Fatalf("unable to read endpoints.txt %v\n", err)
	}

	if len(endpoints) == 0 {
		logger.Fatalf("nothing to monitor, please check the config file %s", configFile)
	}

	monitor(logger, endpoints, interval)

	http.Handle("/metrics", mustAuthenticate(promhttp.Handler()))
	log.Fatal(http.ListenAndServe(":2112", nil))
}

func monitor(logger *log.Logger, endpoints []endpoint, interval time.Duration) {
	for _, endpoint := range endpoints {

		counter := promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_down", endpoint.name),
			Help: fmt.Sprintf("Number of times %s is down", endpoint.name),
		})

		recordMetrics(logger, counter, endpoint, interval)
	}
}

func recordMetrics(log *log.Logger, counter prometheus.Counter, e endpoint, interval time.Duration) {
	go func() {
		for {
			response, err := http.Get(e.url)
			if err != nil {
				log.Printf("%s server is down: %v\n", e.name, err)
				counter.Inc()
			} else {
				log.Printf("%s server is up: %d\n", e.name, response.StatusCode)
			}
			time.Sleep(interval)
		}
	}()
}

func readEndPoints(log *log.Logger, path string) ([]endpoint, error) {
	var (
		endpoints []endpoint
	)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer exceptions.LogFatalError(log, "while closing file", file.Close)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		value := scanner.Text()
		endpoint := parse(value)
		if !endpoint.valid {
			log.Printf("cannot parse name and url out of %s, skipping", value)
			continue
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, scanner.Err()
}

func parse(text string) endpoint {
	var ec endpoint
	config := strings.Split(text, separator)

	if len(config) < 2 {
		ec.valid = false
		return ec
	}
	return endpoint{
		name:  config[0],
		url:   config[1],
		valid: true,
	}
}

func mustAuthenticate(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearerToken := req.Header.Get("Authorization")
		splitToken := strings.Split(bearerToken, "Bearer ")

		if len(splitToken) != 2 || token != splitToken[1] {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, req)
	})
}
