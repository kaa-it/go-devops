package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/buildconfig"
	"github.com/kaa-it/go-devops/internal/server"
)

func main() {
	buildconfig.PrintBuildInfo()

	_ = godotenv.Load()

	config, err := server.NewConfig()
	if err != nil {
		log.Printf("failed to load config: %s", err)
		return
	}

	s, err := server.New(config)
	if err != nil {
		log.Printf("failed to create server: %s", err)
		return
	}

	s.Run()
}
