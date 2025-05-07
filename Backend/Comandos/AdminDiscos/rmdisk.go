package AdminDiscos

import (
	"Gestor/Acciones"
	"Gestor/utils"
	"fmt"
	"os"
	"strings"
)

/*
elimina un archivo que representa a un disco duro.

	rmdisk

		-path (obligatorio) -Este parámetro será la ruta en el que se eliminará el archivo
*/
func Rmdisk(parametros []string) string {

	// Crear un logger para este comando
	logger := utils.NewLogger("rmdisk")

	logger.LogInfo("[ RM DISK ]")

	pathInit := false
	var path string

	// Recorriendo los parametros
	for _, parametro := range parametros[1:] {
		fmt.Println(" -> Parametro: ", parametro)

		// token Parametro (parametro, valor)
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ RM DISK ]: Valor desconocido del parametro, más de 2 valores para: %s", tknParam[0])
			return logger.GetErrors()
		}

		// id(parametro) - valor
		switch strings.ToLower(tknParam[0]) {
		case "path":
			pathInit = true
			path = tknParam[1]
			path = strings.Trim(path, `"`) // Elimina comillas si están presentes
			path = Acciones.RutaCorrecta(path)

		default:
			logger.LogError("ERROR [ RM DISK ]: parametro desconocido: %s", tknParam[0])
			return logger.GetErrors()
		}
	}

	if pathInit {
		// Validar que el archivo exista
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				logger.LogError("ERROR [ RM DISK ]: El disco no existe: %s", path)
				return logger.GetErrors()
			}
			logger.LogError("ERROR [ RM DISK ]: Error al verificar el disco: %v", err)
			return logger.GetErrors()
		}

		// Eliminar el archivo
		err = os.Remove(path)
		if err != nil {
			logger.LogError("ERROR [ RM DISK ]: Error al eliminar el disco: %v", err)
			return logger.GetErrors()
		}

		logger.LogSuccess("[ RM DISK ]: Disco eliminado correctamente: %s", path)

	} else {
		logger.LogError("ERROR [ RM DISK ]: falta el parámetro obligatorio: path")
		return logger.GetErrors()
	}

	// Devolvemos la salida normal si no hay errores
	return logger.GetOutput()
}
