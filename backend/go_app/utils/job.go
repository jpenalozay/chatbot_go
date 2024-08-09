// chatbot/utils

package utils

import (
	"chatbot/logger"
	postgresUtils "chatbot/utils/db"
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var (
	ctx                     = context.Background()
	MAX_INACTIVITY_DURATION = time.Minute * 60 // Duración máxima de inactividad configurada a 2 minutos
)

// StartInactivityCheck inicia la verificación de inactividad de los usuarios
func StartInactivityCheck(db *gorm.DB, rdb *redis.Client) {
	logger.Log.Info("Iniciando el job de verificación de inactividad.")

	// Crear un nuevo scheduler de cron para ejecutar tareas periódicas
	c := cron.New()
	_, err := c.AddFunc("@every 30s", func() {
		CheckAndHandleInactiveThreads(db, rdb)
	})
	if err != nil {
		logger.Log.Fatalf("Error iniciando el job de verificación de inactividad: %v", err)
	}
	c.Start()
}

// CheckAndHandleInactiveThreads verifica y maneja los hilos inactivos almacenados en Redis
func CheckAndHandleInactiveThreads(db *gorm.DB, rdb *redis.Client) {
	currentTime := time.Now()
	sessionKeys, err := rdb.Keys(ctx, "usuario:*").Result()
	if err != nil {
		logger.Log.Fatalf("Error obteniendo claves de sesión de Redis: %v", err)
	}

	for _, sessionKey := range sessionKeys {
		sessionData, err := rdb.Get(ctx, sessionKey).Result()
		if err != nil {
			logger.Log.Errorf("Error obteniendo datos de sesión para %s: %v", sessionKey, err)
			continue
		}

		var session map[string]interface{}
		if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
			logger.Log.Errorf("Error deserializando datos de sesión para %s: %v", sessionKey, err)
			continue
		}

		// Obtener el último tiempo de actividad de la sesión
		lastActivityStr, ok := session["last_activity"].(string)
		if !ok {
			logger.Log.Infof("Campo 'last_activity' no encontrado en la sesión %s", sessionKey)
			continue
		}

		lastActivity, err := time.Parse(time.RFC3339, lastActivityStr)
		if err != nil {
			logger.Log.Errorf("Error parseando 'last_activity' para %s: %v", sessionKey, err)
			continue
		}

		// Verificar si la sesión ha estado inactiva por más del tiempo permitido
		if currentTime.Sub(lastActivity) > MAX_INACTIVITY_DURATION {
			logger.Log.Infof("Se encontró sesión sin actividad: %s", sessionKey)

			phone := sessionKey[len("usuario:"):]
			NotifyUserOfInactivity(phone)
			logger.Log.Infof("La sesion ha eliminar es %v", session)
			if err := postgresUtils.SaveOfRedisToPostgres(db, session); err != nil {
				logger.Log.Errorf("Error guardando sesión %s en Postgres: %v", sessionKey, err)
				continue
			}

			// Eliminar la sesión de Redis
			if err := rdb.Del(ctx, sessionKey).Err(); err != nil {
				logger.Log.Errorf("Error eliminando sesión %s de Redis: %v", sessionKey, err)
				continue
			}

			// Eliminar los intereses del usuario de Redis
			threadAnalyzerKey := "thread_analizer:" + session["thread_analizer"].(string)
			if err := rdb.Del(ctx, threadAnalyzerKey).Err(); err != nil {
				logger.Log.Errorf("Error eliminando los intereses del usuario %s de Redis: %v", threadAnalyzerKey, err)
				continue
			}

			logger.Log.Infof("Sesión y datos asociados eliminados correctamente para usuario %s.", phone)
		}
	}
}

// NotifyUserOfInactivity notifica al usuario sobre la inactividad de la sesión
func NotifyUserOfInactivity(phone string) {
	//message := "Tu sesión ha sido cerrada debido a inactividad. Por favor, inicia nuevamente."
	//SendMessage(phone, message)
}
