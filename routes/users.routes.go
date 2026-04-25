package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/db"
	"github.com/ethien-salinas/go-gorm-rest-api/models"
	"github.com/gorilla/mux"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users []models.User
	result := db.DB.Preload("Tasks").Find(&users)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al obtener los usuarios"))
		return
	}
	json.NewEncoder(w).Encode(users)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var user models.User
	if result := db.DB.Preload("Tasks").First(&user, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func PostUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	if result := db.DB.Create(&user); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al crear el usuario"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	var user models.User
	if result := db.DB.First(&user, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar el usuario"))
		return
	}
	if result := db.DB.Save(&user); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al actualizar el usuario"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var user models.User
	if result := db.DB.First(&user, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("usuario no encontrado"))
		return
	}
	if result := db.DB.Delete(&user); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar el usuario"))
		return
	}
	json.NewEncoder(w).Encode(user)
}
