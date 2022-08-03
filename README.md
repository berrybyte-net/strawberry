# Strawberry
[![Go Report Card](https://goreportcard.com/badge/github.com/berrybyte-net/strawberry)](https://goreportcard.com/report/github.com/berrybyte-net/strawberry)
[![GitHub](https://img.shields.io/github/license/berrybyte-net/strawberry)](https://github.com/berrybyte-net/strawberry)

Strawberry is a fast reverse proxy server with automatic HTTPS, written in Go. 
It is designed for use for [BerryByte](https://berrybyte.net)'s backend systems.

## Installation
### Build from source
Requirements:
- [Go](https://go.dev) 1.18, or newer.

Download the source code for Strawberry from GitHub.
```console
git clone https://github.com/berrybyte-net/strawberry.git
cd strawberry
```
Build the source code using [Go](https://go.dev) compiler.
```console
CGO_ENABLED=0 go build -v -ldflags="-s -w" -o strawberry cmd/strawberry/main.go
```
