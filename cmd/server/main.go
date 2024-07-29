package main

import (
	"fmt"
	"github.com/joho/godotenv"

	"github.com/kaa-it/go-devops/internal/server"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build data: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	_ = godotenv.Load()

	config := server.NewConfig()

	s := server.New(config)

	s.Run()
}
