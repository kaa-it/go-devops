package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaa-it/go-devops/internal/server/viewing"

	"github.com/kaa-it/go-devops/internal/server/http/rest"

	"github.com/go-chi/chi/v5"

	"github.com/kaa-it/go-devops/internal/server/storage/memory"
	"github.com/kaa-it/go-devops/internal/server/updating"
)

type Server struct {
	config *Config
}

func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() {
	log.Println("Server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	storage := memory.NewStorage()

	updater := updating.NewService(storage)
	viewer := viewing.NewService(storage)

	updatingHandler := rest.NewUpdatingHandler(updater)
	viewingHandler := rest.NewViewingHandler(viewer)

	r := chi.NewRouter()

	r.Mount("/update", updatingHandler.Route())
	r.Mount("/", viewingHandler.Route())

	server := &http.Server{
		Addr:    s.config.Server.Address,
		Handler: r,
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
