// Package server contains the web server logic.
package server

import (
	"log"
	"net/http"
	"os"

	"github.com/cobrich/netcfg-backup/backups"
	"github.com/cobrich/netcfg-backup/core"
	"github.com/cobrich/netcfg-backup/storage"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server holds the dependencies for the web server.
type Server struct {
	store           storage.Store
	router          *mux.Router
	backupService   *backups.Service
	coreService     *core.BackupService
	sessionStore    *sessions.CookieStore
	isBackupRunning bool
}

// New creates a new Server instance.
func New(store storage.Store, backupService *backups.Service, coreService *core.BackupService) *Server {
	authKey := []byte(os.Getenv("SESSION_AUTH_KEY"))
	if len(authKey) == 0 {
		log.Println("Warning: SESSION_AUTH_KEY not set. Using a temporary insecure key.")
		authKey = []byte("a-very-insecure-temporary-secret")
	}

	s := &Server{
		store:         store,
		router:        mux.NewRouter(),
		backupService: backupService,
		coreService:   coreService,
		sessionStore:  sessions.NewCookieStore(authKey),
	}
	s.routes()
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start(addr string) {
	log.Printf("Starting web server on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, s.router))
}
