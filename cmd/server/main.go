// @title           python-executor API
// @version         1.0
// @description     Remote Python code execution service.
// @description
// @description     IMPORTANT: Use the client libraries instead of implementing this API directly.
// @description     The API uses multipart/form-data with tar archives, which is complex to implement correctly.
// @description
// @description     Python: pip install git+https://github.com/geraldthewes/python-executor.git#subdirectory=python
// @description     Go: go get github.com/geraldthewes/python-executor/pkg/client

// @contact.name   python-executor
// @contact.url    https://github.com/geraldthewes/python-executor

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geraldthewes/python-executor/internal/api"
	"github.com/geraldthewes/python-executor/internal/config"
	"github.com/geraldthewes/python-executor/internal/executor"
	"github.com/geraldthewes/python-executor/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Server.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.WithFields(logrus.Fields{
		"host":     cfg.Server.Host,
		"port":     cfg.Server.Port,
		"log_level": cfg.Server.LogLevel,
	}).Info("Starting python-executor server")

	// Initialize storage
	var store storage.Storage
	if cfg.Consul.Enabled {
		logger.Info("Using Consul storage")
		consulStore, err := storage.NewConsulStorage(
			cfg.Consul.Address,
			cfg.Consul.Token,
			cfg.Consul.KeyPrefix,
		)
		if err != nil {
			logger.WithError(err).Warn("Failed to connect to Consul, falling back to in-memory storage")
			store = storage.NewMemoryStorage()
		} else {
			store = consulStore
		}
	} else {
		logger.Info("Using in-memory storage")
		store = storage.NewMemoryStorage()
	}
	defer store.Close()

	// Initialize executor
	exec, err := executor.NewDockerExecutor(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create executor")
	}
	defer exec.Close()

	// Create API server
	apiServer := api.NewServer(store, exec, cfg)
	router := api.SetupRouter(apiServer, logger)

	// Start cleanup routine
	go runCleanup(store, cfg.Cleanup.TTL, logger)

	// Start HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.WithField("addr", addr).Info("Server listening")

	// Graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

// runCleanup periodically cleans up old executions
func runCleanup(store storage.Storage, ttl time.Duration, logger *logrus.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		logger.Info("Running cleanup")
		if err := store.Cleanup(context.Background(), ttl); err != nil {
			logger.WithError(err).Error("Cleanup failed")
		}
	}
}
