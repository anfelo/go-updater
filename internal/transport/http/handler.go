package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Handler - app handler structure
type Handler struct {
	Router  *mux.Router
	Service Service
	Server  *http.Server
	Client  *resty.Client
}

// Service - main service interface
type Service interface {
}

// NewHandler - returns a pointer to the handler
func NewHandler() *Handler {
	h := &Handler{
		Client: resty.New(),
	}
	h.Router = mux.NewRouter()
	h.Router.Use(CorsMiddleware)
	h.Router.Use(LoggingMiddleware)
	h.Router.Use(TimeoutMiddleware)
	h.mapRoutes()

	h.Server = &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: h.Router,
	}

	return h
}

func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/alive", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "I am alive")
	}).Methods("GET")

	h.Router.HandleFunc("/api/v1/releases", h.GetReleases).Methods("GET")
	h.Router.HandleFunc("/api/v1/releases/latest", h.GetLatest).Methods("GET")
}

// Serve - Starts the server and handles shutdowns gracefully
func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Error(err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := h.Server.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("shut down gracefully")
	return nil
}
