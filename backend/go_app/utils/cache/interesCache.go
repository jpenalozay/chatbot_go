// utils/cache/interesCache.go

package cache

import (
	"chatbot/models"
	"sync"
)

var (
	interesCache      []models.CatalogoInteres
	interesCacheMutex sync.RWMutex
)

// CargarInteresCache: carga el catálogo de intereses en memoria
func CargarInteresCache(intereses []models.CatalogoInteres) {
	interesCacheMutex.Lock()
	defer interesCacheMutex.Unlock()
	interesCache = intereses
}

// ObtenerInteresCache: devuelve una copia del catálogo de intereses en memoria
func ObtenerInteresCache() []models.CatalogoInteres {
	interesCacheMutex.RLock()
	defer interesCacheMutex.RUnlock()
	return append([]models.CatalogoInteres{}, interesCache...)
}

// ActualizarInteresCache actualiza el catálogo de intereses en memoria
func ActualizarInteresCache(intereses []models.CatalogoInteres) {
	CargarInteresCache(intereses)
}
