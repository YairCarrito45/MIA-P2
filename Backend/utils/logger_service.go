package utils

import (
	"fmt"
	"strings"
)

const (
	// Constantes para el tipo de mensaje
	MSG_INFO    = "INFO"
	MSG_SUCCESS = "SUCCESS"
	MSG_ERROR   = "ERROR"
	MSG_WARNING = "WARNING"
)

// Logger es una estructura que maneja el registro de mensajes
type Logger struct {
	output  strings.Builder // Para la salida normal
	errors  strings.Builder // Para los errores
	comando string          // El comando que se está procesando
}

// NewLogger crea un nuevo logger para un comando específico
func NewLogger(comando string) *Logger {
	return &Logger{
		comando: comando,
	}
}

// LogInfo registra un mensaje informativo
func (l *Logger) LogInfo(formato string, args ...interface{}) {
	mensaje := fmt.Sprintf(formato, args...)
	fmt.Println(mensaje) // Imprimir en consola para debugging
	l.output.WriteString(mensaje + "\n")
}

// LogSuccess registra un mensaje de éxito
func (l *Logger) LogSuccess(formato string, args ...interface{}) {
	mensaje := fmt.Sprintf(formato, args...)
	fmt.Println(mensaje) // Imprimir en consola para debugging
	l.output.WriteString(mensaje + "\n")
}

// LogError registra un mensaje de error
func (l *Logger) LogError(formato string, args ...interface{}) {
	mensaje := fmt.Sprintf(formato, args...)
	fmt.Println(mensaje) // Imprimir en consola para debugging
	l.errors.WriteString(mensaje + "\n")
}

// LogWarning registra un mensaje de advertencia
func (l *Logger) LogWarning(formato string, args ...interface{}) {
	mensaje := fmt.Sprintf(formato, args...)
	fmt.Println(mensaje)                 // Imprimir en consola para debugging
	l.errors.WriteString(mensaje + "\n") // Las advertencias van a errores también
}

// GetOutput devuelve toda la salida normal acumulada
func (l *Logger) GetOutput() string {
	return l.output.String()
}

// GetErrors devuelve todos los errores acumulados
func (l *Logger) GetErrors() string {
	return l.errors.String()
}

// HasErrors indica si hay errores registrados
func (l *Logger) HasErrors() bool {
	return l.errors.Len() > 0
}

// Reset limpia todos los logs
func (l *Logger) Reset() {
	l.output.Reset()
	l.errors.Reset()
}
