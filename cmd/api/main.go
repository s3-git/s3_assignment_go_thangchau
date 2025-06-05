package main

import (
	"assignment/internal/controller"
	"assignment/internal/handler"
	"assignment/internal/repository"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func main() {
	// Load config
	//cfg := config.Load()

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
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
	r.Run(":8080")
}

func initDB() (*sql.DB, error) {
	//TODO: read from config env var
	db, err := sql.Open("postgres", "host=postgres port=5432 user=postgres password=password dbname=assignment-db sslmode=disable")
	if err != nil {
		return nil, err
	}

	// Test connection
	// TODO: retry
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
