# Argus
A simple application to monitor the health of applications with prometheus. Argus goes through the list of endpoints 
for a configured time interval and checks if configured servers are reachable. It then sends a metric <<server_name>>_down per endpoint configured.
This can later be used to configure an alert to indicate if the server is down.

## Security

The default exposed `/metrics` endpoint is secured with bearer token, scraping config of your prometheus server 
config `prometheus.yaml` has to be configured like:

```yaml
global:
  scrape_interval:     15s 
  external_labels:
    monitor: 'codelab-monitor'
scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']

  - job_name: myapp
    scrape_interval: 10s
    bearer_token: 73a54630bd12a603ad277b2538fc25ee #<-- 🤓 This is where your token goes
    static_configs:
      - targets: ['localhost:2112']
```

## Build

```bash
> cd monitor
> go build -race
```

## Configuration

```bash
> ./monitor --help

Usage of ./monitor:
-f string
      config file path (default "endpoints.txt")
-i duration
      time interval to monitor endpoints (default 1m0s)
-token string
      Bearer token value for metrics endpoint authentication
```

## Endpoints file

All the endpoints that needs to be monitored can be configured as `name,url`

### Example
```csv
github,http://localhost:8080
microsoft,http://localhost:8081
yahoo,http://localhost:8082
google,https://google.com
```

## Running

```bash
> ./monitor -token 7cbbe77c63834d6b52251d7dc9b7e7bc -f /home/users/name/endpoints.txt

SLI monitor: 2019/10/12 23:42:53.567567 /Users/shyamz-22/workspace/argus/monitor/main.go:85: yahoo server is down: Get http://localhost:8082: dial tcp [::1]:8082: connect: connection refused
SLI monitor: 2019/10/12 23:42:53.567651 /Users/shyamz-22/workspace/argus/monitor/main.go:85: github server is down: Get http://localhost:8080: dial tcp [::1]:8080: connect: connection refused
SLI monitor: 2019/10/12 23:42:53.567652 /Users/shyamz-22/workspace/argus/monitor/main.go:85: microsoft server is down: Get http://localhost:8081: dial tcp [::1]:8081: connect: connection refused
SLI monitor: 2019/10/12 23:42:54.089784 /Users/shyamz-22/workspace/argus/monitor/main.go:88: google server is up: 200

```