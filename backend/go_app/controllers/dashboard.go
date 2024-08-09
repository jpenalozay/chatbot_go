// chatbot/dashboard.go

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminDashboard maneja la solicitud para el dashboard de administrador
func AdminDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Bienvenido al dashboard de administrador"})
}

// UserDashboard maneja la solicitud para el dashboard de usuario
func UserDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Bienvenido al dashboard de usuario"})
}
