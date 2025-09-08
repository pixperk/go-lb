# Load Balancer

A high-performance HTTP load balancer implementation in Go featuring multiple load balancing algorithms with comprehensive testing and benchmarking.

## Features

- **Multiple Load Balancing Algorithms**
  - Round Robin (RR): Distributes requests evenly across all servers
  - Least Connections (LC): Routes requests to the server with the fewest active connections

- **Thread-Safe Operations**: All algorithms are designed for concurrent access with proper synchronization

- **HTTP Reverse Proxy**: Built-in reverse proxy functionality using Go's `httputil.ReverseProxy`

- **Connection Tracking**: Real-time monitoring of active connections per server for the Least Connections algorithm

- **Comprehensive Testing**: Full test suite including unit tests, concurrency tests, and benchmarks

## Installation

```bash
git clone https://github.com/pixperk/lb.git
cd lb
go mod tidy
```

## Quick Start

### Starting Backend Servers

```bash
go run cmd/servers/main.go
```

This starts 10 backend servers on ports 8000-8009.

### Running the Load Balancer

```bash
# Round Robin Load Balancer
go run cmd/lb/main.go

# The load balancer will start on port 8080
# Edit cmd/lb/main.go to change algorithm from "rr" to "lc"
```

## Usage

### Basic Implementation

```go
package main

import (
    "net/url"
    "github.com/pixperk/lb"
)

func main() {
    // Define your backend servers
    servers := []*url.URL{
        mustParse("http://localhost:8001"),
        mustParse("http://localhost:8002"),
        mustParse("http://localhost:8003"),
    }

    // Create load balancer with desired algorithm
    lb := lb.NewLoadBalancer("rr", servers)  // "rr" for Round Robin
    // lb := lb.NewLoadBalancer("lc", servers)  // "lc" for Least Connections

    // Use as HTTP handler
    http.HandleFunc("/", lb.Handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func mustParse(rawURL string) *url.URL {
    parsed, err := url.Parse(rawURL)
    if err != nil {
        panic(err)
    }
    return parsed
}
```

### Direct Algorithm Usage

```go
// Round Robin
rr := loadbalancers.NewRoundRobin(servers)
nextServer := rr.NextServer()

// Least Connections
lc := loadbalancers.NewLeastConnections(servers)
nextServer := lc.NextServer()
```

## Load Balancing Algorithms

### Round Robin (RR)

Distributes incoming requests sequentially across all available servers. Each server receives an equal number of requests over time.

**Characteristics:**
- Simple and efficient
- Equal distribution of load
- No server state tracking required
- Best for servers with similar capacity

### Least Connections (LC)

Routes requests to the server with the fewest active connections. Tracks connection counts in real-time.

**Characteristics:**
- Dynamic load distribution
- Better for varying request processing times
- Real-time connection tracking
- Optimal for heterogeneous server environments

## Testing

Run the complete test suite:

```bash
# Run all tests
go test ./load_balancers -v

# Run benchmarks
go test ./load_balancers -bench=.

# Run specific test
go test ./load_balancers -run TestRoundRobin
```

### Test Coverage

- Algorithm correctness verification
- Concurrent access safety
- Edge case handling (empty server lists, single server)
- Performance benchmarking
- Load distribution validation

## Performance

Based on benchmarks, both algorithms provide excellent performance:

- **Round Robin**: Extremely fast with minimal overhead
- **Least Connections**: Slightly higher overhead due to connection tracking, but provides better load distribution for varying workloads

## Project Structure

```
.
├── cmd/
│   ├── lb/          # Load balancer executable
│   └── servers/     # Backend server simulator
├── load_balancers/
│   ├── rr.go        # Round Robin implementation
│   ├── lc.go        # Least Connections implementation
│   └── loadbalancer_test.go  # Comprehensive test suite
├── lb.go            # Main load balancer factory
├── servers.go       # Server utilities
└── go.mod
```

## Requirements

- Go 1.23.6 or later
- No external dependencies for core functionality
