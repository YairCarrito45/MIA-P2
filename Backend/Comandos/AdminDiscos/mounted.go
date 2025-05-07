package AdminDiscos

import (
	"Gestor/Estructuras"
	"Gestor/utils"
)

func Mounted(parametros []string) string {
	// 1) validar parrametros
	// Crear un logger para este comando
	logger := utils.NewLogger("mounted")

	// Encabezado
	logger.LogInfo("[ MOUNTED ]")

	logger.LogInfo("Total de particiones montadas: %d", len(Estructuras.Montadas))

	// 2) logica para validar comando
	for _, p := range Estructuras.Montadas {
		logger.LogInfo("ID: %s   Path: %s", p.Id, p.PathM)
	}

	// 3) retornar informacion
	// Devolvemos solo la salida normal si no hay errores
	if logger.HasErrors() {
		// Si hay errores, los concatenamos a la salida
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
