package db

import (
	"chatbot/logger"
	"chatbot/models"
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SaveOfRedisToPostgres guarda los datos de Redis en la base de datos relacional
func SaveOfRedisToPostgres(db *gorm.DB, sessionData map[string]interface{}) error {
	userInfo := sessionData["user_info"].(map[string]interface{})
	var usuario models.UsuarioChat

	// Intentar obtener el usuario de la base de datos
	if err := db.Where("telefono = ?", userInfo["phone"]).First(&usuario).Error; err != nil {
		// Si no se encuentra, crear un nuevo usuario
		usuario = models.UsuarioChat{
			Telefono: userInfo["phone"].(string),
			Nombre:   userInfo["name"].(string),
			WaID:     userInfo["phone"].(string),
			Email:    "email_defecto@example.com",
			Dni:      "12345678",
		}
		if err := db.Create(&usuario).Error; err != nil {
			logger.Log.Errorf("Error al crear el usuario: %v", err)
			return err
		}
		logger.Log.Infof("Usuario %s creado exitosamente.", usuario.Telefono)
	}

	// Crear un nuevo hilo
	hilo := models.Hilo{
		UsuarioID:      usuario.ID,
		HiloOpenAI:     sessionData["thread"].(string),
		HiloAnalizador: sessionData["thread_analizer"].(string),
		EstadoHilo:     "inactivo",
		FechaInicio:    time.Now(),
		FechaFin:       time.Now(),
	}
	if err := db.Create(&hilo).Error; err != nil {
		logger.Log.Errorf("Error al crear el hilo: %v", err)
		return err
	}
	logger.Log.Infof("Hilo %s creado exitosamente para el usuario %s.", hilo.HiloOpenAI, usuario.Telefono)

	// Crear los mensajes asociados al hilo
	for _, msg := range sessionData["messages"].([]interface{}) {
		message := msg.(map[string]interface{})
		mensaje := models.Mensaje{
			HiloID:        hilo.ID,
			TextoMensaje:  message["message"].(string),
			TipoMensaje:   message["type"].(string),
			FechaCreacion: time.Now(),
		}
		if err := db.Create(&mensaje).Error; err != nil {
			logger.Log.Errorf("Error al crear el mensaje: %v", err)
			return err
		}
	}
	logger.Log.Infof("Mensajes para el hilo %s guardados exitosamente.", hilo.HiloOpenAI)

	// Recuperar y guardar los intereses desde Redis
	threadAnalizer := sessionData["thread_analizer"].(string)
	interestsKey := "thread_analizer:" + threadAnalizer

	// Conexión a Redis para obtener los intereses
	redisConn, err := GetRedisConn()
	if err != nil {
		logger.Log.Errorf("Error al conectar con Redis: %v", err)
		return err
	}
	interestsDataRaw, err := redisConn.Get(context.Background(), interestsKey).Result()
	if err != nil {
		logger.Log.Errorf("Error al obtener los intereses desde Redis: %v", err)
		return err
	}

	var interestsData map[string]interface{}
	if err := json.Unmarshal([]byte(interestsDataRaw), &interestsData); err != nil {
		logger.Log.Errorf("Error al deserializar los datos de intereses: %v", err)
		return err
	}

	interests, ok := interestsData["interests"].([]interface{})
	if !ok || len(interests) == 0 {
		logger.Log.Infof("No se encontraron intereses para el hilo %s.", hilo.HiloOpenAI)
		return nil
	}
	for _, interest := range interests {
		interes := models.Interes{
			HiloID:        hilo.ID,
			Interes:       interest.(string),
			Estado:        "inactivo",
			FechaCreacion: time.Now(),
		}
		if err := db.Create(&interes).Error; err != nil {
			logger.Log.Errorf("Error al crear el interés: %v", err)
			return err
		}
	}
	logger.Log.Infof("Intereses para el hilo %s guardados exitosamente.", hilo.HiloOpenAI)

	return nil
}
