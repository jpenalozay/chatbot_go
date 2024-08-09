package models

import (
	"time"

	"gorm.io/gorm"
)

// Prompt representa la entidad principal de un prompt
type Prompt struct {
	gorm.Model
	Nombre      string          `gorm:"uniqueIndex;not null"`
	Descripcion string          `gorm:"type:text"`
	Version     string          `gorm:"not null"`
	Contenido   string          `gorm:"type:text;not null"`
	CreadorID   uint            `gorm:"not null"`
	Creador     User            `gorm:"foreignKey:CreadorID"`
	EsActivo    bool            `gorm:"default:true"`
	Etiquetas   []PromptTag     `gorm:"many2many:prompt_tags;"`
	Versiones   []PromptVersion `gorm:"foreignKey:PromptID"`
	Tests       []PromptTest    `gorm:"foreignKey:PromptID"`
	Metricas    []PromptMetrica `gorm:"foreignKey:PromptID"`
}

// PromptVersion representa una versión específica de un prompt
type PromptVersion struct {
	gorm.Model
	PromptID          uint      `gorm:"not null"`
	VersionNumero     string    `gorm:"not null"`
	Contenido         string    `gorm:"type:text;not null"`
	CambiosRealizados string    `gorm:"type:text"`
	FechaCreacion     time.Time `gorm:"not null"`
}

// PromptTag representa etiquetas para categorizar prompts
type PromptTag struct {
	gorm.Model
	Nombre string `gorm:"uniqueIndex;not null"`
}

// PromptTest representa casos de prueba para un prompt
type PromptTest struct {
	gorm.Model
	PromptID              uint   `gorm:"not null"`
	EntradaPrueba         string `gorm:"type:text;not null"`
	SalidaEsperada        string `gorm:"type:text;not null"`
	UltimaEjecucion       time.Time
	ResultadoUltimaPrueba bool
}

// PromptMetrica representa métricas de rendimiento para un prompt
type PromptMetrica struct {
	gorm.Model
	PromptID                uint      `gorm:"not null"`
	FechaMetrica            time.Time `gorm:"not null"`
	TasaExito               float64
	TiempoPromedioRespuesta float64
	CantidadUsos            int
	CalidadRespuesta        float64 // Puntuación de 0 a 1
	Relevancia              float64 // Puntuación de 0 a 1
	Consistencia            float64 // Puntuación de 0 a 1
}

// PromptVariable representa variables dinámicas en un prompt
type PromptVariable struct {
	gorm.Model
	PromptID        uint   `gorm:"not null"`
	Nombre          string `gorm:"not null"`
	Descripcion     string
	TipoDato        string `gorm:"not null"` // e.g., "string", "int", "float"
	ValorPorDefecto string
}

// PromptFeedback representa retroalimentación de usuarios sobre un prompt
type PromptFeedback struct {
	gorm.Model
	PromptID      uint      `gorm:"not null"`
	UsuarioID     uint      `gorm:"not null"`
	Puntuacion    int       `gorm:"not null"` // e.g., 1-5
	Comentario    string    `gorm:"type:text"`
	FechaFeedback time.Time `gorm:"not null"`
}

// PromptOptimizacion representa sugerencias de optimización para un prompt
type PromptOptimizacion struct {
	gorm.Model
	PromptID             uint   `gorm:"not null"`
	Sugerencia           string `gorm:"type:text;not null"`
	Razonamiento         string `gorm:"type:text"`
	ImpactoEstimado      string
	EstadoImplementacion string `gorm:"default:'pendiente'"` // e.g., "pendiente", "en_progreso", "implementado", "descartado"
}
