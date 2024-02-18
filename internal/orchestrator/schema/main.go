package schema

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	_ "github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/schema/schema" // Это импорт, который содержит аннотации Swagger
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/transport"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Создает swagger

// @title           Распеределенный калькулятор (Яндекс.Лицей)
// @version         0.1
// @description     Распределенный ервис для параллельного вычисления арифметических выражений
// @contact.name   Дмитрий Солопенков
// @contact.url    https://t.me/solopenkovdmitriy
// @host      localhost:8080
// @BasePath  /

type TaskType struct {
	Task     string `json:"task"`
	AddTime  string `json:"add_time"`
	SubTime  string `json:"sub_time"`
	MultTime string `json:"mult_time"`
	DevTime  string `json:"dev_time"`
}

// @Summary AddTask
// @Tags Task
// @Accept json
// @Description Add one task
// @Param input body TaskType true "Request body in JSON format"
// @Router / [post]
func AddTask(c *gin.Context) {

	// Construct the DELETE request
	apiURL := "http://localhost:9999/"

	var requestData TaskType
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Marshal the requestData struct to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the DELETE request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data transport.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

// @Summary GetTask
// @Tags Task
// @Description Get one task
// @Param id path string true "ID" // Add a dummy parameter to make Swagger recognize the route
// @Router /task/{id} [get]
func GetTask(c *gin.Context) {
	// Send a request to your API

	taskID := c.Param("id")
	apiURL := "http://localhost:9999/task/" + taskID // Update with your actual API URL

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data transport.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

// @Summary GetAllTask
// @Tags Task
// @Description Get all task
// @Router /tasks [get]
func GetAllTask(c *gin.Context) {
	// Send a request to your API

	apiURL := "http://localhost:9999/tasks"

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with the API response
	var data transport.Answer
	json.Unmarshal(body, &data)
	c.JSON(http.StatusOK, data)
}

func Swag() {
	r := gin.New()

	r.POST("/", AddTask)
	r.GET("/task/:id", GetTask)
	r.GET("/tasks", GetAllTask)

	// Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}
