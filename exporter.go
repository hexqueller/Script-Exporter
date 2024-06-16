package main

import (
	"flag"
	"fmt"
)

func main() {
	port := flag.Int("p", 8080, "Port to listen on")
	config := flag.String("c", "config.json", "Path to config file")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Listening on port %d", *port))
	fmt.Println(fmt.Sprintf("Config: %s", *config))
}
