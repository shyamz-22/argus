package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	values, err := readLines("endpoints.txt")
	if err != nil {
		log.Fatalf("unable to read endpoints.txt %v\n", err)
	}

	for _, value := range values {
		fmt.Println("-------", value)
		nameAndEndpoint := strings.Split(value, ",")
		name := nameAndEndpoint[0]
		endpoint := nameAndEndpoint[1]

		counter := promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_down", name),
			Help: fmt.Sprintf("Number of times %s is down", name),
		})

		recordMetrics(endpoint, counter)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}

func recordMetrics(url string, counter prometheus.Counter) {
	go func() {
		for {
			response, err := http.Get(url)
			if err != nil {
				fmt.Printf("Server is down for %s: %v\n", url, err)
				counter.Inc()
			} else {
				fmt.Printf("Server is up  for %s: %d\n", url, response.StatusCode)
			}
			time.Sleep(8 * time.Second)
		}
	}()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
