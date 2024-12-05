package main

import (
	"log"
	"net/http"

	"github.com/joaosczip/go-lb/internal/config"
	"github.com/joaosczip/go-lb/internal/proxy"
	"github.com/joaosczip/go-lb/pkg/lb"
)

func main() {
	httpClient := http.DefaultClient
	proxyFactory := proxy.NewReverseProxyFactory()
	fileReader := config.NewOSFileReader()

	configLoader := config.NewConfigLoader("lb-config.yaml", httpClient, proxyFactory, fileReader)
	targetGroups, err := configLoader.Load()

	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	lb := lb.NewLoadBalancer(targetGroups)

	err = lb.ListenAndServe(":9000")

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
