package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	// Initialize the logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Set up HTTP handlers
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server",
		zap.String("port", port),
	)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed to start",
			zap.Error(err),
		)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log request start
	logger.Info("Request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Handle the request
	fmt.Fprintf(w, "Hello, World! Welcome to the AKS Go app.\n")

	// Log request end
	duration := time.Since(startTime)
	logger.Info("Request completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK),
	)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log request start
	logger.Info("Health check started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	// Return health status
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")

	// Log request end
	duration := time.Since(startTime)
	logger.Info("Health check completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", duration),
		zap.Int("status", http.StatusOK),
	)
}
