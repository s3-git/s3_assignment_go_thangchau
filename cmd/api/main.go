package main

import (
	"assignment/internal/controller"
	"assignment/internal/handler"
	"assignment/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	//cfg := config.Load()

	// Initialize database
	db := initDB(cfg)
	defer db.Close()

	// Initialize layers with interfaces
	repos := repository.NewRepositories(db)
	controllers := controller.NewControllers(repos)

	// Setup routes
	r := gin.Default()
	handler.SetupRoutes(r, controllers)

	// Start server
	r.Run(":8080")
}
