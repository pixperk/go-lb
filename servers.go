package lb

import (
	"fmt"
	"log"
	"net/http"
)

func StartServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello from port %d\n", port)
	})
	go func() {
		log.Printf("server running on %s\n", addr)
		log.Fatal(http.ListenAndServe(addr, mux))
	}()
}
