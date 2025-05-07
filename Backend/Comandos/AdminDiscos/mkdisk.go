package AdminDiscos

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"fmt"
	"strconv"
	"strings"
)

/*
Este comando creará un archivo binario que simulará un disco (.mia)

mkdisk

	-size (obligatoria)  Tamaño del disco
	-fit  (opcional)     BF/FF/WF ajuste a utilizar
	-unit (opcional)     Kilobytes (K)/ Megabytes (M) unidades a utilizar
	-path (obligatoria)  ruta en donde se creará el archivo
*/
func Mkdisk(parametros []string) string {
	// Crear un logger para este comando
	logger := utils.NewLogger("mkdisk")

	// Encabezado
	logger.LogInfo("[ MK DISK ]")

	var size int
	fit := "F"      // valor por deferto FF
	unit := 1048576 // valor por defecto M (1024 *1024)
	var path string // para la ruta

	paramCorrectos := true // validar que todos los parametros ingresen de forma correcta
	sizeInit := false      // para saber si entro el parametro size, false cuando no esta inicializado
	pathInit := false      // para verificar la existencia del path

	// Recorriendo los paramtros
	for _, parametro := range parametros[1:] { // a partir del primero, ya que el primero es la ruta
		fmt.Println(" -> Parametro: ", parametro)
		//logger.LogInfo()

		// token Parametro (parametro, valor) --> dos valores: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ MK DISK ]: Valor desconocido del parametro, mas de 2 valores para: %s", tknParam[0])
			paramCorrectos = false
			break // sale de analizar el parametro y no lo ejecuta
		}

		// id(parametro) - valor
		switch strings.ToLower(tknParam[0]) {
		case "size":
			sizeInit = true                       // el valor si viene dentro de las especificaciones
			var err error                         // variable para el error posible
			size, err = strconv.Atoi(tknParam[1]) // string a int
			if err != nil {
				logger.LogError("ERROR [ MK DISK ]: size debe ser un valor numerico. Se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			} else if size <= 0 {
				logger.LogError("ERROR [ MK DISK ]: size debe ser mayor a cero. se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "fit":
			// es B/W/F porque en MBR espera estos valores
			if strings.ToLower(tknParam[1]) == "bf" { //Si el ajuste es BF (best fit)
				fit = "B"
			} else if strings.ToLower(tknParam[1]) == "wf" { //Si el ajuste es WF (worst fit)
				fit = "W"
			} else if strings.ToLower(tknParam[1]) != "ff" { //Si el ajuste es diferente a ff es distinto es un error
				logger.LogError("ERROR [ MK DISK ]: para FIT los valores aceptados son: BF, FF o WF. ingreso: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "unit":
			//si la unidad es k
			if strings.ToLower(tknParam[1]) == "k" {
				unit = 1024 // bites
			} else if strings.ToLower(tknParam[1]) != "m" {
				logger.LogError("ERROR [ MK DISK ]: para UNIT los valores aceptados son: K y M. ingreso: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "path":
			pathInit = true
			path = tknParam[1]
			// TODO: validar errores, por ejemplo la ruta existe?

		default:
			logger.LogError("ERROR [ MK DISK ]: parametro desconocido: %s", tknParam[0])
			paramCorrectos = false
			break
		}
	}

	// si se llego aqui todos los parametros estan correctos.
	// -------------- validación de parametros CORRECTOS --------------
	if paramCorrectos {
		// esta información necesaria para la CREACION real del Disco
		if sizeInit && pathInit { // validar los parametros obligatorios
			// tamanio del disco
			tamanio := size * unit
			logger.LogInfo("-> Tamanio del disco: %d Bytes.", tamanio)

			// nombre del disco
			path = strings.Trim(path, `"`) // Elimina comillas si están presentes
			ruta := strings.Split(path, "/")
			nombreDisco := ruta[len(ruta)-1] // el ultimo valor de la ruta

			logger.LogInfo("-> Nombre del disco: '%s'", nombreDisco)
			logger.LogInfo("-> Fit:  %s", fit)
			logger.LogInfo("-> Unit: %s", strconv.Itoa(unit))
			logger.LogInfo("-> Path: %s", path)

			// CREAR EL DISCO -> hacer el archivo binario que simule el disco
			err := Acciones.CrearDisco(path, nombreDisco)
			if err != nil {
				logger.LogError("ERROR [ MK DISK ]: %v", err)
				return logger.GetOutput() + logger.GetErrors()
			}

			// ABRIR EL DISCO -> para completar su contenido inicial (MBR)
			file, err := Acciones.OpenFile(path)
			if err != nil {
				logger.LogError("ERROR [ MK DISK ]: No se pudo abrir el disco: %v", err)
				if file != nil {
					defer file.Close()
				}
				return logger.GetOutput() + logger.GetErrors()
			}

			// A traves del tamanio establecido llena de 0 hasta esa posición.
			datos := make([]byte, tamanio)                 // llenar el disco de Ceros (0)
			newErr := Acciones.WriteObject(file, datos, 0) //--> desde la posicion 0
			if newErr != nil {
				logger.LogError("ERROR [ MK DISK ]: Error al escribir datos: %v", newErr)
				defer file.Close()
				return logger.GetOutput() + logger.GetErrors()
			}

			// Escribir el MBR para completar el proceso de creacion del DISCO
			file, errr := Estructuras.EscribirMBR(file, tamanio, fit)
			if errr != nil {
				logger.LogError("ERROR [ MK DISK ]: Error al escribir MBR: %v", errr)
				defer file.Close()
				return logger.GetOutput() + logger.GetErrors()
			}

			defer file.Close()
			logger.LogSuccess("\n[ MK DISK ]: Proceso completado, el disco %s Fue creado CORRECTAMENTE. en: %s", nombreDisco, file.Name())
		} else {
			// Faltan parámetros obligatorios
			logger.LogError("ERROR [ MK DISK ]: Faltan parámetros obligatorios (size y/o path)")
		}
	} else {
		// Parámetros incorrectos
		logger.LogError("ERROR [ MK DISK ]: parámetros ingresados incorrectamente")
	}

	// Devolvemos solo la salida normal si no hay errores
	if logger.HasErrors() {
		// Si hay errores, los concatenamos a la salida
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
