// Package server contains the web server logic.
package server

import (
	"log"
	"net/http"

	"github.com/cobrich/netcfg-backup/storage"
	"github.com/gorilla/mux"
)

// Server holds the dependencies for the web server.
type Server struct {
	store  storage.Store
	router *mux.Router
}

// New creates a new Server instance.
func New(store storage.Store) *Server {
	s := &Server{
		store:  store,
		router: mux.NewRouter(),
	}
	s.routes() // Register the routes
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start(addr string) {
	log.Printf("Starting web server on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, s.router))
}