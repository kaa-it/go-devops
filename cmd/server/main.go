package main

import "github.com/kaa-it/go-devops/internal/server"

func main() {
	s := server.New()

	s.Run()
}
