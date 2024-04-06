package server

import (
	"context"
	updatingRest "github.com/kaa-it/go-devops/internal/server/http/rest/updating"
	viewingRest "github.com/kaa-it/go-devops/internal/server/http/rest/viewing"
	"github.com/kaa-it/go-devops/internal/server/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaa-it/go-devops/internal/server/viewing"

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
	log, err := logger.New(s.config.Server.LogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("Server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	storage := memory.NewStorage()

	updater := updating.NewService(storage)
	viewer := viewing.NewService(storage)

	updatingHandler := updatingRest.NewHandler(updater, log)
	viewingHandler := viewingRest.NewHandler(viewer, log)

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
			log.Errorw(err.Error())
		}
	}()

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalw(err.Error())
	}
}
