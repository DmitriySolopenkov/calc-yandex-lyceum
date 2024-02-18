package config

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var RedisClientQ = redis.NewClient(&redis.Options{})

// Создает подключение к Redis
func init() {

	var RedisURL = Conf.Redis_host + ":" + Conf.Redis_port

	RedisClientQ = redis.NewClient(&redis.Options{
		Addr: RedisURL,
		DB:   0,
	})

	_, err := RedisClientQ.Ping(context.Background()).Result()
	if err != nil {
		panic("Ошибка при подключении к Redis: " + err.Error())
	}
}
