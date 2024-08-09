// middlewares/auth.go
package middlewares

import (
	"net/http"
	"os"
	"strings"
	"time"

	"chatbot/initializers"
	"chatbot/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// CheckAuth es un middleware que verifica la autenticidad del token JWT en la cabecera Authorization.
func CheckAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		c.Abort()
		return
	}

	tokens := strings.Split(authHeader, " ")
	if len(tokens) != 2 || tokens[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		c.Abort()
		return
	}

	tokenString := tokens[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	if time.Now().Unix() > int64(claims["exp"].(float64)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		c.Abort()
		return
	}

	var user models.User
	initializers.DB.Where("ID = ?", claims["id"]).Preload("Roles").First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		c.Abort()
		return
	}

	c.Set("currentUser", user)
	c.Next()
}

// AuthRequired es un middleware que verifica los roles de los usuarios autenticados.
func AuthRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range user.(models.User).Roles {
			for _, requiredRole := range roles {
				if role.Name == requiredRole {
					hasRole = true
					break
				}
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tiene los permisos necesarios"})
			c.Abort()
			return
		}

		c.Next()
	}
}
