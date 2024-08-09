// logger/logger.go
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// CustomFormatter define un formateador personalizado para Logrus
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05,000")
	level := entry.Level.String()
	message := entry.Message

	logLine := fmt.Sprintf("%s - %s - %s\n", timestamp, level, message)
	return []byte(logLine), nil
}

func Init() {
	Log = logrus.New()

	// Configura la ruta del archivo de log correctamente
	logDir := "log_logrus"
	logFile := "myapp.log"

	// Crear el directorio si no existe
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	// Crear o abrir el archivo de log
	file, err := os.OpenFile(filepath.Join(logDir, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Configurar salida a archivo y terminal
	Log.SetFormatter(&CustomFormatter{})
	Log.SetLevel(logrus.InfoLevel)

	// Salida a archivo y a consola
	Log.SetOutput(io.MultiWriter(os.Stdout, file))

	Log.Info("Logger inicializado correctamente, salida tanto en la terminal como en el archivo.")
}
