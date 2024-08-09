//chatbot/backend/initializers/loadEnvs.go

package initializers

import (
	"log" // Importa el paquete "log" para registrar mensajes y errores

	"github.com/joho/godotenv" // Importa el paquete "godotenv" para manejar archivos .env
)

// LoadEnvs carga las variables de entorno desde un archivo .env
// ubicado en la raíz del directorio del proyecto.
func LoadEnvs() {
	// Intenta cargar el archivo .env usando la función Load de godotenv
	err := godotenv.Load()

	// Verifica si hubo algún error al cargar el archivo .env
	if err != nil {
		// Si hay un error, registra el mensaje de error y detiene la ejecución del programa
		log.Fatal("Error loading .env file:", err)
	}
}
