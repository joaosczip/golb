# GOLB

An Application Load Balancer (ALB) written in Golang.

## Summary

This project is an ALB that exposes an HTTP Server that listens for incoming requests and forwards them to a pool of servers. The ALB supports two load balancing algorithms: `Round Robin` and `Least Response Time`.

## Features

- [x] Round Robin
- [x] Least Response Time
- [x] Health Check
- [ ] Least Connections
- [ ] IP Hashing
- [ ] Sticky Sessions

## Requirements

To run this project you need to have at least Go 1.23 installed on your machine.

```sh
$ go version
go version go1.23.2 darwin/arm64
```

## Usage

The first thing you need to do is to correctly configure the `lb-config.yml` file. This file contains the configuration for the ALB, such as the load balancing algorithm, the pool of servers, and the health check configuration.

Once your file is correctly configured, you can start the ALB by running the following command:

```sh
$ go run cmd/main.go
```

When the ALB starts to run, it will start listening for incoming requests on the port defined in the `lb-config.yml` file and also start the health check process for the pool of servers.

```sh
➜  lb git:(main) ✗ go run cmd/main.go
health check passed for target localhost:8082, 1, 1
target localhost:8082 is healthy
health check passed for target localhost:8080, 1, 1
target localhost:8080 is healthy
health check passed for target localhost:8081, 1, 1
target localhost:8081 is healthy
```

Now you can start sending requests to the ALB and it will forward them to the pool of servers:

```sh
➜  lb git:(main) ✗ curl -si -G http://localhost:9000
HTTP/1.1 200 OK
Content-Length: 21
Content-Type: text/plain
Date: Thu, 05 Dec 2024 10:31:57 GMT

hello from server 01
➜  lb git:(main) ✗ curl -si -G http://localhost:9000
HTTP/1.1 200 OK
Content-Length: 21
Content-Type: text/plain
Date: Thu, 05 Dec 2024 10:32:13 GMT

hello from server 02
```