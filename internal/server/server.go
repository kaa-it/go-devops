package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaa-it/go-devops/internal/server/http/rest"
	"github.com/kaa-it/go-devops/internal/server/storage/memory"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Server struct{}

func New() *Server {
	return &Server{}
}

func (s *Server) Run() {
	storage := memory.NewStorage()

	updater := updating.NewService(storage)

	handler := rest.NewHandler(updater)

	log.Println("Server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler.Route(),
	}

	go func() {
		<-c
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
