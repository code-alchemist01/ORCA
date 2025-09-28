package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orca/pkg/config"
	"orca/pkg/container"
	"orca/pkg/scheduler"
	"orca/pkg/storage"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// OrcaServer represents the main orchestrator server
type OrcaServer struct {
	config           *config.Config
	logger           *logrus.Logger
	containerManager *container.Manager
	scheduler        *scheduler.Scheduler
	storage          *storage.Storage
	router           *mux.Router
}

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Konfigürasyon yüklenemedi: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		fmt.Printf("Geçersiz log seviyesi: %v\n", err)
		os.Exit(1)
	}
	logger.SetLevel(level)

	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// Create server
	server, err := NewOrcaServer(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Server oluşturulamadı")
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Server başlatılamadı")
	}
}

// NewOrcaServer creates a new Orca server
func NewOrcaServer(cfg *config.Config, logger *logrus.Logger) (*OrcaServer, error) {
	// Create container manager
	containerManager, err := container.NewManager(logger)
	if err != nil {
		return nil, fmt.Errorf("container manager oluşturulamadı: %w", err)
	}

	// Create scheduler
	sched := scheduler.NewScheduler(containerManager, logger)

	// Create storage
	store, err := storage.NewStorage(cfg.Storage.DataDir, logger)
	if err != nil {
		return nil, fmt.Errorf("storage oluşturulamadı: %w", err)
	}

	server := &OrcaServer{
		config:           cfg,
		logger:           logger,
		containerManager: containerManager,
		scheduler:        sched,
		storage:          store,
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// Start starts the Orca server
func (s *OrcaServer) Start() error {
	// Load existing deployments and services from storage
	if err := s.loadFromStorage(); err != nil {
		s.logger.WithError(err).Warn("Storage'dan veri yüklenemedi")
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		s.logger.WithField("address", addr).Info("Orca orchestrator başlatılıyor")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("HTTP server hatası")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Orca orchestrator kapatılıyor...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Server kapatma hatası")
		return err
	}

	s.logger.Info("Orca orchestrator başarıyla kapatıldı")
	return nil
}

// loadFromStorage loads deployments and services from storage
func (s *OrcaServer) loadFromStorage() error {
	// Load deployments
	deployments, err := s.storage.LoadAllDeployments()
	if err != nil {
		return fmt.Errorf("deployments yüklenemedi: %w", err)
	}

	s.logger.WithField("count", len(deployments)).Info("Deployments storage'dan yüklendi")

	// Load services
	services, err := s.storage.LoadAllServices()
	if err != nil {
		return fmt.Errorf("services yüklenemedi: %w", err)
	}

	s.logger.WithField("count", len(services)).Info("Services storage'dan yüklendi")

	return nil
}

// setupRoutes sets up HTTP routes
func (s *OrcaServer) setupRoutes() {
	s.router = mux.NewRouter()

	// Health check
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Container routes
	s.router.HandleFunc("/containers", s.listContainersHandler).Methods("GET")
	s.router.HandleFunc("/containers", s.createContainerHandler).Methods("POST")
	s.router.HandleFunc("/containers/{name}/start", s.startContainerHandler).Methods("POST")
	s.router.HandleFunc("/containers/{name}/stop", s.stopContainerHandler).Methods("POST")
	s.router.HandleFunc("/containers/{name}/remove", s.removeContainerHandler).Methods("DELETE")
	s.router.HandleFunc("/containers/{name}/logs", s.containerLogsHandler).Methods("GET")
	s.router.HandleFunc("/containers/{name}", s.getContainerHandler).Methods("GET")

	// Deployment routes
	s.router.HandleFunc("/deployments", s.listDeploymentsHandler).Methods("GET")
	s.router.HandleFunc("/deployments", s.createDeploymentHandler).Methods("POST")
	s.router.HandleFunc("/deployments/{name}", s.getDeploymentHandler).Methods("GET")
	s.router.HandleFunc("/deployments/{name}", s.deleteDeploymentHandler).Methods("DELETE")

	// Service routes
	s.router.HandleFunc("/services", s.listServicesHandler).Methods("GET")
	s.router.HandleFunc("/services", s.createServiceHandler).Methods("POST")
	s.router.HandleFunc("/services/{name}", s.getServiceHandler).Methods("GET")
	s.router.HandleFunc("/services/{name}", s.deleteServiceHandler).Methods("DELETE")

	// Stats route
	s.router.HandleFunc("/stats", s.statsHandler).Methods("GET")

	// Add logging middleware
	s.router.Use(s.loggingMiddleware)
}

// Middleware for logging requests
func (s *OrcaServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
			"remote":   r.RemoteAddr,
		}).Info("HTTP request")
	})
}