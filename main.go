package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethien-salinas/go-gorm-rest-api/db"
	"github.com/ethien-salinas/go-gorm-rest-api/models"
	"github.com/ethien-salinas/go-gorm-rest-api/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Archivo .env no encontrado, usando variables de entorno del sistema")
	}

	db.DBConnection()
	db.DB.AutoMigrate(&models.User{}, &models.Task{})

	r := mux.NewRouter()
	r.HandleFunc("/", routes.HomeHandler).Methods("GET")

	r.HandleFunc("/users", routes.GetUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", routes.GetUserHandler).Methods("GET")
	r.HandleFunc("/users", routes.PostUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", routes.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", routes.DeleteUserHandler).Methods("DELETE")

	r.HandleFunc("/tasks", routes.GetTasksHandler).Methods("GET")
	r.HandleFunc("/tasks/{id}", routes.GetTaskHandler).Methods("GET")
	r.HandleFunc("/tasks", routes.PostTaskHandler).Methods("POST")
	r.HandleFunc("/tasks/{id}", routes.UpdateTaskHandler).Methods("PUT")
	r.HandleFunc("/tasks/{id}", routes.DeleteTaskHandler).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error al iniciar el servidor:", err)
		}
	}()
	log.Println("Servidor corriendo en el puerto", port)

	<-quit
	log.Println("Apagando el servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Error al apagar el servidor:", err)
	}
	log.Println("Servidor apagado correctamente")
}
