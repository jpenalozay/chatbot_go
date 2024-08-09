package models

import (
	"gorm.io/gorm"
)

// Role representa la estructura de un rol en la base de datos
type Role struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"`
}

// Tabla de roles posibles
const (
	AdminRole = "admin"
	UserRole  = "user"
)
