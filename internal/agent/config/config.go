package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	Worker       int
	Redis_host   string
	Redis_port   string
	Port         string
	Connect_to   []string
	Сonnect_path string
	I_host       string
	DbUsername   string
	DbPassword   string
	DbName       string
	DbHost       string
}

var Conf config
var TaskChan = make(chan []byte)

// Создает структуру config
func init() {
	relativePath := "internal/agent/config/.env"
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

	Conf.Worker, _ = strconv.Atoi(os.Getenv("worker"))
	Conf.Redis_host = os.Getenv("redis_host")
	Conf.Redis_port = os.Getenv("redis_port")
	Conf.Port = os.Getenv("port")
	Conf.I_host = os.Getenv("i_host")
	Conf.Connect_to = strings.Split(os.Getenv("connect_to"), ",")
	Conf.Сonnect_path = os.Getenv("connect_path")
	Conf.DbUsername = os.Getenv("db_user")
	Conf.DbPassword = os.Getenv("db_pass")
	Conf.DbName = os.Getenv("db_name")
	Conf.DbHost = os.Getenv("db_host")

	fmt.Println(Conf)
}
