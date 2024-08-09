package services

import (
	"chatbot/logger"
	"context"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func init() {
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Log.Fatalf("OPENAI_API_KEY no está configurada en el archivo .env")
	}

	client = openai.NewClient(apiKey)
}

// CreateThread simula la creación de un hilo utilizando la API de OpenAI
func CreateThread() string {
	logger.Log.Info("Creando hilo...")
	// Simulación de la creación de un hilo
	threadID := "new_thread_id"
	logger.Log.Infof("Hilo %s creado.", threadID)
	return threadID
}

/*
// RetrieveThread simula la recuperación de un hilo utilizando la API de OpenAI
func RetrieveThread(threadID string) *openai.ChatCompletion {
	logger.Log.Infof("Recuperando hilo %s...", threadID)
	// Simulación de la recuperación de un hilo
	return nil
}
*/

// CreateThreadAnalizer simula la creación de un hilo analizador utilizando la API de OpenAI
func CreateThreadAnalizer() string {
	logger.Log.Info("Creando hilo analizador...")
	// Simulación de la creación de un hilo analizador
	threadID := "new_thread_analizer_id"
	logger.Log.Infof("Hilo analizador %s creado.", threadID)
	return threadID
}

/*
// RetrieveThreadAnalizer simula la recuperación de un hilo analizador utilizando la API de OpenAI
func RetrieveThreadAnalizer(threadID string) *openai.ChatCompletion {
	logger.Log.Infof("Recuperando hilo analizador %s...", threadID)
	// Simulación de la recuperación de un hilo analizador
	return nil
}
*/

// DeleteThread simula la eliminación de un hilo utilizando la API de OpenAI
func DeleteThread(threadID string) {
	logger.Log.Infof("Eliminando hilo %s debido a inactividad...", threadID)
	// Simulación de la eliminación de un hilo
	logger.Log.Infof("Hilo %s eliminado.", threadID)
}

// RetrieveThreadContext simula la recuperación de mensajes de un hilo utilizando la API de OpenAI
func RetrieveThreadContext(threadID string) []*openai.ChatCompletionMessage {
	logger.Log.Infof("Recuperando el contexto del hilo %s...", threadID)
	// Simulación de la recuperación de mensajes de un hilo
	messages := []*openai.ChatCompletionMessage{
		{Role: "user", Content: "Ejemplo de mensaje"},
	}
	for _, message := range messages {
		logger.Log.Infof("Rol: %s, Contenido: %s", message.Role, message.Content)
	}
	return messages
}

// RunAssistant ejecuta el asistente en un hilo de conversación y recupera el mensaje generado
func RunAssistant(threadID string) string {
	RetrieveThreadContext(threadID)

	logger.Log.Info("Ejecutando asistente...")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "text-davinci-003",
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hola, ¿cómo estás?"},
			},
		},
	)
	if err != nil {
		logger.Log.Errorf("Error al ejecutar asistente: %v", err)
		return "Lo siento, ocurrió un error al procesar tu solicitud."
	}

	newMessage := resp.Choices[0].Message.Content
	logger.Log.Infof("Mensaje generado: %s", newMessage)
	return newMessage
}

// RunAssistantAnalizer ejecuta el asistente analizador en un hilo de conversación y recupera el mensaje generado
func RunAssistantAnalizer(threadID string) string {
	RetrieveThreadContext(threadID)

	logger.Log.Info("Ejecutando asistente analizador...")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "text-davinci-003",
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Analiza este mensaje"},
			},
		},
	)
	if err != nil {
		logger.Log.Errorf("Error al ejecutar asistente analizador: %v", err)
		return "Lo siento, ocurrió un error al procesar tu solicitud."
	}

	newMessage := resp.Choices[0].Message.Content
	logger.Log.Infof("Mensaje generado: %s", newMessage)
	return newMessage
}

// GenerateResponse genera una respuesta basada en un mensaje de usuario, utilizando el asistente de OpenAI
func GenerateResponse(name, phone, threadID, messageBody string) string {
	logger.Log.Infof("Pregunta del usuario %s con hilo %s: %s", phone, threadID, messageBody)
	response := RunAssistant(threadID)
	logger.Log.Infof("Última actividad para el hilo %s fue actualizada.", threadID)
	return response
}

// GenerateResponseAnalizer genera una respuesta analizador basada en un mensaje de usuario, utilizando el asistente de OpenAI
func GenerateResponseAnalizer(threadID, messageBody string) string {
	logger.Log.Infof("Pregunta del hilo %s para analizar: %s", threadID, messageBody)
	response := RunAssistantAnalizer(threadID)
	logger.Log.Infof("Última actividad para el hilo %s fue actualizada.", threadID)
	return response
}

// ListThreads simula la lista de todos los hilos
func ListThreads() []string {
	logger.Log.Info("Obteniendo todos los hilos...")
	threads := []string{"thread_1", "thread_2", "thread_3"}
	logger.Log.Infof("Threads listados correctamente, cantidad: %d", len(threads))
	return threads
}

// DeleteAllThreads elimina todos los hilos simulando la función
func DeleteAllThreads() {
	logger.Log.Info("Eliminando todos los hilos...")
	threads := ListThreads()
	for _, threadID := range threads {
		DeleteThread(threadID)
	}
	logger.Log.Info("Todos los hilos eliminados.")
}
