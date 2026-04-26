// Command api starts the HTTP server for the Go GORM REST API.
//
//	@title			Go GORM REST API
//	@version		1.0
//	@description	API REST para gestionar usuarios y tareas.
//	@contact.name	Ethien Salinas
//	@host			localhost:3000
//	@BasePath		/
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ethien-salinas/go-gorm-rest-api/docs"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/config"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/database"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/handlers"
	applogger "github.com/ethien-salinas/go-gorm-rest-api/internal/logger"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/middleware"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"github.com/ethien-salinas/go-gorm-rest-api/internal/repository"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	httpswagger "github.com/swaggo/http-swagger"
)

func main() {
	bootstrap := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := godotenv.Load(); err != nil {
		bootstrap.Warn(".env file not found, falling back to system environment variables")
	}

	cfg := config.Load()

	log, rotator, err := applogger.NewLogger(cfg.LogToFile, cfg.LogDir)
	if err != nil {
		bootstrap.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	db := database.Connect(cfg, log)
	if cfg.AutoMigrate {
		db.AutoMigrate(&models.User{}, &models.Task{})
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db, log)
	taskRepo := repository.NewTaskRepository(db, log)

	userHandler := handlers.NewUserHandler(userRepo, log)
	taskHandler := handlers.NewTaskHandler(taskRepo, log)

	r := mux.NewRouter()
	r.Use(middleware.Logging(log))

	r.PathPrefix("/swagger/").Handler(httpswagger.WrapHandler)
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/health", handlers.NewHealthHandler(sqlDB)).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/users", userHandler.GetAll).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	api.HandleFunc("/users", userHandler.Create).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.Update).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	api.HandleFunc("/tasks", taskHandler.GetAll).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.GetByID).Methods("GET")
	api.HandleFunc("/tasks", taskHandler.Create).Methods("POST")
	api.HandleFunc("/tasks/{id}", taskHandler.Update).Methods("PUT")
	api.HandleFunc("/tasks/{id}", taskHandler.Delete).Methods("DELETE")

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()
	log.Info("server started", "port", cfg.Port)

	<-quit
	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("server shutdown complete")

	if rotator != nil {
		if err := rotator.Close(); err != nil {
			log.Error("failed to close log rotator", "error", err)
		}
	}
}
