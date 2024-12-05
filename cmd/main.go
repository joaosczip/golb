package main

import (
	"log"
	"net/http"

	"github.com/joaosczip/go-lb/internal/config"
	"github.com/joaosczip/go-lb/internal/proxy"
)

func main() {
	httpClient := http.DefaultClient
	proxyFactory := proxy.NewReverseProxyFactory()
	fileReader := config.NewOSFileReader()

	configLoader := config.NewConfigLoader("lb-config.yml", httpClient, proxyFactory, fileReader)
	loadBalancer, err := configLoader.Load()

	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	err = loadBalancer.ListenAndServe()

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
