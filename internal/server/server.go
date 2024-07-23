// Package server contains implementation for metric server.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"

	serviceRest "github.com/kaa-it/go-devops/internal/server/http/rest/service"
	updatingRest "github.com/kaa-it/go-devops/internal/server/http/rest/updating"
	viewingRest "github.com/kaa-it/go-devops/internal/server/http/rest/viewing"
	"github.com/kaa-it/go-devops/internal/server/logger"
	"github.com/kaa-it/go-devops/internal/server/service"
	"github.com/kaa-it/go-devops/internal/server/storage/db"
	"github.com/kaa-it/go-devops/internal/server/storage/memory"
	"github.com/kaa-it/go-devops/internal/server/updating"
	"github.com/kaa-it/go-devops/internal/server/viewing"
)

// Server describes metric server.
type Server struct {
	config *Config
}

// New creates metric server instance.
func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}

// Run runs server and controls its lifecycle.
//
// Except metric server runs pprof server on port 7777.
func (s *Server) Run() {
	log, err := logger.New(s.config.Server.LogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("Server started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	var r *chi.Mux
	var storage *memory.Storage

	if s.config.DBStorage.DSN != "" {
		var err error
		var storage *db.Storage
		r, storage, err = s.initDB(log)
		if err != nil {
			log.Fatal(err.Error())
		}

		defer storage.Close()
	} else {
		var err error
		r, storage, err = s.initMemory(log)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	server := &http.Server{
		Addr:    s.config.Server.Address,
		Handler: r,
	}

	pprofServer := &http.Server{
		Addr: ":7777",
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := pprofServer.ListenAndServe(); err != nil {
			log.Error(fmt.Sprintf("pprof server failed: %s", err.Error()))
		}
	}()

	go func() {
		<-c
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error(err.Error())
		}

		if err := pprofServer.Shutdown(context.Background()); err != nil {
			log.Error(err.Error())
		}

		wg.Done()
	}()

	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}

	wg.Wait()

	if storage != nil {
		storage.Wait()

		if err := storage.Save(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

func (s *Server) initMemory(log *logger.Logger) (*chi.Mux, *memory.Storage, error) {
	storage, err := memory.NewStorage(&s.config.Storage)
	if err != nil {
		return nil, nil, err
	}

	updater := updating.NewService(storage)
	viewer := viewing.NewService(storage)

	updatingHandler := updatingRest.NewHandler(updater, log)
	viewingHandler := viewingRest.NewHandler(viewer, log)

	r := chi.NewRouter()

	r.Mount("/update", updatingHandler.Route(s.config.Server.Key))
	r.Mount("/", viewingHandler.Route())
	r.Mount("/updates", updatingHandler.Updates(s.config.Server.Key))

	return r, storage, nil
}

func (s *Server) initDB(log *logger.Logger) (*chi.Mux, *db.Storage, error) {
	storage, err := db.NewStorage(&s.config.DBStorage)
	if err != nil {
		return nil, nil, err
	}

	if err := storage.Initialize(context.Background()); err != nil {
		return nil, nil, err
	}

	updater := updating.NewService(storage)
	viewer := viewing.NewService(storage)
	service := service.NewService(storage)

	updatingHandler := updatingRest.NewHandler(updater, log)
	viewingHandler := viewingRest.NewHandler(viewer, log)
	serviceHandler := serviceRest.NewHandler(service, log)

	r := chi.NewRouter()

	r.Mount("/update", updatingHandler.Route(s.config.Server.Key))
	r.Mount("/", viewingHandler.Route())
	r.Mount("/ping", serviceHandler.Route())
	r.Mount("/updates", updatingHandler.Updates(s.config.Server.Key))

	return r, storage, nil
}
