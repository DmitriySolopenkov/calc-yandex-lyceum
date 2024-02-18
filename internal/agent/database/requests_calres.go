package database

import (
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/config"
	"github.com/DmitriySolopenkov/calc-yandex-lyceum/internal/agent/models"

	"gorm.io/gorm"
)

// Добавляет выполянемое задание в БД
func AddCalRes(id, ex string, time int) {
	res := models.CalRes{
		Model:      gorm.Model{},
		RId:        id,
		Expression: ex,
		Res:        "",
		Err:        "",
		ToDoTime:   time,
	}

	info := db.Create(&res)
	if info.Error != nil {
		config.Log.Warn(info.Error)
	}
}

// Обновляет данные о задании
func UpdateCalRes(id, ex, res, err string) {
	calRes := models.CalRes{}

	if err := db.First(&calRes, "r_id = ?", id).Error; err != nil {
		config.Log.WithField("DB", "Не удалось найти значение").Warn(err)
		return
	}

	if ex != "" {
		calRes.Expression = ex
	}
	if res != "" {
		calRes.Res = res
	}
	if err != "" {
		calRes.Err = err
	}

	if err := db.Save(&calRes).Error; err != nil {
		config.Log.WithField("DB", "Не удалось сохранить значение").Warn(err)
		return
	}

}

// Выдает выражение по ID
func GetCalRes(id string) (models.CalRes, bool) {
	res := models.CalRes{}

	if err := db.First(&res, "r_id = ?", id).Error; err != nil {
		config.Log.WithField("DB", "Не удалось найти значение").Warn(err)
		return res, false
	}

	return res, true
}

// Выдает все выражения
func GetAllCalRes() ([]models.CalRes, bool) {
	res := []models.CalRes{}

	if err := db.Find(&res).Error; err != nil {
		config.Log.WithField("DB", "Не удалось найти значения").Warn(err)
		return res, false
	}

	return res, true
}
