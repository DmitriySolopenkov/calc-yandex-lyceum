package config

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var RedisClient = redis.NewClient(&redis.Options{})

// Создает подключение к Redis
func init() {

	var RedisURL = Conf.Redis_host + ":" + Conf.Redis_port

	RedisClient = redis.NewClient(&redis.Options{
		Addr: RedisURL,
		DB:   2,
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		panic("Ошибка при подключении к Redis: " + err.Error())
	}

}
