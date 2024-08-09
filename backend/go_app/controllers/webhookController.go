// go_app/controllers/webhookController.go

package controllers

import (
	"bytes"
	"chatbot/logger"
	"chatbot/utils"
	db "chatbot/utils/db"
	pb "chatbot/utils/proto"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

var ctx = context.Background()

// WebhookGet maneja las solicitudes de verificación del webhook de WhatsApp.
func WebhookGet(c *gin.Context) {
	logger.Log.Info("Recibida solicitud GET para verificación del webhook")
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == os.Getenv("VERIFY_TOKEN") {
		logger.Log.Info("Webhook verificado exitosamente")
		c.String(http.StatusOK, challenge)
	} else {
		logger.Log.Warn("Falló la verificación del webhook")
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Falló la verificación"})
	}
}

// WebhookPost maneja las solicitudes entrantes del webhook de WhatsApp.
func WebhookPost(c *gin.Context) {
	logger.Log.Info("Recibida solicitud POST para el webhook")
	var jsonBody map[string]interface{}

	body, exists := c.Get("body")
	if !exists {
		logger.Log.Warn("Cuerpo de la solicitud no encontrado en el contexto")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la solicitud no encontrado"})
		return
	}

	if err := json.Unmarshal(body.([]byte), &jsonBody); err != nil {
		logger.Log.Error("Error al deserializar JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	if utils.IsWhatsAppStatusUpdate(jsonBody) {
		logger.Log.Info("Recibida actualización de estado de WhatsApp")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	if utils.IsValidWhatsAppMessage(jsonBody) {
		logger.Log.Info("Procesando mensaje de WhatsApp")
		if err := processWhatsAppMessage(jsonBody); err != nil {
			logger.Log.Error("Error al procesar mensaje de WhatsApp:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al procesar el mensaje"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	} else {
		logger.Log.Warn("El evento recibido no es un mensaje válido de WhatsApp")
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "No es un evento de la API de WhatsApp"})
	}
}

// processWhatsAppMessage procesa el mensaje de WhatsApp.
func processWhatsAppMessage(body map[string]interface{}) error {
	logger.Log.Info("Iniciando procesamiento del mensaje de WhatsApp")

	// Extraer datos del mensaje
	phone, name, messageBody, err := extractMessageData(body)
	if err != nil {
		return fmt.Errorf("fallo al extraer datos del mensaje: %w", err)
	}

	if messageBody == "" {
		logger.Log.Warn("Recibido mensaje de WhatsApp con texto vacío")
		return nil
	}

	// Obtener conexión a Redis
	redisConn, err := db.GetRedisConn()
	if err != nil {
		return fmt.Errorf("fallo al obtener conexión a Redis: %w", err)
	}

	// Construir la clave de la sesión en Redis
	sessionKey := "usuario:" + phone
	logger.Log.Infof("Clave de sesión: %v", sessionKey)

	// Verificar si la sesión existe en Redis
	exists, sessionDataRaw, err := db.SessionKeyExists(ctx, redisConn, sessionKey)
	if err != nil {
		return fmt.Errorf("fallo al verificar la clave de sesión en Redis: %w", err)
	}

	var threadID, threadIDAnalizer string
	if !exists {
		logger.Log.Info("Usuario no encontrado, creando nuevos hilos en OpenAI")
		threadID, threadIDAnalizer, err = createNewThreads()
		if err != nil {
			return fmt.Errorf("fallo al crear nuevos hilos: %w", err)
		}
		db.CreateSession(ctx, redisConn, name, phone, threadID, messageBody, threadIDAnalizer)
		logger.Log.Info("Nueva sesión creada en Redis")
	} else {
		logger.Log.Info("Usuario encontrado en Redis")
		threadID, threadIDAnalizer = retrieveThreads(sessionDataRaw)
		db.UpdateSession(ctx, redisConn, name, phone, messageBody, "incoming")
		logger.Log.Info("Sesión actualizada en Redis")
	}

	logger.Log.Infof("Usuario %v: threadID=%v, threadIDAnalizer=%v, mensaje=%v", phone, threadID, threadIDAnalizer, messageBody)

	// Genera el interés del usuario utilizando el analizador
	userInterests, err := generateResponseAnalizer(threadIDAnalizer, messageBody)
	if err != nil {
		return fmt.Errorf("fallo al generar respuesta del analizador: %w", err)
	}

	//if len(userInterests) > 0 {s
	//	db.UpdateUserInterest(ctx, redisConn, threadIDAnalizer, threadID, userInterests)
	//	logger.Log.Info("Intereses del usuario actualizados en Redis")
	//}

	if len(userInterests) > 0 {
		db.UpdateUserInterest(ctx, redisConn, threadIDAnalizer, threadID, userInterests)
		logger.Log.Info("Proceso de actualización de intereses del usuario completado")
	} else {
		logger.Log.Info("No se encontraron intereses para actualizar en Redis")
	}

	// Genera una respuesta para el usuario
	response, followUpQuestion, options, err := generateResponse(phone, threadID, messageBody)
	if err != nil {
		return fmt.Errorf("fallo al generar respuesta: %w", err)
	}

	response = utils.ProcessTextForWhatsApp(response)
	if response == "" {
		response = "Disculpa, no pude procesar tu solicitud correctamente."
	}

	logger.Log.Infof("Respuesta generada: %s", response)

	err = updateSession(redisConn, name, phone, response, "outgoing")
	if err != nil {
		return fmt.Errorf("fallo al actualizar sesión con la respuesta: %w", err)
	}

	// Envía la respuesta principal al usuario
	err = utils.SendMessage(phone, utils.GetTextMessageInput(phone, response))
	if err != nil {
		return fmt.Errorf("fallo al enviar respuesta principal: %w", err)
	}

	// Si hay una pregunta de seguimiento, envíala con opciones
	if followUpQuestion != "" {
		logger.Log.Infof("Enviando pregunta de seguimiento: %s", followUpQuestion)
		err = utils.SendMessage(phone, utils.GetInteractiveMessageInput(phone, followUpQuestion, options))
		if err != nil {
			return fmt.Errorf("fallo al enviar pregunta de seguimiento: %w", err)
		}
	}

	logger.Log.Info("Mensaje de WhatsApp procesado exitosamente")
	return nil
}

// extractMessageData extrae los datos relevantes del mensaje de WhatsApp.
func extractMessageData(body map[string]interface{}) (string, string, string, error) {
	entry := body["entry"].([]interface{})[0].(map[string]interface{})
	changes := entry["changes"].([]interface{})[0].(map[string]interface{})
	value := changes["value"].(map[string]interface{})

	phone := value["contacts"].([]interface{})[0].(map[string]interface{})["wa_id"].(string)
	name := value["contacts"].([]interface{})[0].(map[string]interface{})["profile"].(map[string]interface{})["name"].(string)
	messageBody := value["messages"].([]interface{})[0].(map[string]interface{})["text"].(map[string]interface{})["body"].(string)

	return phone, name, messageBody, nil
}

// createNewThreads crea nuevos hilos para el usuario y el analizador.
func createNewThreads() (string, string, error) {
	threadID, err := createThread()
	if err != nil {
		return "", "", fmt.Errorf("fallo al crear hilo principal: %w", err)
	}
	threadIDAnalizer, err := createThreadAnalizer()
	if err != nil {
		return "", "", fmt.Errorf("fallo al crear hilo analizador: %w", err)
	}
	return threadID, threadIDAnalizer, nil
}

// createThread crea un nuevo hilo para el usuario.
func createThread() (string, error) {
	logger.Log.Info("Intentando establecer conexión gRPC con el servidor en el puerto 50052")
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return "", fmt.Errorf("fallo al conectar con el servidor gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewWhatsAppServiceClient(conn)
	logger.Log.Info("Cliente de WhatsAppService creado")

	logger.Log.Info("Llamando al método CreateThread del servicio gRPC")
	res, err := client.CreateThread(context.Background(), &pb.CreateThreadRequest{})
	if err != nil {
		return "", fmt.Errorf("fallo al crear el hilo en el servidor gRPC: %w", err)
	}

	logger.Log.Infof("Hilo creado exitosamente con ID: %s", res.ThreadId)
	return res.ThreadId, nil
}

// createThreadAnalizer crea un nuevo hilo para el analizador.
func createThreadAnalizer() (string, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return "", fmt.Errorf("fallo al conectar con el servidor gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewWhatsAppServiceClient(conn)
	res, err := client.CreateThreadAnalizer(context.Background(), &pb.CreateThreadAnalizerRequest{})
	if err != nil {
		return "", fmt.Errorf("fallo al crear el hilo analizador en el servidor gRPC: %w", err)
	}

	logger.Log.Infof("Hilo analizador creado exitosamente con ID: %s", res.ThreadIdAnalizer)
	return res.ThreadIdAnalizer, nil
}

// generateResponse genera una respuesta para el usuario.
func generateResponse(phone, threadID, messageBody string) (string, string, []string, error) {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return "", "", nil, fmt.Errorf("fallo al conectar con el servidor gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewWhatsAppServiceClient(conn)
	res, err := client.GenerateResponse(context.Background(), &pb.GenerateResponseRequest{
		Phone:       phone,
		ThreadId:    threadID,
		MessageBody: messageBody,
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("fallo al generar respuesta: %w", err)
	}

	parts := strings.Split(res.Response, "|||")
	response := parts[0]
	var followUpQuestion string
	var options []string
	if len(parts) > 1 {
		followUpQuestion = parts[1]
		if len(parts) > 2 {
			options = strings.Split(parts[2], "|")
		}
	}

	return response, followUpQuestion, options, nil
}

// generateResponseAnalizer genera una respuesta del analizador.
func generateResponseAnalizer(threadIDAnalizer string, messageBody string) ([]string, error) {
	logger.Log.Info("Generando respuesta del analizador")

	// Llamada al servicio gRPC para obtener la respuesta del analizador
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		logger.Log.Errorf("Fallo al conectar con el servidor gRPC: %v", err)
		return nil, fmt.Errorf("fallo al conectar con el servidor gRPC: %w", err)
	}
	defer conn.Close()

	client := pb.NewWhatsAppServiceClient(conn)
	res, err := client.GenerateResponseAnalizer(context.Background(), &pb.GenerateResponseAnalizerRequest{
		ThreadIdAnalizer: threadIDAnalizer,
		MessageBody:      messageBody,
	})
	if err != nil {
		logger.Log.Errorf("Fallo al generar respuesta del analizador: %v", err)
		return nil, fmt.Errorf("fallo al generar respuesta del analizador: %w", err)
	}

	logger.Log.Infof("Respuesta cruda del analizador: %s", res.Response)

	interesesRaw := strings.TrimSpace(res.Response)
	if interesesRaw == "" {
		logger.Log.Info("No se encontraron intereses en la respuesta del analizador")
		return []string{}, nil
	}

	// Procesar y validar los intereses usando la función ProcesarInteresesUsuario
	interesesValidados := utils.ProcesarInteresesUsuario(interesesRaw)

	if len(interesesValidados) > 0 {
		logger.Log.Infof("Intereses validados encontrados: %v", interesesValidados)
	} else {
		logger.Log.Info("No se encontraron intereses válidos")
	}

	logger.Log.Infof("Respuesta del analizador procesada. Intereses validados: %v", interesesValidados)
	return interesesValidados, nil
}

// retrieveThreads recupera los hilos del usuario y del analizador desde Redis.
func retrieveThreads(sessionDataRaw string) (string, string) {
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionDataRaw), &sessionData); err != nil {
		logger.Log.Errorf("Error al deserializar datos de sesión: %v", err)
		return "", ""
	}

	thread, ok := sessionData["thread"].(string)
	if !ok {
		logger.Log.Warn("Clave 'thread' no encontrada en la sesión")
		thread = ""
	}

	threadAnalizer, ok := sessionData["thread_analizer"].(string)
	if !ok {
		logger.Log.Warn("Clave 'thread_analizer' no encontrada en la sesión")
		threadAnalizer = ""
	}

	return thread, threadAnalizer
}

// updateSession actualiza la sesión del usuario en Redis.
func updateSession(redisConn *redis.Client, name, phone, message, messageType string) error {
	sessionKey := "usuario:" + phone
	sessionDataRaw, err := redisConn.Get(ctx, sessionKey).Result()
	if err != nil {
		return fmt.Errorf("fallo al recuperar datos de sesión de Redis: %w", err)
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(sessionDataRaw), &sessionData); err != nil {
		return fmt.Errorf("fallo al deserializar datos de sesión: %w", err)
	}

	messages := sessionData["messages"].([]interface{})
	messages = append(messages, map[string]string{
		"message":   message,
		"sender":    name,
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      messageType,
	})
	sessionData["messages"] = messages

	sessionDataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("fallo al serializar datos de sesión actualizados: %w", err)
	}

	if err := redisConn.Set(ctx, sessionKey, sessionDataBytes, 0).Err(); err != nil {
		return fmt.Errorf("fallo al actualizar sesión en Redis: %w", err)
	}

	logger.Log.Infof("Sesión actualizada en Redis con clave: %s", sessionKey)
	return nil
}

// IsWhatsAppStatusUpdate verifica si el mensaje es una actualización de estado de WhatsApp.
func IsWhatsAppStatusUpdate(body map[string]interface{}) bool {
	_, ok := body["entry"].([]interface{})[0].(map[string]interface{})["changes"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["statuses"]
	return ok
}

// GetTextMessageInput prepara los datos en formato JSON para enviar mensajes de texto a través de WhatsApp.
func GetTextMessageInput(recipient, text string) map[string]interface{} {
	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "text",
		"text": map[string]string{
			"preview_url": "false",
			"body":        text,
		},
	}
}

// GetInteractiveMessageInput prepara los datos en formato JSON para enviar mensajes interactivos con botones.
func GetInteractiveMessageInput(recipient, text string, buttons []string) map[string]interface{} {
	buttonObjects := make([]map[string]interface{}, len(buttons))
	for i, button := range buttons {
		buttonObjects[i] = map[string]interface{}{
			"type": "reply",
			"reply": map[string]string{
				"id":    fmt.Sprintf("button_%d", i+1),
				"title": button,
			},
		}
	}

	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "button",
			"body": map[string]string{
				"text": text,
			},
			"action": map[string]interface{}{
				"buttons": buttonObjects,
			},
		},
	}
}

// SendMessage envía un mensaje a través de la API de WhatsApp.
func SendMessage(phone string, messageData map[string]interface{}) error {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", os.Getenv("VERSION"), os.Getenv("PHONE_NUMBER_ID"))

	reqBody, err := json.Marshal(messageData)
	if err != nil {
		logger.Log.Errorf("Fallo al serializar cuerpo de la solicitud: %v", err)
		return fmt.Errorf("fallo al serializar cuerpo de la solicitud: %w", err)
	}

	logger.Log.Infof("Enviando mensaje a la API de WhatsApp. URL: %s, Cuerpo: %s", url, string(reqBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Log.Errorf("Fallo al crear solicitud: %v", err)
		return fmt.Errorf("fallo al crear solicitud: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("ACCESS_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Errorf("Fallo al enviar mensaje: %v", err)
		return fmt.Errorf("fallo al enviar mensaje: %w", err)
	}
	defer resp.Body.Close()

	logger.Log.Infof("Código de estado de la respuesta: %d", resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("Fallo al leer cuerpo de la respuesta: %v", err)
		return fmt.Errorf("fallo al leer cuerpo de la respuesta: %w", err)
	}
	logger.Log.Infof("Cuerpo de la respuesta: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("Fallo al enviar mensaje, código de estado: %d", resp.StatusCode)
		return fmt.Errorf("fallo al enviar mensaje, código de estado: %d", resp.StatusCode)
	}

	logger.Log.Info("Mensaje enviado exitosamente")
	return nil
}

// ProcessTextForWhatsApp formatea el texto para WhatsApp.
func ProcessTextForWhatsApp(text string) string {
	logger.Log.Info("Procesando texto para WhatsApp")

	// Crear una expresión regular para encontrar el texto entre ** **
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)

	// Reemplazar el texto entre ** ** por * *
	processed := re.ReplaceAllString(text, "*$1*")

	logger.Log.Info("Texto procesado exitosamente")
	return processed
}

// IsValidWhatsAppMessage valida la estructura del mensaje de WhatsApp.
func IsValidWhatsAppMessage(body map[string]interface{}) bool {
	logger.Log.Info("Validando estructura del mensaje de WhatsApp")

	// Verificar si el mensaje tiene la estructura esperada
	_, ok := body["entry"].([]interface{})[0].(map[string]interface{})["changes"].([]interface{})[0].(map[string]interface{})["value"].(map[string]interface{})["messages"]

	if ok {
		logger.Log.Info("Estructura del mensaje de WhatsApp válida")
	} else {
		logger.Log.Warn("Estructura del mensaje de WhatsApp inválida")
	}

	return ok
}
