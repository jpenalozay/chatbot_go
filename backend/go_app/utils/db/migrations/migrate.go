package main

import (
	"chatbot/initializers"
)

func init() {
	// Cargar variables de entorno
	initializers.LoadEnv()
}

func main() {
	// Conectar a la base de datos
	initializers.InitPostgres()

	// Ejecutar migraciones y configuraciones iniciales
	initializers.Migrate()
}
