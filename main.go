package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func main() {
	encoderCfg := zap.NewProductionEncoderConfig()
	// FYI, the Fluent Bit output plugin has a time_key setting
	encoderCfg.TimeKey = "real_timestamp"
	// ISO 8601 is supported by Azure Monitor Logs: https://learn.microsoft.com/en-us/kusto/query/scalar-data-types/datetime?view=azure-monitor#supported-formats
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zapcore.InfoLevel,
	).With([]zap.Field{
		// "The TimeGenerated value cannot be older than 2 days before received time or more than a day in the future."
		zap.Time("timestamp", time.Now().Add(-47*time.Hour)),
	})

	// Initialize the logger
	logger = zap.New(core)
	defer logger.Sync()

	// Set up HTTP handlers
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server",
		zap.Any("data", map[string]any{
			"port": port,
		}),
	)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		// TODO: What fields are errors associated with? Should we add an error column to the custom table schema?
		logger.Error("Server failed to start",
			zap.Error(err),
		)
		os.Exit(1)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Log request start
	logger.Info("Request started",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("data", map[string]any{
			"remote_addr": r.RemoteAddr,
			"user_agent": r.UserAgent(),
		}),
	)

	// Handle the request
	fmt.Fprintf(w, "Hello, World! Welcome to the AKS Go app.\n")

	// Log request end
	duration := time.Since(startTime)
	logger.Info("Request completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("data", map[string]any{
			"duration_ms": duration.Milliseconds(),
			"status":      http.StatusOK,
		}),
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

	// Test other log levels, verify they appear in the logs
	logger.Warn("Health check completed",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Any("data", map[string]any{
			"duration_ms": duration.Milliseconds(),
			"status":      http.StatusOK,
			"actual_time":  time.Now().Format(time.RFC3339),
		}),
	)
}
