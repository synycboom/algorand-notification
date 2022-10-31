# Algorand Notification Service
The Notification Service is a set of services that handles block polling from the Algorand indexer and provides web socket api to enable event subscriptions.

## Overview
This project consists of two services.
- **Monitor Service**: the monitor service takes care of block polling and publishing. Redis is used for a messaing channel.
- **Websocket Service**: the websocket service subscribes events from Redis and also provides the websocket api.

## Video Demo
TODO

## Architecture
TODO

## Requirements
- Docker (development)
- Docker Compose (development)
- Go1.18
- make

## Running with docker-compose (development)
The docker-compose file will run those two services with Redis, Grafana and Prometheus.
```shell
$ docker-compose up
```

## Running as a standalone service
Those two services need to be run together. Only one instance of Monitor Service is needed. Websocket Service can be scaled to multiple instances in case it has to serve many websocket connections. Default configs for both monitor and websocket services are places in the `config` folder.

### Build
```shell
$ make build
```

### Monitor Service
```shell
$ ./build/algorand-notification monitor --config ./config/monitor.yaml
```

### Websocket Service
```shell
$ ./build/algorand-notification server --config ./config/server.yaml
```

## Configuration
Both monitor and server commands accept configuration file via `--config` or `-c` flag. 
- `redis_host` and `redis_password`: Redis host/password are set to support running in Docker, so if these services are running in standalone, they need to be set correctly.
- `start_round`: is the start round for fetching blocks, and it should be set as latest as possible.
- `fetcher_rps`: defines maximum RPS for fetching blocks.

## API Usage
This section provide websocket specification for event subscription.

### General Websocket Information
- The base endpoint is: ws://localhost:8080
- The websocket server will send a ping frame every 3 minutes. If the websocket server does not receive a pong frame back from the connection within a 3 minute period, the connection will be disconnected.

### Subscribing/Unsubscribing
- The id used in the JSON payloads is an unsigned INT used as an identifier to uniquely identify the messages going back and forth.
- Available events are
  - "NEW_BLOCK"
  - "NEW_PAYMENT_TX"
  - "NEW_KEY_REGISTRATION_TX"
  - "NEW_ASSET_CONFIG_TX"
  - "NEW_ASSET_TRANSFER_TX"
  - "NEW_ASSET_FREEZE_TX"
  - "NEW_APPLICATION_CALL_TX"
  - "NEW_STATE_PROOF_TX"

#### Subscribe to events
Request:
```json
{
  "method": "SUBSCRIBE",
  "params": [
    "NEW_PAYMENT_TX",
    "NEW_KEY_REGISTRATION_TX"
  ],
  "id": 1
}
```
Response:
```json
{
  "id": 1
}
```
  
#### Unsubscribe an event
Request:
```json
{
  "method": "SUBSCRIBE",
  "params": [
    "NEW_PAYMENT_TX"
  ],
  "id": 2
}
```
Response:
```json
{
  "id": 2
}
```
## Monitoring Dashboard
The default metrics port for `Monitor Service` is `9361` and `9360` for `Websocket Service`. Data sources for Grafana are set in `dashboard/grafana_prometheus_datasource.docker.yaml`. Check that configurations for Prometheus source is correct or Grafana will not have the metrics.
After running up docker-compose, Grafana is running on http://localhost:3000; default login (admin/admin).

### View metrics on grafana
- Go to Import and upload `dashboard/go_runtime_dashboard.json` and `dashboard/server_stats_dashboard.json`
- `go_runtime_dashboard` dashboard shows general information of Golang runtime.
- `server_stats_dashboard` dashbaord shows websocket connection related information.
