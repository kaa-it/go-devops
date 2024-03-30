package main

import (
	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/server"
)

func main() {
	_ = godotenv.Load()

	config := server.NewConfig()

	s := server.New(config)

	s.Run()
}
