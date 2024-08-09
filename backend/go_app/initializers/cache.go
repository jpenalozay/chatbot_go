// initializers/cache.go
package initializers

import (
	"chatbot/logger"
	"chatbot/models"
	"chatbot/utils/cache"
	"fmt"

	"gorm.io/gorm"
)

// InitCacheDatabase: inicializa la base de datos y carga el catálogo de intereses en caché
func InitCacheDatabase() error {
	logger.Log.Info("Iniciando inicialización de caché de base de datos...")

	// Conectar a la base de datos si no está conectado
	if err := InitPostgres(); err != nil {
		logger.Log.Errorf("Error al inicializar PostgreSQL: %v", err)
		return err
	}

	// Migrar el esquema
	if err := DB.AutoMigrate(&models.CatalogoInteres{}); err != nil {
		logger.Log.Errorf("Error al migrar el esquema de CatalogoInteres: %v", err)
		return err
	}

	// Cargar el catálogo de intereses en memoria
	if err := cargarCatalogoIntereses(DB); err != nil {
		logger.Log.Errorf("Error al cargar el catálogo de intereses en caché: %v", err)
		return err
	}

	logger.Log.Info("Caché de base de datos inicializada exitosamente.")
	return nil
}

func cargarCatalogoIntereses(db *gorm.DB) error {
	logger.Log.Info("Cargando catálogo de intereses en caché...")

	var intereses []models.CatalogoInteres
	if err := db.Find(&intereses).Error; err != nil {
		logger.Log.Errorf("Error al obtener intereses de la base de datos: %v", err)
		return err
	}

	cacheIntereses := make([]string, len(intereses))
	for i, interes := range intereses {
		cacheIntereses[i] = fmt.Sprintf("%s %s", interes.Codigo, interes.Descripcion)
	}

	cache.CargarInteresCache(intereses)
	logger.Log.Infof("Catálogo de intereses cargado en caché. Total de intereses: %d", len(intereses))
	return nil
}
