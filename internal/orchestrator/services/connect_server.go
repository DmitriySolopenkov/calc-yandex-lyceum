package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
	"golang.org/x/text/encoding/unicode"
)

type Server struct {
	Id       string    `json:"id"`
	URL      string    `json:"url"`
	Status   int       `json:"status"`
	LastPing time.Time `json:"last_ping"`
}

// Засовывает сервер(агент) в Redis
func serverToRedis(s Server) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	info1 := config.RedisClient.Set(context.Background(), s.Id, data, 0)
	if info1.Err() != nil {
		return info1.Err()
	}

	return nil
}

// Удаляет сервер(агент) из Redis
func RemoveServerFromRedis(id string) error {
	_, err := config.RedisClient.Del(context.Background(), id).Result()
	if err != nil {
		config.Log.Error(err)
		return err
	}

	return nil
}

// Создает хэш значения
func HashSome(val string) string {
	utf8Encoder := unicode.UTF8.NewEncoder()

	utf8Bytes, _ := utf8Encoder.Bytes([]byte(val))

	hasher := sha256.New()
	hasher.Write(utf8Bytes)
	hashInBytes := hasher.Sum(nil)

	hashString := fmt.Sprintf("%x", hashInBytes)

	return hashString
}

// Получает IP из запроса
func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	return ""
}

// Добавляет или обновляет подключение сервера(агента)
func AddServer(id, URL string) error {

	info := config.RedisClient.Get(context.Background(), id)

	dataByte, _ := info.Bytes()
	if len(dataByte) > 0 {
		var oldServ Server
		err := json.Unmarshal(dataByte, &oldServ)
		if err != nil {
			return err
		}

		oldServ.LastPing = time.Now()
		oldServ.Status = 1

		err = serverToRedis(oldServ)
		if err != nil {
			return err
		}

		config.Log.Info("update connect server " + URL)
		return nil
	}

	server := Server{
		Id:       id,
		URL:      URL,
		Status:   1,
		LastPing: time.Now(),
	}

	err := serverToRedis(server)
	if err != nil {
		return err
	}

	config.Log.Info("connect new server " + URL)
	return nil
}

// Выдает все сервера(агенты)
func AllServer() []Server {
	keys, err := config.RedisClient.Keys(context.Background(), "*").Result()
	if err != nil {
		config.Log.Error(err)
		return nil
	}

	allS := []Server{}
	for _, key := range keys {
		info := config.RedisClient.Get(context.Background(), key)

		dataByte, _ := info.Bytes()

		var Serv Server
		err = json.Unmarshal(dataByte, &Serv)
		if err != nil {
			config.Log.Error(err)
		}
		allS = append(allS, Serv)
	}

	return allS
}

// Проверяет последнее подключение к серверу(агенту)
func CheckServer() {
	for {
		keys, err := config.RedisClient.Keys(context.Background(), "*").Result()
		if err != nil {
			config.Log.Error(err)
			continue
		}

		for _, key := range keys {
			info := config.RedisClient.Get(context.Background(), key)

			dataByte, _ := info.Bytes()

			var Serv Server
			json.Unmarshal(dataByte, &Serv)

			if time.Now().Sub(Serv.LastPing) >= (time.Minute) {
				Serv.Status = 2
				serverToRedis(Serv)
			}
		}
		time.Sleep(time.Second * 20)
	}
}
