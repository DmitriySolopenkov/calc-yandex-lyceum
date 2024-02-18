package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/database"
	"github.com/Knetic/govaluate"
)

type AnswerData struct {
	Ex     string `json:"ex"`
	Answer string `json:"answer"`
	Err    string `json:"err"`
}

type JSONdata struct {
	Id       string        `json:"id"`
	Task     string        `json:"task"`
	WaitTime time.Duration `json:"wait_time"`
}

// Запускает горутины(воркеры)
func StartWorkers(max int, task chan []byte) {
	for i := 0; i < max; i++ {
		go func() {
			config.Log.Info("Worker start")
			var data = &JSONdata{}
			for v := range task {
				json.Unmarshal(v, data)
				config.Log.Info("Start do - " + data.Task)

				calRes, err := calculation(fmt.Sprintf("%s", data.Task))
				if err != nil {
					config.Log.Error(err)
				}

				time.Sleep(data.WaitTime)
				go database.UpdateCalRes(fmt.Sprintf("%v", data.Id), calRes.Ex, calRes.Answer, calRes.Err)

			}
		}()
	}
}

// Добавляет задание для ворекров которые были созданы в StartWorkers
func AddTask(task chan []byte) {
	for {
		keys, err := config.RedisClientQ.Keys(context.Background(), "*").Result()
		if err != nil {
			config.Log.Error(err)
			continue
		}

		for _, key := range keys {
			val := config.RedisClientQ.Get(context.Background(), key)
			if val.Err() != nil {
				config.Log.Error(val.Err())
				continue
			}
			jsonByte, err1 := val.Bytes()
			if err1 != nil {
				config.Log.Error(err)
				continue
			}

			task <- jsonByte
			err = config.RedisClientQ.Del(context.Background(), key).Err()
			if err != nil {
				config.Log.Error(err)
				continue
			}
		}
	}
}

// Вычисляет выражение
func calculation(data string) (AnswerData, error) {
	a := AnswerData{}
	expression, err := govaluate.NewEvaluableExpression(data)
	if err != nil {
		config.Log.WithField("err", "Ошибка при создании выражения").Error(err)
		a.Err = fmt.Sprintf("%v", err)
		return a, err
	}

	result, err := expression.Evaluate(nil)
	if err != nil {
		config.Log.WithField("err", "Ошибка при вычислении выражения").Error(err)
		a.Err = fmt.Sprintf("%v", err)
		return a, err
	}

	a.Ex = data
	a.Answer = fmt.Sprintf("%v", result)

	return a, nil
}

// Проверяет какие задание не выполнены и запускает их выполнение (Запускается при запуске агента)
func CheckNoReadyEx(task chan []byte) {
	var jsonData = &JSONdata{}
	if data, ok := database.GetAllCalRes(); ok {
		for _, v := range data {
			if v.Res == "" && v.Err == "" {
				jsonData.Id = v.RId
				jsonData.Task = v.Expression
				jsonData.WaitTime = time.Second * time.Duration(v.ToDoTime)

				jsonByte, err := json.Marshal(jsonData)
				if err != nil {
					config.Log.Error(err)
					continue
				}

				task <- jsonByte
			}
		}
	}
}
