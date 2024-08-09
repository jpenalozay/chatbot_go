package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User representa la estructura de un usuario en la base de datos
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Roles    []Role `gorm:"many2many:user_roles;"`
}

// SetPassword cifra la contraseña del usuario
func (u *User) SetPassword(password string) error {
	// Generar un hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compara la contraseña ingresada con la contraseña cifrada almacenada
func (u *User) CheckPassword(password string) error {
	// Comparar el hash de la contraseña con la contraseña almacenada
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// UsuarioChat representa la estructura de un usuario de chat en la base de datos
type UsuarioChat struct {
	ID                uint   `gorm:"primaryKey"`
	WaID              string `gorm:"unique;not null"`
	Nombre            string `gorm:"not null"`
	Email             string
	EmailVerificado   bool `gorm:"default:false"`
	Dni               string
	Telefono          string    `gorm:"unique;not null"`
	FechaCreacion     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	FechaModificacion time.Time `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// Hilo representa la estructura de un hilo de conversación en la base de datos
type Hilo struct {
	ID             uint   `gorm:"primaryKey"`
	UsuarioID      uint   `gorm:"not null"`
	HiloOpenAI     string `gorm:"not null"`
	HiloAnalizador string
	EstadoHilo     string    `gorm:"default:archivado"`
	FechaInicio    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	FechaFin       time.Time `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// Mensaje representa la estructura de un mensaje en la base de datos
type Mensaje struct {
	ID            uint      `gorm:"primaryKey"`
	HiloID        uint      `gorm:"not null"`
	TextoMensaje  string    `gorm:"not null"`
	TipoMensaje   string    `gorm:"not null"`
	FechaCreacion time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// Interes representa la estructura de un interés en la base de datos
type Interes struct {
	ID            uint      `gorm:"primaryKey"`
	HiloID        uint      `gorm:"not null"`
	Estado        string    `gorm:"default:archivado"`
	Interes       string    `gorm:"not null"`
	FechaCreacion time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

/*
// SaveOfRedisToRelationDB guarda los datos de Redis en la base de datos relacional
func SaveOfRedisToRelationDB(db *gorm.DB, sessionData map[string]interface{}) error {
	userInfo := sessionData["user_info"].(map[string]interface{})
	var usuario UsuarioChat

	// Intentar obtener el usuario de la base de datos
	if err := db.Where("telefono = ?", userInfo["phone"]).First(&usuario).Error; err != nil {
		// Si no se encuentra, crear un nuevo usuario
		usuario = UsuarioChat{
			Telefono: userInfo["phone"].(string),
			Nombre:   userInfo["name"].(string),
			WaID:     userInfo["phone"].(string),
			Email:    "email_defecto@example.com",
			Dni:      "12345678",
		}
		if err := db.Create(&usuario).Error; err != nil {
			logger.Log.Errorf("Error al crear el usuario: %v", err)
			return err
		}
		logger.Log.Infof("Usuario %s creado exitosamente.", usuario.Telefono)
	}

	// Crear un nuevo hilo
	hilo := Hilo{
		UsuarioID:      usuario.ID,
		HiloOpenAI:     sessionData["thread"].(string),
		HiloAnalizador: sessionData["thread_analizer"].(string),
		EstadoHilo:     "inactivo",
		FechaInicio:    time.Now(),
		FechaFin:       time.Now(),
	}
	if err := db.Create(&hilo).Error; err != nil {
		logger.Log.Errorf("Error al crear el hilo: %v", err)
		return err
	}
	logger.Log.Infof("Hilo %s creado exitosamente para el usuario %s.", hilo.HiloOpenAI, usuario.Telefono)

	// Crear los mensajes asociados al hilo
	for _, msg := range sessionData["messages"].([]interface{}) {
		message := msg.(map[string]interface{})
		mensaje := Mensaje{
			HiloID:        hilo.ID,
			TextoMensaje:  message["message"].(string),
			TipoMensaje:   message["type"].(string),
			FechaCreacion: time.Now(),
		}
		if err := db.Create(&mensaje).Error; err != nil {
			logger.Log.Errorf("Error al crear el mensaje: %v", err)
			return err
		}
	}
	logger.Log.Infof("Mensajes para el hilo %s guardados exitosamente.", hilo.HiloOpenAI)

	return nil
}
*/
