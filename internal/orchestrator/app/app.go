package app

import (
	"fmt"
	"net/http"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/schema"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/services"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/transport"

	"github.com/gorilla/mux"
)

func Run() {
	defer config.RedisClient.Close()

	go schema.Swag()
	go services.CheckServer()
	go services.CheckNoReadyEx()

	router := mux.NewRouter()

	router.HandleFunc("/server/newcon", transport.Connect).Methods("POST")
	// router.HandleFunc("/server/all", transport.AllServ).Methods("GET")
	// router.HandleFunc("/server/add/{id}/{add}", transport.AddWorkerFor).Methods("POST")
	// router.HandleFunc("/server/del/{id}", transport.DeleteServer).Methods("DELETE")
	router.HandleFunc("/", transport.Calc).Methods("POST")
	router.HandleFunc("/task/{id}", transport.GetOneTask).Methods("GET")
	// router.HandleFunc("/user/{id}", transport.GetAllTaskFromUser).Methods("GET")
	router.HandleFunc("/tasks", transport.GetAllTask).Methods("GET")

	err := http.ListenAndServe(":"+config.Conf.Port, router)
	fmt.Println("Server start - " + config.Conf.Port)

	if err != nil {
		panic(err)
	}
}
