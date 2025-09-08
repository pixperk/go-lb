package lb

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type LoadBalancer struct {
	Servers []*url.URL
	counter uint64
}

func (lb *LoadBalancer) NextServer() *url.URL {
	//round robin
	idx := atomic.AddUint64(&lb.counter, 1)
	return lb.Servers[int(idx)%len(lb.Servers)]
}

func (lb *LoadBalancer) Handler(w http.ResponseWriter, r *http.Request) {
	target := lb.NextServer()
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
