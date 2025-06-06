package main

import (
	"assignment/internal/config"
	"assignment/internal/controller"
	"assignment/internal/handler"
	"assignment/internal/infrastructure/database/migration"
	"assignment/internal/repository"
	"database/sql"
	"log"

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

	// Start server
	r.Run(":" + cfg.Server.Port)
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
