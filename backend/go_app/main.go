// chatbot/main.go
package main

import (
	"chatbot/controllers"
	"chatbot/initializers"
	"chatbot/logger"
	"chatbot/middlewares"
	"os"

	"github.com/gin-gonic/gin"
)

// init se ejecuta antes de main y se usa para cargar variables de entorno e inicializar conexiones
func init() {
	// Inicializar logger
	logger.Init()
	logger.Log.Info("Logger inicializado correctamente.")

	// Cargar variables de entorno
	initializers.LoadEnvs()
	logger.Log.Info("Variables de entorno cargadas.")

	// Conectar a la base de datos
	if err := initializers.InitPostgres(); err != nil {
		logger.Log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	logger.Log.Info("Pool de conexiones de PostGres inicializado.")

	// Inicializar conexión a Redis
	if err := initializers.InitRedis(); err != nil {
		logger.Log.Fatalf("No se pudo inicializar Redis: %v", err)
	}
	logger.Log.Info("Pool de conexiones de Redis inicializado.")

	// Migrar la base de datos (opcional, si es necesario)
	//if err := initializers.Migrate(); err != nil {
	//	logger.Log.Fatalf("Error al migrar la base de datos: %v", err)
	//}
	//logger.Log.Info("Migraciones de la base de datos completadas.")

	// Inicializar la caché de la base de datos
	if err := initializers.InitCacheDatabase(); err != nil {
		logger.Log.Fatalf("Error al inicializar la caché de la base de datos: %v", err)
	}
	logger.Log.Info("Migraciones de Cache de la base de datos completadas.")

}

// main es el punto de entrada principal de la aplicación
func main() {
	logger.Log.Info("Iniciando el servidor...")

	// Iniciar el job de verificación de inactividad (si es necesario)
	// utils.StartInactivityCheck(pgdb, rdb)
	// logger.Log.Info("Job de verificación de inactividad iniciado.")

	// Crear un nuevo router de Gin
	router := gin.New()
	router.Use(gin.Recovery())

	// Configurar middleware de Logrus para registrar todas las solicitudes
	router.Use(middlewares.LogrusMiddleware(logger.Log))
	logger.Log.Info("Middleware de Logrus configurado.")

	// Configuración del middleware CORS para permitir solicitudes cruzadas
	router.Use(middlewares.CORS())
	logger.Log.Info("Middleware CORS configurado.")

	// Configurar rutas de la aplicación
	setupRoutes(router)
	logger.Log.Info("Rutas configuradas.")

	// Iniciar el servidor en el puerto 8000
	logger.Log.Info("Iniciando servidor en :8000")
	if err := router.Run(":8000"); err != nil {
		logger.Log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}

// setupRoutes configura todas las rutas de la aplicación
func setupRoutes(router *gin.Engine) {
	// Rutas de autenticación que no requieren estar autenticadas
	router.POST("/api/auth/login", controllers.Login)
	logger.Log.Info("Ruta POST /api/auth/login configurada.")

	router.POST("/api/auth/signup", controllers.CreateUser)
	logger.Log.Info("Ruta POST /api/auth/signup configurada.")

	// Rutas que requieren autenticación
	authGroup := router.Group("/")
	authGroup.Use(middlewares.CheckAuth)
	{
		authGroup.GET("/user/profile", controllers.GetUserProfile)
		logger.Log.Info("Ruta GET /user/profile configurada.")
	}

	// Rutas que requieren autenticación y roles específicos para admin
	adminGroup := router.Group("/admin")
	adminGroup.Use(middlewares.CheckAuth, middlewares.AuthRequired("admin"))
	{
		adminGroup.GET("/dashboard", controllers.AdminDashboard)
		logger.Log.Info("Ruta GET /admin/dashboard configurada.")
	}

	// Rutas que requieren autenticación y roles específicos para usuarios
	userGroup := router.Group("/user")
	userGroup.Use(middlewares.CheckAuth, middlewares.AuthRequired("user"))
	{
		userGroup.GET("/dashboard", controllers.UserDashboard)
		logger.Log.Info("Ruta GET /user/dashboard configurada.")
	}

	// Rutas para procesar mensajes de WhatsApp con verificación de firma
	router.GET("/webhook", controllers.WebhookGet)
	logger.Log.Info("Ruta GET /webhook configurada.")

	router.POST("/webhook", middlewares.SignatureRequired(os.Getenv("APP_SECRET")), controllers.WebhookPost)
	logger.Log.Info("Ruta POST /webhook configurada.")
}
