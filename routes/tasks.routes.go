package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ethien-salinas/go-gorm-rest-api/db"
	"github.com/ethien-salinas/go-gorm-rest-api/models"
	"github.com/gorilla/mux"
)

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []models.Task
	result := db.DB.Preload("User").Find(&tasks)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al obtener las tareas"))
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var task models.Task
	if result := db.DB.Preload("User").First(&task, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	json.NewEncoder(w).Encode(task)
}

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	if result := db.DB.Create(&task); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al crear la tarea"))
		return
	}
	json.NewEncoder(w).Encode(task)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	params := mux.Vars(r)
	var task models.Task
	if result := db.DB.First(&task, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error al decodificar la tarea"))
		return
	}
	if result := db.DB.Save(&task); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al actualizar la tarea"))
		return
	}
	json.NewEncoder(w).Encode(task)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var task models.Task
	if result := db.DB.First(&task, params["id"]); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("tarea no encontrada"))
		return
	}
	if result := db.DB.Delete(&task); result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error al eliminar la tarea"))
		return
	}
	json.NewEncoder(w).Encode(task)
}
