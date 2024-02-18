package services

import (
	"net/http"
	"time"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
)

// Говорит оркестратору, что сервер агент доступенн
func PING(urls []string, path, ihost string) {
	for {
		for _, url := range urls {
			fullUrl := url + path

			req, err := http.NewRequest("POST", fullUrl, nil)
			if err != nil {
				config.Log.Error(err)
				continue
			}

			req.Header.Set("X-Forwarded-For", ihost)

			client := http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				config.Log.Error(err)
				continue
			}
			resp.Body.Close()
		}

		time.Sleep(time.Second * 30)
	}
}
