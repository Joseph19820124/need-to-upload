package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/github-mcp-http/internal/handlers"
	"github.com/github-mcp-http/pkg/sse"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Host        string
	Port        int
	TLSCert     string
	TLSKey      string
	GitHubToken string
	ReadOnly    bool
}

type Server struct {
	config      *ServerConfig
	router      *mux.Router
	corsHandler http.Handler
	mcpHandler  *handlers.MCPHandler
	sseHub      *sse.Hub
	sessions    sync.Map
	logger      *logrus.Logger
}

type Session struct {
	ID         string
	Client     *sse.Client
	Context    context.Context
	Cancel     context.CancelFunc
	LastActive time.Time
}

func NewServer(config *ServerConfig) (*Server, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	mcpHandler, err := handlers.NewMCPHandler(config.GitHubToken, config.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP handler: %w", err)
	}

	s := &Server{
		config:     config,
		router:     mux.NewRouter(),
		mcpHandler: mcpHandler,
		sseHub:     sse.NewHub(),
		logger:     logger,
	}

	s.setupRoutes()
	go s.sseHub.Run()
	go s.cleanupSessions()

	return s, nil
}

func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	api.HandleFunc("/connect", s.handleConnect).Methods("POST")
	api.HandleFunc("/disconnect", s.handleDisconnect).Methods("POST")
	api.HandleFunc("/rpc", s.handleRPC).Methods("POST")
	api.HandleFunc("/events", s.handleSSE).Methods("GET")
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.authMiddleware)
	
	// Store the CORS handler separately - don't reassign to router
	s.corsHandler = c.Handler(s.router)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.corsHandler != nil {
		s.corsHandler.ServeHTTP(w, r)
	} else {
		s.router.ServeHTTP(w, r)
	}
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionID := generateSessionID()
	ctx, cancel := context.WithCancel(context.Background())
	
	session := &Session{
		ID:         sessionID,
		Context:    ctx,
		Cancel:     cancel,
		LastActive: time.Now(),
	}
	
	s.sessions.Store(sessionID, session)

	initResult, err := s.mcpHandler.Initialize(ctx, req.ClientInfo.Name, req.ClientInfo.Version)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to initialize MCP connection")
		return
	}

	response := map[string]interface{}{
		"sessionId":    sessionID,
		"serverInfo":   initResult.ServerInfo,
		"capabilities": initResult.Capabilities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	if session, ok := s.sessions.LoadAndDelete(sessionID); ok {
		sess := session.(*Session)
		sess.Cancel()
		if sess.Client != nil {
			s.sseHub.Unregister(sess.Client)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	sessionVal, exists := s.sessions.Load(sessionID)
	if !exists {
		s.writeError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	session := sessionVal.(*Session)
	session.LastActive = time.Now()

	var rpcReq json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rpcReq); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid RPC request")
		return
	}

	response, err := s.mcpHandler.ProcessRPC(session.Context, rpcReq)
	if err != nil {
		s.logger.WithError(err).Error("RPC processing failed")
		s.writeError(w, http.StatusInternalServerError, "RPC processing failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.writeError(w, http.StatusBadRequest, "Missing session ID")
		return
	}

	sessionVal, exists := s.sessions.Load(sessionID)
	if !exists {
		s.writeError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	session := sessionVal.(*Session)
	
	client := sse.NewClient(sessionID, w)
	session.Client = client
	s.sseHub.Register(client)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	client.Send(sse.Event{
		Type: "connected",
		Data: map[string]string{"sessionId": sessionID},
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-session.Context.Done():
			return
		case <-r.Context().Done():
			s.sseHub.Unregister(client)
			return
		case <-ticker.C:
			client.Send(sse.Event{Type: "ping"})
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"time":    time.Now().UTC(),
		"version": "1.0.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		
		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.statusCode,
			"duration":   time.Since(start),
			"remote_addr": r.RemoteAddr,
		}).Info("HTTP request")
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (s *Server) cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		s.sessions.Range(func(key, value interface{}) bool {
			session := value.(*Session)
			if now.Sub(session.LastActive) > 30*time.Minute {
				s.logger.WithField("sessionId", session.ID).Info("Cleaning up inactive session")
				session.Cancel()
				if session.Client != nil {
					s.sseHub.Unregister(session.Client)
				}
				s.sessions.Delete(key)
			}
			return true
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateSessionID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}