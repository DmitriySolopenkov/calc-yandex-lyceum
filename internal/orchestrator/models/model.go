package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Expression string `gorm:"type:varchar(500)"`
	Req_id     string `gorm:"type:varchar(65);unique"` // Хэш времени
	Status     bool   `gorm:"default:false"`
	ToDoTime   int    `gorm:"type:integer"`
	Res        string `gorm:"type:string"`
	Err        string `gorm:"type:string"`
}

type CalRes struct {
	gorm.Model
	RId        string `gorm:"type:varchar(65);unique"` // Хэш выражения
	Expression string `gorm:"type:varchar(500)"`
	Res        string `gorm:"type:string"`
	Err        string `gorm:"type:string"`
	ToDoTime   int    `gorm:"type:integer"`
}
