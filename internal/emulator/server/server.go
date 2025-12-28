package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robmorgan/infraspec/internal/emulator/auth"
	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/metadata"
)

type Server struct {
	httpServer     *http.Server
	router         *mux.Router
	handler        *EmulatorHandler
	authMiddleware *auth.SigV4Middleware
}

func NewServer(port int, emulatorRouter emulator.RequestRouter, keyStore auth.KeyStore, state emulator.StateManager) *Server {
	handler := NewEmulatorHandler(emulatorRouter)

	router := mux.NewRouter()

	// Create authentication middleware first (before registering routes)
	var authMiddleware *auth.SigV4Middleware
	var finalHandler http.Handler

	if keyStore != nil {
		// Authentication enabled - exempt health, services, and metadata endpoints
		authMiddleware = auth.NewSigV4Middleware(keyStore, []string{"/_health", "/_services", "/latest/"})
		finalHandler = authMiddleware.Middleware(handler)
	} else {
		// Authentication disabled
		finalHandler = handler
	}

	// Health check endpoint (exempt from authentication)
	router.HandleFunc("/_health", handler.HealthCheck).Methods("GET")

	// Services list endpoint (exempt from authentication)
	router.HandleFunc("/_services", handler.ListServices).Methods("GET")

	// EC2 metadata service endpoint (exempt from authentication)
	// CRITICAL: Must be registered BEFORE the PathPrefix("/") catch-all
	// Use a subrouter with StrictSlash to ensure proper matching
	metadataHandler := metadata.NewHandler(state)
	metadataRouter := router.PathPrefix("/latest/").Subrouter()
	metadataRouter.PathPrefix("/").Handler(metadataHandler)

	// Root status endpoint for non-AWS clients (exempt from authentication)
	router.HandleFunc("/", handler.RootStatus).Methods("GET")

	// Catch-all for AWS service emulation (MUST be last)
	router.PathPrefix("/").Handler(finalHandler)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:     httpServer,
		router:         router,
		handler:        handler,
		authMiddleware: authMiddleware,
	}
}

func (s *Server) Start() error {
	log.Printf("Starting AWS emulator server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// StartWithListener starts the server using the provided listener.
// This is useful for embedded mode where we need to control the port.
func (s *Server) StartWithListener(listener net.Listener) error {
	log.Printf("Starting AWS emulator server on %s", listener.Addr().String())
	return s.httpServer.Serve(listener)
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down AWS emulator server...")
	return s.httpServer.Shutdown(ctx)
}
