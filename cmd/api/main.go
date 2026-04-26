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
	"sync"
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
	"gorm.io/gorm"
)

func main() {
	bootstrap := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := godotenv.Load(); err != nil {
		bootstrap.Warn(".env file not found, falling back to system environment variables")
	}

	cfg := config.Load()

	// Bootstrap paralelo: logger y DB son IO independientes; se inician en paralelo
	// con sync.WaitGroup. La ganancia de latencia es mínima (disco local vs red),
	// pero el patrón es fundamental en Go: lanzar goroutines para trabajo independiente
	// y sincronizar con wg.Wait() antes de usar los resultados.
	var (
		log     *slog.Logger
		rotator *applogger.DailyRotator
		db      *gorm.DB
		logErr  error
		dbErr   error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log, rotator, logErr = applogger.NewLogger(cfg.LogToFile, cfg.LogDir)
	}()

	go func() {
		defer wg.Done()
		// database.Connect retorna error en lugar de llamar os.Exit,
		// delegando la decisión de terminar al caller (main).
		db, dbErr = database.Connect(cfg, bootstrap)
	}()

	wg.Wait() // esperar a que AMBAS goroutines terminen antes de leer las variables

	if logErr != nil {
		bootstrap.Error("failed to initialize logger", "error", logErr)
		os.Exit(1)
	}
	if dbErr != nil {
		bootstrap.Error("failed to connect to database", "error", dbErr)
		os.Exit(1)
	}

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
	statsHandler := handlers.NewStatsHandler(userRepo, taskRepo, log)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours, log)

	// AsyncLogger: envía logs al canal sin bloquear el goroutine del request.
	// El worker goroutine los escribe a slog desde su propio goroutine.
	asyncLog := middleware.NewAsyncLogger(log, 512)
	asyncLog.Start()

	r := mux.NewRouter()
	r.Use(middleware.Logging(asyncLog))

	r.PathPrefix("/swagger/").Handler(httpswagger.WrapHandler)
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/health", handlers.NewHealthHandler(sqlDB)).Methods("GET")

	r.HandleFunc("/auth/signup", authHandler.Signup).Methods("POST")
	r.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.Auth(cfg.JWTSecret))

	api.HandleFunc("/users", userHandler.GetAll).Methods("GET")
	api.HandleFunc("/users/batch", userHandler.BatchCreate).Methods("POST") // antes de /users/{id}
	api.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	api.HandleFunc("/users", userHandler.Create).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.Update).Methods("PATCH")
	api.HandleFunc("/users/{id}/password", userHandler.ChangePassword).Methods("PATCH")
	api.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	api.HandleFunc("/tasks", taskHandler.GetAll).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.GetByID).Methods("GET")
	api.HandleFunc("/tasks", taskHandler.Create).Methods("POST")
	api.HandleFunc("/tasks/{id}", taskHandler.Update).Methods("PUT")
	api.HandleFunc("/tasks/{id}", taskHandler.Delete).Methods("DELETE")

	api.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")

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

	// Orden de cierre: primero detener el servidor (ya hecho arriba),
	// luego vaciar el canal del AsyncLogger (todos los logs pendientes se procesan),
	// finalmente cerrar el archivo físico de logs.
	asyncLog.Stop()

	if rotator != nil {
		if err := rotator.Close(); err != nil {
			log.Error("failed to close log rotator", "error", err)
		}
	}
}
