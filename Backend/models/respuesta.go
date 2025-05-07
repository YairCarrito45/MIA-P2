package models

// Respuesta representa la estructura de respuesta que enviaremos al frontend
type Respuesta struct {
	Mensaje  string   `json:"mensaje"`            // Mensaje principal
	Tipo     string   `json:"tipo"`               // Tipo de mensaje: "error", "success", "info", "warning"
	Salida   string   `json:"salida,omitempty"`   // Salida normal del procesamiento
	Errores  []string `json:"errores,omitempty"`  // Lista de errores encontrados
	Comandos []string `json:"comandos,omitempty"` // Lista de comandos procesados
}

// EntradaComando representa la estructura JSON que recibiremos desde el frontend
type EntradaComando struct {
	Texto string `json:"text"` // El texto del comando a analizar
}

// ResultadoComando representa el resultado de procesar un comando individual
type ResultadoComando struct {
	Comando string // El comando procesado
	Salida  string // La salida normal del comando
	Errores string // Errores del comando (si hay)
	Exito   bool   // Indica si el comando se ejecutó exitosamente
}
