package main

import (
	"fmt"
	"log"
	"net/http"
)

func startServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from port %d\n", port)
	})
	go func() {
		log.Printf("server running on %s\n", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}

func main() {
	for i := 8000; i < 8010; i++ {
		startServer(i)
	}

	select {}
}
