package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/database"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/services"
)

type TaskType struct {
	Task     string `json:"task"`
	AddTime  string `json:"add_time"`
	SubTime  string `json:"sub_time"`
	MultTime string `json:"mult_time"`
	DevTime  string `json:"dev_time"`
}

func Calc(w http.ResponseWriter, r *http.Request) {
	a := Answer{}
	var data TaskType

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		a.Err = err
		a.Info = "Не удалось декодировать JSON"
		w.WriteHeader(400)
		jsonResp, _ := json.Marshal(a)
		w.Write(jsonResp)
		return
	}

	reqId := services.HashSome(data.Task)

	add, _ := strconv.Atoi(fmt.Sprintf("%v", data.AddTime))
	sub, _ := strconv.Atoi(fmt.Sprintf("%v", data.SubTime))
	mult, _ := strconv.Atoi(fmt.Sprintf("%v", data.MultTime))
	dev, _ := strconv.Atoi(fmt.Sprintf("%v", data.DevTime))
	waitTime := services.GetWaitTime(data.Task, add, sub, mult, dev)

	if _, ok := database.GetTask(reqId); !ok {
		go services.Direct(data.Task, reqId, add, sub, mult, dev)
		go database.AddTask(data.Task, reqId, int(waitTime.Seconds()))
	} else {
		go database.UpdateTask(reqId, "", false, "")
	}

	a.Data = map[string]string{
		"reqID": reqId,
	}

	jsonResp, _ := json.Marshal(a)
	w.Write(jsonResp)
}
