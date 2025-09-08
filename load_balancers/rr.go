package loadbalancers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type RR struct {
	Servers []*url.URL
	counter uint64
}

func NewRoundRobin(servers []*url.URL) *RR {
	return &RR{
		Servers: servers,
		counter: 0,
	}
}

func (lb *RR) NextServer() *url.URL {
	if len(lb.Servers) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&lb.counter, 1)
	return lb.Servers[(idx-1)%uint64(len(lb.Servers))]
}

func (lb *RR) Handler(w http.ResponseWriter, r *http.Request) {
	target := lb.NextServer()
	if target == nil {
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
