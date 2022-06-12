# Strawberry
[![Go Report Card](https://goreportcard.com/badge/github.com/berrybyte-net/strawberry)](https://goreportcard.com/report/github.com/berrybyte-net/strawberry)

Strawberry is a reverse proxy server with automatic HTTPS, written in Go. 
Designed for use for [BerryByte](https://berrybyte.net)'s backend systems.

## Build from source
Requirements:
- Go 1.18, or newer.

```console
git clone https://github.com/berrybyte-net/strawberry.git
CGO_ENABLED=0 go build -v -ldflags="-s -w" -o strawberry cmd/strawberry/main.go
```
