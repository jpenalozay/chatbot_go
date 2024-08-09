package middlewares

import (
	"chatbot/logger"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignatureRequired valida la firma de la solicitud.
func SignatureRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Hub-Signature-256")
		if len(signature) < 7 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature format"})
			c.Abort()
			return
		}
		signature = signature[7:]

		// Leer el cuerpo de la solicitud solo una vez
		body, err := c.GetRawData()
		if err != nil {
			logger.Log.Info("Error reading request body:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			c.Abort()
			return
		}
		c.Set("body", body) // Almacenar el cuerpo en el contexto
		//logger.Log.Info("Raw body:", string(body))

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSignature := hex.EncodeToString(expectedMAC)

		if signature != expectedSignature {
			logger.Log.Info("Invalid signature: verification failed. Expected:", expectedSignature, "Received:", signature)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(len(body)))
		c.Next()
	}
}
