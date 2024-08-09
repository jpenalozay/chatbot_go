package utils

import (
	"chatbot/logger"
	"chatbot/models"
	"chatbot/utils/cache"
	"strings"

	"gorm.io/gorm"
)

// ProcesarInteresesUsuario: procesa los intereses del usuario y los valida contra el caché
func ProcesarInteresesUsuario(interesesRaw string) []string {
	logger.Log.Info("Procesando intereses del usuario")
	catalogoIntereses := cache.ObtenerInteresCache()
	interesesList := strings.Split(interesesRaw, "\n")
	var interesesValidados []string

	for _, interes := range interesesList {
		interes = strings.TrimSpace(interes)
		if interes == "" {
			continue
		}
		partes := strings.SplitN(interes, " ", 2)
		if len(partes) != 2 {
			continue
		}
		codigo := partes[0]
		descripcion := strings.TrimSuffix(partes[1], ";")

		for _, catalogoItem := range catalogoIntereses {
			if catalogoItem.Codigo == codigo && catalogoItem.Descripcion == descripcion {
				interesesValidados = append(interesesValidados, interes)
				break
			}
		}
	}

	logger.Log.Infof("Intereses procesados. Total de intereses válidos: %d", len(interesesValidados))
	return interesesValidados
}

// ActualizarCatalogoIntereses: actualiza el catálogo de intereses en la base de datos y en caché
func ActualizarCatalogoIntereses(db *gorm.DB, nuevosIntereses []models.CatalogoInteres) error {
	logger.Log.Info("Actualizando catálogo de intereses")
	// Actualizar la base de datos
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.CatalogoInteres{}, "1=1").Error; err != nil {
			return err
		}
		if err := tx.Create(&nuevosIntereses).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Log.Errorf("Error al actualizar catálogo en la base de datos: %v", err)
		return err
	}

	// Actualizar el caché
	cache.ActualizarInteresCache(nuevosIntereses)
	logger.Log.Info("Catálogo de intereses actualizado exitosamente")
	return nil
}
