package models

// ComandoError representa un error en la ejecución de un comando
type ComandoError struct {
	Mensaje  string `json:"mensaje"`  // Mensaje descriptivo del error
	Tipo     string `json:"tipo"`     // Tipo de error (ej: "sintaxis", "parametro", "ejecucion")
	Comando  string `json:"comando"`  // Comando que generó el error
	Detalles string `json:"detalles"` // Detalles adicionales del error
}

// Error implementa la interfaz error
func (e *ComandoError) Error() string {
	return e.Mensaje
}

// NewComandoError crea un nuevo error de comando
func NewComandoError(mensaje, tipo, comando, detalles string) *ComandoError {
	return &ComandoError{
		Mensaje:  mensaje,
		Tipo:     tipo,
		Comando:  comando,
		Detalles: detalles,
	}
}
