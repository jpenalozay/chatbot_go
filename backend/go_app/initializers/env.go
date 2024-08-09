package initializers

import (
	"chatbot/logger"

	"github.com/joho/godotenv"
)

// LoadEnv carga las variables de entorno desde un archivo .env
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatalf("Error loading .env file")
	}
}
