package main

import (
	"flag"

	"github.com/kaa-it/go-devops/internal/server"
)

func main() {
	address := flag.String("a", ":8080", "server address as \"host:port\"")

	flag.Parse()

	s := server.New(*address)

	s.Run()
}
