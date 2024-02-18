package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/database"

	"github.com/gorilla/mux"
)

func GetOneTask(w http.ResponseWriter, r *http.Request) {
	var a Answer

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		a.Err = errors.New("id не обнаружен")
		a.Info = "/task/{id}, то как должен выглядеть путь"
		w.WriteHeader(400)
		jsonResp, _ := json.Marshal(a)
		w.Write(jsonResp)
	}

	if task, ok := database.GetTask(id); ok {
		a.Data = task
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(a)
		w.Write(jsonResp)
		return
	}

	a.Info = "Не удалось найти запись"
	w.WriteHeader(400)
	jsonResp, _ := json.Marshal(a)
	w.Write(jsonResp)

}

func GetAllTask(w http.ResponseWriter, r *http.Request) {
	var a Answer

	if task, ok := database.GetAllTask(); ok {
		a.Data = task
		w.WriteHeader(200)
		jsonResp, _ := json.Marshal(a)
		w.Write(jsonResp)
		return
	}

	a.Info = "Нет записей"
	w.WriteHeader(400)
	jsonResp, _ := json.Marshal(a)
	w.Write(jsonResp)

}
