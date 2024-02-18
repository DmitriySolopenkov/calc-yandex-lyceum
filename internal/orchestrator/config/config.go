package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type config struct {
	Redis_host string
	Redis_port string
	Port       string
	DbUsername string
	DbPassword string
	DbName     string
	DbHost     string
}

var Conf config

// Создает структуру config
func init() {
	relativePath := "internal/orchestrator/config/.env"
	currentDir, err := filepath.Abs(".")
	if err != nil {
		Log.WithField("err", "Ошибка при получении текущей директории").Error(err)
		return
	}
	absolutePath := filepath.Join(currentDir, relativePath)

	err = godotenv.Load(absolutePath)
	if err != nil {
		Log.Error(err)
		panic(err)
	}

	Conf.Redis_host = os.Getenv("redis_host")
	Conf.Redis_port = os.Getenv("redis_port")
	Conf.Port = os.Getenv("port")
	Conf.DbUsername = os.Getenv("db_user")
	Conf.DbPassword = os.Getenv("db_pass")
	Conf.DbName = os.Getenv("db_name")
	Conf.DbHost = os.Getenv("db_host")

	fmt.Println(Conf)
}
