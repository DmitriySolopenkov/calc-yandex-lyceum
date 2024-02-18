package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/services"

	"github.com/gorilla/mux"
)

type Answer struct {
	Err  error       `json:"err"`
	Data interface{} `json:"data"`
	Info string      `json:"info"`
}

func Connect(w http.ResponseWriter, r *http.Request) {
	a := Answer{}

	fromURL := services.GetClientIP(r)
	if fromURL == "" {
		a.Err = errors.New("не удалось получить хост")
		a.Info = "Хост передается в Header X-Forwarded-For"
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	err := services.AddServer(services.HashSome(fromURL), fromURL)
	if err != nil {
		w.WriteHeader(400)
		a.Err = err
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	data, _ := json.Marshal(a)
	w.Write(data)
}

func AllServ(w http.ResponseWriter, r *http.Request) {
	res := services.AllServer()

	a := Answer{
		Err:  nil,
		Data: res,
	}

	data, _ := json.Marshal(a)
	w.Write(data)
}

func DeleteServer(w http.ResponseWriter, r *http.Request) {
	a := Answer{}

	vars := mux.Vars(r)
	servId, ok := vars["id"]
	if !ok {
		config.Log.WithField("err", "Не удалось найти id").Error(ok)
		w.WriteHeader(400)
		a.Err = errors.New("не удалось найти id")
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	info := config.RedisClient.Get(context.Background(), servId)
	if info.Err() != nil || info.Val() == "" {
		config.Log.WithField("err", "Не удалось найти сервер").Error(info.Err())
		w.WriteHeader(400)
		a.Err = info.Err()
		a.Info = "Не удалось найти сервер"
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	err := services.RemoveServerFromRedis(servId)
	if err != nil {
		config.Log.WithField("err", "Не удалось удалить").Error(err)
		w.WriteHeader(400)
		a.Err = err
		a.Info = "Не удалось удалить"
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	a.Data = servId
	a.Info = "Successful delete"

	data, _ := json.Marshal(a)
	w.Write(data)
}

func AddWorkerFor(w http.ResponseWriter, r *http.Request) {
	a := Answer{}
	var serv services.Server

	vars := mux.Vars(r)
	servId := vars["id"]
	maxAdd := vars["add"]

	info := config.RedisClient.Get(context.Background(), servId)
	if info.Err() != nil {
		config.Log.WithField("err", "Не удалось найти").Error(info.Err())
		w.WriteHeader(400)
		a.Err = info.Err()
		a.Info = "Не удалось найти"
		data, _ := json.Marshal(a)
		w.Write(data)
		return
	}

	data, _ := info.Bytes()
	err := json.Unmarshal(data, &serv)
	if err != nil {
		config.Log.WithField("err", "Не удалось декодировать данные сервера").Error(err)
		w.WriteHeader(500)
		a.Err = err
		a.Info = "Не удалось декодировать данные сервера"
		data, _ = json.Marshal(a)
		w.Write(data)
		return
	}

	fullUrl := serv.URL + "add/" + maxAdd

	req, err := http.NewRequest("POST", fullUrl, nil)
	if err != nil {
		config.Log.Error(err)
		config.Log.WithField("err", "Не удалось создать запрос").Error(err)
		w.WriteHeader(500)
		a.Err = err
		a.Info = "Не удалось создать запрос"
		data, _ = json.Marshal(a)
		w.Write(data)
		return
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error(err)
		config.Log.WithField("err", "Не удалось отправить запрос").Error(err)
		w.WriteHeader(500)
		a.Err = err
		a.Info = "Не удалось отправить запрос"
		data, _ = json.Marshal(a)
		w.Write(data)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(200)
	a.Data = maxAdd
	a.Info = "Воркеры добавлены"
	data, _ = json.Marshal(a)
	w.Write(data)
	// return

}
