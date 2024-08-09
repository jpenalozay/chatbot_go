// go_app/utils/db/redisUtils.go
package db

import (
	"context"
	"encoding/json"
	"time"

	"chatbot/initializers"
	"chatbot/logger"

	"github.com/go-redis/redis/v8"
)

// GetRedisConn devuelve la instancia del cliente de Redis
func GetRedisConn() (*redis.Client, error) {
	return initializers.GetRedisConn()
}

// SessionKeyExists verifica si una clave de sesión existe en Redis y devuelve sus datos
func SessionKeyExists(ctx context.Context, redisConn *redis.Client, sessionKey string) (bool, string, error) {
	logger.Log.Infof("Verificando existencia de la clave de sesión: %s", sessionKey)
	result, err := redisConn.Exists(ctx, sessionKey).Result()
	if err != nil {
		logger.Log.Errorf("Error al verificar la existencia de la clave de sesión en Redis: %v", err)
		return false, "", err
	}
	if result > 0 {
		sessionDataRaw, err := redisConn.Get(ctx, sessionKey).Result()
		if err != nil {
			logger.Log.Errorf("Error al obtener los datos de sesión de Redis: %v", err)
			return false, "", err
		}
		logger.Log.Infof("La clave de sesión %s existe en Redis", sessionKey)
		return true, sessionDataRaw, nil
	}
	logger.Log.Infof("La clave de sesión %s no existe en Redis", sessionKey)
	return false, "", nil
}

// SaveMessage guarda un mensaje en Redis utilizando la conexión de Redis
func SaveMessage(ctx context.Context, redisConn *redis.Client, sessionID, waID, message, sender string) (bool, string) {
	key := "session:" + sessionID + ":messages"
	logger.Log.Infof("Guardando mensaje en la sesión: %s", key)

	isNewSession := redisConn.Exists(ctx, key).Val() == 0
	if isNewSession {
		logger.Log.Infof("Inicializando nueva sesión para wa_id: %s", waID)
	}

	messageData := map[string]string{"wa_id": waID, "message": message, "sender": sender}
	messageDataJSON, err := json.Marshal(messageData)
	if err != nil {
		errorMessage := "Error al serializar los datos del mensaje: " + err.Error()
		logger.Log.Info(errorMessage)
		return false, errorMessage
	}

	err = redisConn.RPush(ctx, key, messageDataJSON).Err()
	if err != nil {
		errorMessage := "Error al guardar el mensaje en Redis: " + err.Error()
		logger.Log.Info(errorMessage)
		return false, errorMessage
	}

	logger.Log.Infof("Mensaje de '%s' guardado bajo '%s' con datos: %v", sender, key, messageData)

	if isNewSession {
		err = redisConn.Set(ctx, "session:"+sessionID+":initialized", "true", 0).Err()
		if err != nil {
			errorMessage := "Error al inicializar nueva sesión: " + err.Error()
			logger.Log.Info(errorMessage)
			return false, errorMessage
		}
		logger.Log.Infof("Nueva sesión %s inicializada con datos de configuración.", sessionID)
	}

	return true, ""
}

// GetMessages recupera todos los mensajes asociados con un ID de WhatsApp (waID) utilizando la conexión de Redis
func GetMessages(ctx context.Context, redisConn *redis.Client, waID string) ([]map[string]string, error) {
	keyPattern := "session:" + waID + ":messages*"
	logger.Log.Infof("Recuperando mensajes con patrón: %s", keyPattern)

	sessionKeys, err := redisConn.Keys(ctx, keyPattern).Result()
	if err != nil {
		logger.Log.Errorf("Error al recuperar claves de sesión para wa_id %s: %v", waID, err)
		return nil, err
	}

	if len(sessionKeys) == 0 {
		logger.Log.Infof("No se encontraron mensajes para wa_id %s", waID)
		return []map[string]string{}, nil
	}

	var allMessages []map[string]string
	for _, sessionKey := range sessionKeys {
		messagesJSON, err := redisConn.LRange(ctx, sessionKey, 0, -1).Result()
		if err != nil {
			logger.Log.Errorf("Error al recuperar mensajes para la clave %s: %v", sessionKey, err)
			continue
		}
		for _, messageJSON := range messagesJSON {
			var messageData map[string]string
			if err := json.Unmarshal([]byte(messageJSON), &messageData); err != nil {
				logger.Log.Errorf("Error al deserializar datos del mensaje: %v", err)
				continue
			}
			allMessages = append(allMessages, messageData)
		}
	}

	logger.Log.Infof("Recuperados %d mensajes para wa_id %s", len(allMessages), waID)
	return allMessages, nil
}

// CreateSession crea una nueva sesión de usuario en Redis utilizando la conexión de Redis
func CreateSession(ctx context.Context, redisConn *redis.Client, name, phone, threadID, message, threadAnalizer string) {
	sessionKey := "usuario:" + phone
	currentTime := time.Now().Format(time.RFC3339)
	sessionData := map[string]interface{}{
		"user_info":       map[string]string{"phone": phone, "name": name},
		"thread":          threadID,
		"state":           "active",
		"thread_analizer": threadAnalizer,
		"messages":        []map[string]string{{"message": message, "sender": name, "timestamp": currentTime, "type": "incoming"}},
		"start_timestamp": currentTime,
		"last_activity":   currentTime,
	}
	sessionDataBytes, _ := json.Marshal(sessionData)
	err := redisConn.Set(ctx, sessionKey, sessionDataBytes, 0).Err()
	if err != nil {
		logger.Log.Errorf("Error al crear la sesión en Redis: %v", err)
		return
	}
	logger.Log.Infof("Sesión creada en Redis con la clave: %s", sessionKey)
}

// UpdateSession actualiza una sesión de usuario existente en Redis utilizando la conexión de Redis
func UpdateSession(ctx context.Context, redisConn *redis.Client, name, phone, message, messageType string) {
	sessionKey := "usuario:" + phone
	logger.Log.Infof("Actualizando sesión para la clave: %s", sessionKey)

	sessionDataRaw, err := redisConn.Get(ctx, sessionKey).Result()
	if err != nil {
		logger.Log.Errorf("Error al recuperar los datos de la sesión de Redis: %v", err)
		return
	}
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionDataRaw), &sessionData); err != nil {
		logger.Log.Errorf("Error al deserializar datos de la sesión: %v", err)
		return
	}
	messages := sessionData["messages"].([]interface{})
	messages = append(messages, map[string]string{
		"message":   message,
		"sender":    name,
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      messageType,
	})
	sessionData["messages"] = messages
	sessionData["last_activity"] = time.Now().Format(time.RFC3339)
	sessionDataBytes, _ := json.Marshal(sessionData)
	err = redisConn.Set(ctx, sessionKey, sessionDataBytes, 0).Err()
	if err != nil {
		logger.Log.Errorf("Error al actualizar la sesión en Redis: %v", err)
		return
	}
	logger.Log.Infof("Sesión actualizada en Redis con la clave: %s", sessionKey)
}

// UpdateUserInterest actualiza los intereses del usuario en Redis.
func UpdateUserInterest(ctx context.Context, redisConn *redis.Client, threadIDAnalizer, threadID string, userInterests []string) {
	logger.Log.Info("Actualizando intereses del usuario en Redis")
	messageKey := "thread_analizer:" + threadIDAnalizer
	currentTime := time.Now().Format(time.RFC3339)

	// Intentar obtener datos existentes
	messageDataRaw, err := redisConn.Get(ctx, messageKey).Result()
	var messageData map[string]interface{}

	if err == redis.Nil {
		logger.Log.Info("No se encontraron datos previos, creando nuevo registro")
		// Filtrar intereses vacíos
		// No hay datos previos, crear nuevo registro
		if len(userInterests) == 0 {
			logger.Log.Warn("No hay intereses para guardar, operación cancelada")
			return
		}
		messageData = map[string]interface{}{
			"thread_analizer": threadIDAnalizer,
			"thread":          threadID,
			"interests":       userInterests,
			"start_timestamp": currentTime,
			"last_activity":   currentTime,
		}
	} else if err != nil {
		logger.Log.Errorf("Error al recuperar datos del mensaje de Redis: %v", err)
		return
	} else {
		// Datos existentes encontrados, actualizar
		logger.Log.Info("Datos previos encontrados, actualizando registro")
		if err := json.Unmarshal([]byte(messageDataRaw), &messageData); err != nil {
			logger.Log.Errorf("Error al deserializar datos del mensaje: %v", err)
			return
		}

		// Combinar intereses existentes con nuevos
		existingInterests := messageData["interests"].([]interface{})
		updatedInterests := make([]string, 0)
		interestMap := make(map[string]bool)

		// Añadir intereses existentes
		for _, ei := range existingInterests {
			interest := ei.(string)
			if !interestMap[interest] {
				updatedInterests = append(updatedInterests, interest)
				interestMap[interest] = true
			}
		}

		// Añadir nuevos intereses
		for _, ni := range userInterests {
			if !interestMap[ni] {
				updatedInterests = append(updatedInterests, ni)
				interestMap[ni] = true
			}
		}

		messageData["interests"] = updatedInterests
		messageData["last_activity"] = currentTime
	}

	// Serializar y guardar datos actualizados
	messageDataBytes, err := json.Marshal(messageData)
	if err != nil {
		logger.Log.Errorf("Error al serializar datos del mensaje: %v", err)
		return
	}

	err = redisConn.Set(ctx, messageKey, messageDataBytes, 0).Err()
	if err != nil {
		logger.Log.Errorf("Error al actualizar los intereses del usuario en Redis: %v", err)
		return
	}

	logger.Log.Infof("Intereses del usuario actualizados exitosamente en Redis. Clave: %s, Total intereses: %d", messageKey, len(messageData["interests"].([]string)))
}
