package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/database"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/services"
	"github.com/gorilla/mux"
)

// Добавление задания
func AddCal(w http.ResponseWriter, r *http.Request) {
	data := &services.JSONdata{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		config.Log.Error(err)
		w.WriteHeader(400)
		resp, _ := json.Marshal(`{"err":"Не удалось декодировать JSON"}`)
		w.Write(resp)
		return
	}

	if _, ok := database.GetCalRes(data.Id); !ok {
		go database.AddCalRes(data.Id, data.Task, int(data.WaitTime.Seconds()))

		jsonByte, errJson := json.Marshal(data)
		if errJson != nil {
			config.Log.Error(err)
			return
		}
		err = config.RedisClientQ.Set(context.Background(), data.Id, jsonByte, 0).Err()
		if err != nil {
			config.Log.Error(err)
			return
		}
		config.Log.Info("Add task - " + data.Task)
	}
}

// Добавление воркера
func AddWorkers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	maxAdd := vars["add"]
	val, _ := strconv.Atoi(maxAdd)
	go services.StartWorkers(val, config.TaskChan)
	config.Log.Info("Add workers - " + maxAdd)
}
