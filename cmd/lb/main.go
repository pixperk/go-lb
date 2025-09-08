package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pixperk/lb"
)

func main() {
	var rawServers = make([]string, 10)
	for i := 0; i < 10; i++ {
		rawServers[i] = fmt.Sprintf("http://localhost:%d", 8000+i)
	}

	var servers []*url.URL
	for _, raw := range rawServers {
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		servers = append(servers, parsed)
	}

	lb := lb.NewLoadBalancer("lc", servers)

	http.HandleFunc("/", lb.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
