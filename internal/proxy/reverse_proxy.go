package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type ProxyFactory interface {
	Create(host string, port int) Proxy
}

type HttpProxy struct {
	proxy *httputil.ReverseProxy
}

func (p *HttpProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.proxy.ServeHTTP(w, req)
}

type reverseProxyFactory struct{}

func NewReverseProxyFactory() ProxyFactory {
	return &reverseProxyFactory{}
}

func (f *reverseProxyFactory) Create(host string, port int) Proxy {
	return &HttpProxy{
		proxy: httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", host, port),
		}),
	}
}
