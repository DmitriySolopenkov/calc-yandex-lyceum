package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/database"

	"github.com/sirupsen/logrus"
)

type dataForReq struct {
	Id       string        `json:"id"`
	Task     string        `json:"task"`
	WaitTime time.Duration `json:"wait_time"`
}

// Перемешивает слайс
func shuffleSlice(input []string) {
	// rand.Seed(time.Now().UnixNano()) // Инициализация генератора случайных чисел с текущим временем
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	n := len(input)

	for i := n - 1; i > 0; i-- {
		// Генерация случайного индекса от 0 до i (включительно)
		j := rng.Intn(i + 1)

		// Обмен значениями между i-м и случайно выбранным индексом
		input[i], input[j] = input[j], input[i]
	}
}

// Проверка на содержание букв
func containsLetters(input string) bool {
	re := regexp.MustCompile("[a-zA-Z]")
	return re.MatchString(input)
}

// Удаление объекта из салйса по его ID
func removeFromSlice(sl []string, id int) []string {
	if id >= 0 && id < len(sl) {
		sl = append(sl[:id], sl[id+1:]...)
	} else {
		fmt.Println("Index out of bounds")
	}
	return sl
}

// Выдает рандомный сервер(агент)
func getRandomServer() (Server, error) {
	keys, err := config.RedisClient.Keys(context.Background(), "*").Result()
	if err != nil {
		config.Log.Error(err)
		return Server{}, err
	}

	shuffleSlice(keys)

	for _, key := range keys {
		info := config.RedisClient.Get(context.Background(), key)

		dataByte, _ := info.Bytes()

		var Serv Server
		json.Unmarshal(dataByte, &Serv)

		if Serv.Status == 1 {
			return Serv, nil
		}
	}

	return Server{}, errors.New("нету доступных серверов")
}

// Отправляет к серверу(агенту) запрос на выполнение
func requestToCalculation(serv Server, subst string, add, sub, mult, div int) (map[string]interface{}, error) {
	var data dataForReq

	data.Id = HashSome(subst)
	data.Task = subst
	data.WaitTime = GetWaitTime(subst, add, sub, mult, div)

	jsonData, err := json.Marshal(data)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	req, err := http.NewRequest("POST", serv.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}
	defer resp.Body.Close()

	var answer map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&answer)
	if err != nil {
		config.Log.Error(err)
		return map[string]interface{}{}, err
	}

	return answer, nil
}

// Забирает результаты из БД и меняет структуру выражения
func takeCalRes(ids []string, val string) ([]string, string, string) {
	for {
		if len(ids) == 0 {
			break
		}
		for i, vid := range ids {
			if calRes, ok := database.GetCalRes(vid); ok && calRes.Res != "" {
				if calRes.Err != "" {
					return []string{}, "", calRes.Err
				}
				val = strings.Replace(val, calRes.Expression, calRes.Res, 1)
				ids = removeFromSlice(ids, i)
				config.Log.WithFields(logrus.Fields{
					"val":        val,
					"Expression": calRes.Expression,
					"Res":        calRes.Res,
					"AllId":      ids,
				}).Info("OK")
			}
		}
		time.Sleep(time.Second)
	}

	return ids, val, ""
}

// Выбирает подвыражения из выражения
func makeTask(val string, ids []string) ([]string, []string, error) {
	var allTask []string
	var resT []string
	var resId []string
	for _, subexpression := range findSubexpressions(val) {
		res, ok := keepSignBetweenNumbers(subexpression)
		if ok != nil {
			config.Log.Error(ok)
			return []string{}, []string{}, ok
		} else if res == "+" || res == "-" || res == "*" || res == "/" || res == "**" {
			allTask = append(allTask, subexpression)
			ids = append(ids, HashSome(subexpression))
		} else {
			resT, resId, ok = makeTask(subexpression, ids)
			if ok != nil {
				config.Log.Error(ok)
				return []string{}, []string{}, ok
			}
			allTask = append(allTask, resT...)
			ids = append(ids, resId...)
		}
	}

	return allTask, ids, nil
}

// Обрабатывает выражение
func Direct(val, id string, add, sub, mult, div int) (string, error) {
	tempVal := val

	if containsLetters(val) {
		config.Log.WithField("ex", val).Error("Выражение содержит буквы")
		go database.UpdateTask(id, "", false, "Выражение содержит буквы")
		return "", errors.New("выражение содержит буквы")
	}

	trueServ, err := getRandomServer()
	if err != nil {
		config.Log.Error(err)
		return "", err
	}

	allId := []string{}

	var errTake string
	var allTask []string
	for {
		val, err = extractAllType(val)
		if err != nil {
			config.Log.Error(err)
			return "", err
		}

		allTask, allId, _ = makeTask(val, allId)

		for _, v := range allTask {
			go requestToCalculation(trueServ, v, add, sub, mult, div)
		}

		allId, val, errTake = takeCalRes(allId, val)
		if errTake != "" {
			go database.UpdateTask(id, "", false, errTake)
			config.Log.Error(errTake)
			return "", errors.New(errTake)
		}

		val, err = removeParenthesesAroundNumbers(val)
		if err != nil {
			config.Log.Error(err)
			return "", err
		}

		res, ok := keepSignBetweenNumbers(val)
		if ok != nil {
			config.Log.Error(ok)
			return "", ok
		} else if res == "+" || res == "-" || res == "*" || res == "/" || res == "**" {
			go requestToCalculation(trueServ, val, add, sub, mult, div)
			allId = append(allId, HashSome(val))
			_, val, errTake = takeCalRes(allId, val)
			if errTake != "" {
				go database.UpdateTask(id, "", false, errTake)
				config.Log.Error(errTake)
				return "", errors.New(errTake)
			}
			go database.UpdateTask(id, val, true, "")
			config.Log.Info("Готово - ", tempVal)
			return val, nil
		} else if res == "" || strings.Replace(res, "*", "", 1) == "-e" || strings.Replace(res, "*", "", 1) == "e" || strings.Replace(res, "*", "", 1) == "e-" || strings.Replace(res, "*", "", 1) == "+e" || strings.Replace(res, "*", "", 1) == "e+" {
			go database.UpdateTask(id, val, true, "")
			config.Log.Info("Готово - ", tempVal)
			return val, nil
		}
	}
}

// Ищет подвыражения в скобках
func findSubexpressions(expression string) []string {
	var subexpressions []string
	stack := 0
	start := -1

	for i, char := range expression {
		if char == '(' {
			if stack == 0 {
				start = i
			}
			stack++
		} else if char == ')' {
			stack--
			if stack == 0 && start != -1 {
				subexpression := expression[start+1 : i]
				subexpressions = append(subexpressions, subexpression)
				start = -1
			}
		}
	}

	return subexpressions
}

// Выдет время ожидания для выражения
func GetWaitTime(val string, add, sub, mult, dev int) time.Duration {
	res := 0
	res += strings.Count(val, "+") * add
	res += strings.Count(val, "-") * sub
	res += strings.Count(val, "*") * mult
	res += strings.Count(val, "/") * dev

	return time.Duration(res) * time.Second
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (умножение, деление)
func extractMultAndDev(expression string) (string, error) {
	// Используем регулярное выражение для поиска сумм и разностей
	re := regexp.MustCompile(`(\d+)\s*[\*\/]\s*(\d+)`)
	matches := re.FindAllString(expression, -1)

	// Заменяем найденные суммы и разности на скобки
	for _, match := range matches {
		expression = strings.Replace(expression, match, "("+match+")", 1)
	}

	return expression, nil
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (сложение, вычитание)
func extractAddAndSub(expression string) (string, error) {
	// Используем регулярное выражение для поиска сумм и разностей
	re := regexp.MustCompile(`(\d+)\s*[\+\-]\s*(\d+)`)
	matches := re.FindAllString(expression, -1)

	// Заменяем найденные суммы и разности на скобки
	for _, match := range matches {
		expression = strings.Replace(expression, match, "("+match+")", 1)
	}

	return expression, nil
}

// Функция для выделения частей математического выражения в скобки в порядке выполнения (степень)
func extractDegree(expression string) (string, error) {
	// Используем регулярное выражение для поиска степеней
	re := regexp.MustCompile(`(\d+)\s*\*\*\s*(\d+)`)
	matches := re.FindAllString(expression, -1)

	// Заменяем найденные степени на скобки
	for _, match := range matches {
		expression = strings.Replace(expression, match, "("+match+")", 1)
	}

	return expression, nil
}

// Выделяет все части мат выражения
func extractAllType(expression string) (string, error) {
	expression, err := extractDegree(expression)
	if err != nil {
		return "", err
	}
	expression, err = extractMultAndDev(expression)
	if err != nil {
		return "", err
	}
	expression, err = extractAddAndSub(expression)
	if err != nil {
		return "", err
	}

	return expression, nil
}

// Функция для удаления скобок вокруг чисел
func removeParenthesesAroundNumbers(expression string) (string, error) {
	// Используем регулярное выражение для поиска скобок вокруг чисел
	re := regexp.MustCompile(`\((\-?\d+)\)`)
	matches := re.FindAllStringSubmatch(expression, -1)

	// Удаляем найденные скобки вокруг чисел
	for _, match := range matches {
		expression = strings.Replace(expression, match[0], match[1], 1)
	}

	return expression, nil
}

// Функция для удаления всего, кроме знака между двумя числами
func keepSignBetweenNumbers(expression string) (string, error) {
	// Используем регулярное выражение для замены всего, кроме знаков между числами
	re := regexp.MustCompile(`-?\d+(\.\d+)?`)
	res := re.ReplaceAllString(expression, "$2")

	re = regexp.MustCompile(`[()]`)
	res = re.ReplaceAllString(res, "")

	return strings.Replace(res, " ", "", -1), nil
}

// Проверяет какие задание не выполнены и запускает их выполнение (Запускается при запуске)
func CheckNoReadyEx() {
	if data, ok := database.GetAllTask(); ok {
		for _, v := range data {
			if v.Res == "" && v.Err == "" {
				config.Log.Info("Начата обработка не завершённой зодачий - " + v.Expression)
				go Direct(v.Expression, v.Req_id, 0, 0, 0, 0)
			}
		}
	}
}
