package algorithms

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/joaosczip/go-lb/pkg/lb"
)

type roundRobin struct {
	current int64
	targets []lb.Target
}

func NewRoundRobin(targets []lb.Target) *roundRobin {
	return &roundRobin{
		current: 0,
		targets: targets,
	}
}

func (r *roundRobin) Handle(w http.ResponseWriter, req *http.Request) {
	currentTarget := r.targets[r.current]

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", currentTarget.Host, currentTarget.Port),
	})

	if r.current == int64(len(r.targets)-1) {
		r.current = 0
	} else {
		r.current++
	}

	proxy.ServeHTTP(w, req)
}