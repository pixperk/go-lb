package main

import "github.com/pixperk/lb"

func main() {
	for i := 8000; i < 8010; i++ {
		lb.StartServer(i)
	}

	select {}
}
