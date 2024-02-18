package database

import (
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/orchestrator/models"

	"gorm.io/gorm"
)

// Добавляет задание в БД
func AddTask(ex, req_id string, time int) {
	task := models.Task{
		Model:      gorm.Model{},
		Expression: ex,
		Req_id:     req_id,
		Status:     false,
		ToDoTime:   time,
		Res:        "",
		Err:        "",
	}

	res := db.Create(&task)
	if res.Error != nil {
		config.Log.Warn(res.Error)
	}
}

// Обновляют данные в задание
func UpdateTask(reqId, res string, status bool, err string) {
	var task models.Task
	if err := db.First(&task, "req_id = ?", reqId).Error; err != nil {
		config.Log.Error(err)
		return
	}

	if status {
		task.Status = status
	}

	if res != "" {
		task.Res = res
	}

	if err != "" {
		task.Err = err
	}

	if err := db.Save(&task).Error; err != nil {
		config.Log.Error(err)
		return
	}

}

// Выдает задание по ID
func GetTask(reqId string) (models.Task, bool) {
	var task models.Task
	if err := db.First(&task, "req_id = ?", reqId).Error; err != nil {
		config.Log.WithField("DB", "Не удалось найти значение").Warn(err)
		return task, false
	}
	return task, true
}

// Выдает все задания
func GetAllTask() ([]models.Task, bool) {
	var task []models.Task
	if err := db.Find(&task).Error; err != nil {
		config.Log.WithField("DB", "Не удалось найти значение").Warn(err)
		return task, false
	}
	return task, true
}
