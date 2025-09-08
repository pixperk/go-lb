package lb

import (
	"net/http"
	"net/url"

	loadbalancers "github.com/pixperk/lb/load_balancers"
)

type LoadBalancer interface {
	NextServer() *url.URL
	Handler(w http.ResponseWriter, r *http.Request)
}

func NewLoadBalancer(strategy string, servers []*url.URL) LoadBalancer {
	switch strategy {
	case "rr":
		return loadbalancers.NewRoundRobin(servers)
	case "lc":
		return loadbalancers.NewLeastConnections(servers)
	}
	return nil
}
