package AdminDiscos

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
)

/*
Este comando maneja las particiones en el disco
permite:  crear, elimina o modificar particiones.

	fdisk

		-size  (obligatorio)     -Tamaño de la partición a crear
		-unit  (opcional)        -Bytes(B) / Kilobytes(K) / Megabytes(M) unidades que utilizara
		-path  (obligatorio)     -Ruta en la que se encuentra el disco en el que se creará la partición
		-type  (opcional)        -Tipo de particion: Primaria (P) / Extendida (E) / Logica (L)
		-fit   (opcional)        -Tipo de ajuste de la partición. BF (Best), FF (First) o WF (worst)
	 	-name  (obligatorio)     -Indicará el nombre de la partición.
*/
func Fdisk(parametros []string) string {
	// Crear un logger para este comando
	logger := utils.NewLogger("fdisk")

	// Encabezado
	logger.LogInfo("[ F DISK ]")

	var size int                   // Obligatorio al momento de crear, luego no.
	var unit int = 1024            // Kilobytes por defecto, 1024 bytes
	var path string                // ruta del Disco
	var typePartition string = "P" // particion Primaria por defecto
	var fit string = "WF"          // peor ajuste por defecto
	var name string                // nombre de la parcion

	var actionComando int = 0 // Por defecto 0 -> crear; 1 -> edit; 2 -> delete

	paramCorrectos := true // validar que todos los parametros ingresen de forma correcta
	sizeInit := false      // para saber si entro el parametro size, false cuando no esta inicializado
	pathInit := false      // para verificar la existencia del path
	nameInit := false

	// Recorriendo los paramtros
	for _, parametro := range parametros[1:] { // a partir del primero, ya que el primero es la ruta
		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos valores: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ F DISK ]: Valor desconocido del parametro, más de 2 valores para: %s", tknParam[0])
			return logger.GetErrors()
		}

		// ---------- VALIDANDO PARAMATROS ---------------------
		switch strings.ToLower(tknParam[0]) {
		case "size":

			sizeInit = true // el valor OBLIGATORIO si viene dentro de las especificaciones
			var err error   // variable para el error posible

			size, err = strconv.Atoi(tknParam[1]) // string a int

			if err != nil {
				logger.LogError("ERROR [ F DISK ]: size debe ser un valor numerico. Se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			} else if size <= 0 {
				logger.LogError("ERROR [ F DISK ]: size debe ser mayor a cero. se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "unit":
			// K por defecto
			if strings.ToLower(tknParam[1]) == "b" {
				unit = 1
			} else if strings.ToLower(tknParam[1]) == "m" {
				unit = 1048576 // 1024*1024
			} else if strings.ToLower(tknParam[1]) != "k" {
				logger.LogError("ERROR [ F DISK ]: en -unit. Valores aceptados: b, k, m. ingreso: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "path":
			path = tknParam[1]

			if path != "" {
				// ruta correcta
				path = tknParam[1]
				path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				path = Acciones.RutaCorrecta(path)

				// nombre del disco
				//path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				ruta := strings.Split(path, "/")
				nombreDisco := ruta[len(ruta)-1] // el ultimo valor de la ruta

				pathInit = true

				_, err := os.Stat(path)
				if os.IsNotExist(err) {
					logger.LogError("ERROR [ F DISK ]: El disco %s no existe", nombreDisco)
					paramCorrectos = false
					break // Terminar el bucle porque encontramos un nombre único
				}
			} else {
				logger.LogError("ERROR [ F DISK ]: error en ruta")
				paramCorrectos = false
				break
			}

		case "type":
			// P por defecto
			if strings.ToLower(tknParam[1]) == "e" {
				typePartition = "E"
			} else if strings.ToLower(tknParam[1]) == "l" {
				typePartition = "L"
			} else if strings.ToLower(tknParam[1]) != "p" {
				logger.LogError("ERROR [ F DISK ]: en -type. Valores aceptados: e, l, p. ingreso: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "fit":
			// WF por defecto
			if strings.ToLower(tknParam[1]) == "bf" {
				fit = "B"
			} else if strings.ToLower(tknParam[1]) == "ff" {
				fit = "F"
				//Si el ajuste es ff ya esta definido por lo que si es distinto es un error
			} else if strings.ToLower(tknParam[1]) != "wf" {
				logger.LogError("ERROR [ F DISK ]: en -fit. Valores aceptados: BF, FF o WF. ingreso: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "name":
			// Eliminar comillas
			name = strings.ReplaceAll(tknParam[1], "\"", "")
			// Eliminar espacios en blanco al final
			name = strings.TrimSpace(name)
			if name != "" {
				nameInit = true
			} else {
				logger.LogError("ERROR [ F DISK ]: name parametro obligatorio, no se permite vacio")
				paramCorrectos = false
				break
			}

		default:
			logger.LogError("ERROR [ F DISK ]: parametro desconocido: %s", tknParam[0])
			paramCorrectos = false
			break
		}
	}

	if paramCorrectos {
		switch actionComando {
		case 0: // crear una particion
			fmt.Println("[ F DISK ] crear particion ")

			if sizeInit && pathInit && nameInit {

				// --------- LOGICA PARA F DISK ------------------
				// Abrir y cargar el disco
				filepath := path
				disco, err := Acciones.OpenFile(filepath) // se abre el Disco
				if err != nil {
					logger.LogError("ERROR [ F DISK ]: No se pudo leer el disco")
					return logger.GetErrors()
				}

				// EscribirParticion(disco *os.File, typePartition string, name string, size int, unit int, fit string)
				exito := Estructuras.EscribirParticion(disco, typePartition, name, size, unit, fit, logger)

				if !exito {
					// Los errores ya se registraron en el logger, no necesitas hacer nada más
					defer disco.Close() // cerrar el disco
				} else {
					defer disco.Close() // cerrar el disco
					//				fmt.Println("\n[ MK DISK ]: Proceso completado, el disco", nombreDisco, " Fue creado CORRECTAMENTE. en: ", file.Name())

					//fmt.Println("-- info fdisk --")
					logger.LogInfo("Size: %s", strconv.Itoa(size))
					logger.LogInfo("Unit: %s", strconv.Itoa(unit))
					logger.LogInfo("type: %s", typePartition)
					logger.LogInfo("fit:  %s", fit)
					logger.LogInfo("name: %s", name)

					logger.LogSuccess("\n[ F DISK ]: Proceso completado, la particion:  %s Fue creado CORRECTAMENTE en el disco: %s", name, disco.Name())
				}

			} else {
				logger.LogError("ERROR [ F DISK ]: parametros minimos obligatirios incompletos")
			}

		case 1: // editar particion
			fmt.Println(" editar particion")

		case 2: // borrar particion
			fmt.Println(" borrar particion")

		default:
			logger.LogError("ERROR [ F DISK ]: Accion para la Parti`cion no valido")
		}

	} else {
		logger.LogError("ERROR [ F DISK ]: parametros ingresados incorrectamente ")
	}

	// Al final de la función:
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
