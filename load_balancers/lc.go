package loadbalancers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type LCServer struct {
	URL         *url.URL
	Connections int64
	mutex       sync.RWMutex
}

func (s *LCServer) IncrementConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Connections++
}

func (s *LCServer) DecrementConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Connections > 0 {
		s.Connections--
	}
}

func (s *LCServer) GetConnections() int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Connections
}

type LeastConnections struct {
	Servers []*LCServer
	mutex   sync.RWMutex
}

func NewLeastConnections(urls []*url.URL) *LeastConnections {
	servers := make([]*LCServer, len(urls))
	for i, url := range urls {
		servers[i] = &LCServer{
			URL:         url,
			Connections: 0,
		}
	}
	return &LeastConnections{
		Servers: servers,
	}
}

func (lb *LeastConnections) NextServer() *url.URL {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if len(lb.Servers) == 0 {
		return nil
	}

	leastConnServer := lb.Servers[0]
	minConnections := leastConnServer.GetConnections()

	for _, server := range lb.Servers[1:] {
		connections := server.GetConnections()
		if connections < minConnections {
			minConnections = connections
			leastConnServer = server
		}
	}

	return leastConnServer.URL
}

func (lb *LeastConnections) Handler(w http.ResponseWriter, r *http.Request) {

	targetURL := lb.NextServer()
	if targetURL == nil {
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

	lb.mutex.RLock()
	var targetServer *LCServer
	for _, server := range lb.Servers {
		if server.URL.String() == targetURL.String() {
			targetServer = server
			break
		}
	}
	lb.mutex.RUnlock()

	if targetServer == nil {
		http.Error(w, "Server not found", http.StatusInternalServerError)
		return
	}

	targetServer.IncrementConnections()

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalTransport := proxy.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	proxy.Transport = &connectionTrackingTransport{
		transport: originalTransport,
		server:    targetServer,
	}

	proxy.ServeHTTP(w, r)
}

// connectionTrackingTransport wraps RoundTripper to track connection completion
type connectionTrackingTransport struct {
	transport http.RoundTripper
	server    *LCServer
}

func (t *connectionTrackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.transport.RoundTrip(req)
	defer t.server.DecrementConnections()
	return resp, err
}
