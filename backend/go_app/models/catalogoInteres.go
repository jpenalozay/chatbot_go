// models/catalogoInteres.go

package models

import (
	"gorm.io/gorm"
)

// CatalogoInteres representa la estructura de un ítem en el catálogo de intereses
type CatalogoInteres struct {
	gorm.Model
	Codigo      string `gorm:"uniqueIndex;not null"`
	Descripcion string `gorm:"not null"`
}
