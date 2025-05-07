package services

import (
	"bytes"
	"sync"
)

// ConsoleWriter es un writer personalizado que captura todo lo que se escribe en él
type ConsoleWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

// Write implementa la interfaz io.Writer
func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

// String devuelve todo el contenido capturado como string
func (w *ConsoleWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.String()
}

// Reset limpia el buffer
func (w *ConsoleWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

// NewConsoleWriter crea un nuevo ConsoleWriter
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}
