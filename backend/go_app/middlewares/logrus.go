// go_app/middlewares/logrus.go

package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LogrusMiddleware registra todas las solicitudes entrantes
func LogrusMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Procesar la solicitud
		c.Next()

		// Calcular la latencia
		latency := time.Since(startTime)

		// Obtener el estado de la respuesta
		statusCode := c.Writer.Status()

		// Registro estructurado con Logrus
		logger.WithFields(logrus.Fields{
			"status":    statusCode,
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"ip":        c.ClientIP(),
			"userAgent": c.Request.UserAgent(),
			"time":      time.Now().Format(time.RFC3339),
			"latency":   latency,
		}).Info("Request details")
	}
}
