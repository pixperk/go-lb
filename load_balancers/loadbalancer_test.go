package loadbalancers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

type MockServer struct {
	server       *httptest.Server
	requestCount int
	mutex        sync.Mutex
}

func NewMockServer(id int) *MockServer {
	ms := &MockServer{}
	ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms.mutex.Lock()
		ms.requestCount++
		count := ms.requestCount
		ms.mutex.Unlock()

		time.Sleep(10 * time.Millisecond)
		fmt.Fprintf(w, "Server %d - Request #%d", id, count)
	}))
	return ms
}

func (ms *MockServer) GetRequestCount() int {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	return ms.requestCount
}

func (ms *MockServer) Close() {
	ms.server.Close()
}

func (ms *MockServer) URL() *url.URL {
	url, _ := url.Parse(ms.server.URL)
	return url
}

func TestRoundRobin(t *testing.T) {
	servers := make([]*MockServer, 3)
	urls := make([]*url.URL, 3)

	for i := 0; i < 3; i++ {
		servers[i] = NewMockServer(i)
		urls[i] = servers[i].URL()
		defer servers[i].Close()
	}

	rr := NewRoundRobin(urls)

	serverHits := make(map[string]int)

	for i := 0; i < 30; i++ {
		server := rr.NextServer()
		if server != nil {
			serverHits[server.String()]++
		}
	}

	for serverURL, hits := range serverHits {
		if hits != 10 {
			t.Errorf("Round Robin failed: Server %s got %d requests, expected 10", serverURL, hits)
		}
	}

	t.Logf("Round Robin test passed. Distribution: %v", serverHits)
}

func TestLeastConnections(t *testing.T) {
	servers := make([]*MockServer, 3)
	urls := make([]*url.URL, 3)

	for i := 0; i < 3; i++ {
		servers[i] = NewMockServer(i)
		urls[i] = servers[i].URL()
		defer servers[i].Close()
	}
	lc := NewLeastConnections(urls)

	for i, server := range lc.Servers {
		if server.GetConnections() != 0 {
			t.Errorf("Server %d should start with 0 connections, got %d", i, server.GetConnections())
		}
	}

	lc.Servers[0].IncrementConnections()
	lc.Servers[0].IncrementConnections()
	lc.Servers[1].IncrementConnections()

	nextServer := lc.NextServer()
	expectedURL := lc.Servers[2].URL.String()

	if nextServer.String() != expectedURL {
		t.Errorf("Least Connections failed: Expected server %s, got %s", expectedURL, nextServer.String())
	}

	t.Logf("Least Connections test passed. Selected server with 0 connections.")
}

func TestRoundRobinConcurrency(t *testing.T) {
	urls := make([]*url.URL, 3)
	for i := 0; i < 3; i++ {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", 8000+i))
		urls[i] = url
	}

	rr := NewRoundRobin(urls)

	var wg sync.WaitGroup
	serverHits := make(map[string]int)
	var mutex sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server := rr.NextServer()
			if server != nil {
				mutex.Lock()
				serverHits[server.String()]++
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	totalHits := 0
	for _, hits := range serverHits {
		totalHits += hits
	}

	if totalHits != 100 {
		t.Errorf("Concurrent Round Robin failed: Total hits %d, expected 100", totalHits)
	}

	expectedPerServer := 100 / 3
	for serverURL, hits := range serverHits {
		if hits < expectedPerServer-5 || hits > expectedPerServer+5 {
			t.Errorf("Concurrent Round Robin distribution issue: Server %s got %d requests, expected around %d", serverURL, hits, expectedPerServer)
		}
	}

	t.Logf("Concurrent Round Robin test passed. Distribution: %v", serverHits)
}

func TestLeastConnectionsConcurrency(t *testing.T) {
	urls := make([]*url.URL, 3)
	for i := 0; i < 3; i++ {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", 8000+i))
		urls[i] = url
	}

	lc := NewLeastConnections(urls)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(serverIdx int) {
			defer wg.Done()
			server := lc.Servers[serverIdx%3]
			server.IncrementConnections()
			time.Sleep(1 * time.Millisecond)
			server.DecrementConnections()
		}(i)
	}

	wg.Wait()

	for i, server := range lc.Servers {
		connections := server.GetConnections()
		if connections != 0 {
			t.Errorf("Server %d should have 0 connections after test, got %d", i, connections)
		}
	}

	t.Logf("Concurrent Least Connections test passed. All servers back to 0 connections.")
}

func BenchmarkRoundRobin(b *testing.B) {
	urls := make([]*url.URL, 10)
	for i := 0; i < 10; i++ {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", 8000+i))
		urls[i] = url
	}

	rr := NewRoundRobin(urls)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr.NextServer()
		}
	})
}

func BenchmarkLeastConnections(b *testing.B) {
	urls := make([]*url.URL, 10)
	for i := 0; i < 10; i++ {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", 8000+i))
		urls[i] = url
	}

	lc := NewLeastConnections(urls)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lc.NextServer()
		}
	})
}

func TestEdgeCases(t *testing.T) {

	t.Run("No servers - Round Robin", func(t *testing.T) {
		rr := NewRoundRobin([]*url.URL{})
		server := rr.NextServer()
		if server != nil {
			t.Error("Expected nil for empty server list")
		}
	})

	t.Run("No servers - Least Connections", func(t *testing.T) {
		lc := NewLeastConnections([]*url.URL{})
		server := lc.NextServer()
		if server != nil {
			t.Error("Expected nil for empty server list")
		}
	})

	t.Run("Single server - Round Robin", func(t *testing.T) {
		serverURL, _ := url.Parse("http://localhost:8000")
		rr := NewRoundRobin([]*url.URL{serverURL})

		for i := 0; i < 10; i++ {
			server := rr.NextServer()
			if server.String() != serverURL.String() {
				t.Error("Single server should always be returned")
			}
		}
	})

	t.Run("Single server - Least Connections", func(t *testing.T) {
		serverURL, _ := url.Parse("http://localhost:8000")
		lc := NewLeastConnections([]*url.URL{serverURL})

		for i := 0; i < 10; i++ {
			server := lc.NextServer()
			if server.String() != serverURL.String() {
				t.Error("Single server should always be returned")
			}
		}
	})
}
