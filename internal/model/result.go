package model

import "gorm.io/gorm"

type InspectResult struct {
	gorm.Model
	Host   string
	Port   int
	DBName string
	Table  string
}
