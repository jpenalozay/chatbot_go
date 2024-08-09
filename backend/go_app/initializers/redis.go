package initializers

import (
	"chatbot/logger"
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client

// InitRedis inicializa el cliente de Redis con un pool de conexiones
func InitRedis() error {
	// Obtener y validar la configuración del pool de conexiones de Redis desde las variables de entorno
	maxConnAge, err := strconv.Atoi(os.Getenv("REDIS_MAX_CONN_AGE"))
	if err != nil {
		maxConnAge = 5 // Valor por defecto de 5 minutos si no está configurado
		logger.Log.Infof("REDIS_MAX_CONN_AGE no configurado, usando valor por defecto: %d minutos", maxConnAge)
	} else {
		logger.Log.Infof("REDIS_MAX_CONN_AGE configurado: %d minutos", maxConnAge)
	}

	poolSize, err := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	if err != nil {
		poolSize = 10 // Valor por defecto de 10 conexiones si no está configurado
		logger.Log.Infof("REDIS_POOL_SIZE no configurado, usando valor por defecto: %d", poolSize)
	} else {
		logger.Log.Infof("REDIS_POOL_SIZE configurado: %d", poolSize)
	}

	// Inicializar el cliente de Redis con las opciones configuradas
	rdb = redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_ADDR"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           0, // Usar la base de datos por defecto
		PoolSize:     poolSize,
		MinIdleConns: 2,
		MaxConnAge:   time.Duration(maxConnAge) * time.Minute,
	})

	// Verificar la conexión a Redis
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		logger.Log.Fatalf("No se pudo conectar a Redis: %v", err)
		return err
	}

	logger.Log.Info("Conexión a Redis establecida exitosamente con un pool de conexiones.")
	return nil
}

// GetRedisConn devuelve la instancia del cliente de Redis
func GetRedisConn() (*redis.Client, error) {
	// Si el cliente de Redis no está inicializado, inicializarlo
	if rdb == nil {
		if err := InitRedis(); err != nil {
			return nil, err
		}
		logger.Log.Info("Cliente de Redis inicializado.")
	}
	return rdb, nil
}
