package main

import (
	"assignment/internal/config"
	"assignment/internal/controller"
	"assignment/internal/handler"
	"assignment/internal/infrastructure/database/migration"
	"assignment/internal/repository"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Run migrations
	migrationsPath := "db/migrations"
	if err := migration.RunMigrations(db, migrationsPath); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Set the global database connection for SQLBoiler
	boil.SetDB(db)

	// Initialize layers with interfaces
	repos := repository.NewRepositories(db)
	controllers := controller.NewControllers(repos)

	// Setup routes
	r := gin.Default()
	handler.SetupRoutes(r, controllers)

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

func initDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
