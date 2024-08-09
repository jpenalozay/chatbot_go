package initializers

import (
	"fmt"
	"os"
	"time"

	"chatbot/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// InitPostgres inicializa y devuelve una instancia de la conexión a la base de datos.
// Implementa un patrón singleton para asegurar que la conexión a la base de datos se crea solo una vez.
func InitPostgres() error {
	if DB == nil {
		// Obtén las variables de entorno necesarias para la conexión a la base de datos
		host := os.Getenv("HOST_DB")
		port := os.Getenv("PORT_DB")
		user := os.Getenv("USER_DB")
		password := os.Getenv("PWD_DB")
		dbname := os.Getenv("NAME_DB")

		// Validar que todas las variables de entorno estén configuradas
		if host == "" || port == "" || user == "" || password == "" || dbname == "" {
			return fmt.Errorf("las variables de entorno de la base de datos no están configuradas")
		}

		// Construir el Data Source Name (DSN) para la conexión a PostgreSQL
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", host, user, password, dbname, port)

		var err error
		// Inicializar la conexión a la base de datos usando GORM
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // Utiliza nombres de tabla en singular
			},
			Logger: gormLogger.Default.LogMode(gormLogger.Info), // Configuración de log a nivel Info
		})

		if err != nil {
			return fmt.Errorf("error al conectar con la base de datos: %v", err)
		}

		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("error al obtener la interfaz genérica de la base de datos: %v", err)
		}

		// Configuración del pool de conexiones
		sqlDB.SetMaxIdleConns(10)           // Número máximo de conexiones inactivas
		sqlDB.SetMaxOpenConns(100)          // Número máximo de conexiones abiertas
		sqlDB.SetConnMaxLifetime(time.Hour) // Tiempo máximo de vida de una conexión

		logger.Log.Info("Conexión a la base de datos establecida exitosamente.")
	}

	return nil
}

// GetPostgresConn obtiene una conexión del pool de conexiones de la base de datos
func GetPostgresConn() (*gorm.DB, error) {
	if DB == nil {
		if err := InitPostgres(); err != nil {
			return nil, fmt.Errorf("error al obtener la conexión de la base de datos: %v", err)
		}
	}
	return DB, nil
}
