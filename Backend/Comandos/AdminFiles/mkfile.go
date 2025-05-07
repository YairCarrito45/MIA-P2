package AdminFiles

import (
	"Gestor/Acciones"
	accionComando "Gestor/Comandos/AdminFiles/AccionesFileSystem"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

/*
Este comando permitirá crear un archivo,
el propietario será el usuario que actualmente ha iniciado sesión.
Tendrá los permisos 664.

mkfile

	-path 	(Obligatorio) 	Este parámetro será la ruta del archivo que se creará.
							Si ya existe debe mostrar un mensaje si se desea sobreescribir el archivo.

	-r 		(Opcional) 		Si se utiliza este parámetro y las carpetas especificadas por el parámetro
							path no existen, entonces deben crearse las carpetas padres.
							Si ya existen, no deberá crear las carpetas.

	-size 	(Opcional) 		Este parámetro indicará el tamaño en bytes del archivo, el contenido serán
							números del 0 al 9 cuantas veces sea necesario hasta cumplir el tamaño ingresado.

	-cont	(Opcional)		Indicará un archivo en el disco de la computadora,que tendrá el contenido
							del archivo. Se utilizará para cargar contenido en el archivo.
*/
func Mkfile(parametros []string) string {
	// 1) validar estructura para respuestas
	// 1) estructura para devolver respuesta
	logger := utils.NewLogger("mkfile")
	// Encabezado
	logger.LogInfo("[ MKFILE ]")

	// 2) validar parametros
	var pathDisco string // ruta dentro del disco
	r := false           // para validar la creacion de carpetas padres
	size := 0            // tamanio por defecto 0
	var pathReal string  // ruta real dentro del sistema

	usuario := Estructuras.UsuarioActual

	paramCorrectos := true
	pathInit := false
	pathRealInit := false

	// Para utilizar este comando es obligatorio que un usuario tenga una sesion abierta
	//validar que haya un usuario logeado
	if !usuario.Status {
		logger.LogError("ERROR [ MKFILE ]: Actualmente no hay ninguna sesion activa")
		return logger.GetErrors()
	}

	for _, parametro := range parametros[1:] {

		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		switch strings.ToLower(tknParam[0]) {
		case "path":
			// si el token parametro no tiene su identificador y valor es un error
			if len(tknParam) != 2 {
				logger.LogError("ERROR [ MKFILE ]: Valor desconocido del parametro, se acepta solo 1 valor para: %s", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}

			pathDisco = tknParam[1] // es el path donde se hara la carpeta. por ejemplo:
			// -path=/home/archivos/user/docs/usac --> hacer la carpeta usac

			if pathDisco != "" {

				pathDisco = strings.Trim(pathDisco, `"`) // Elimina comillas si están presentes
				//path = Acciones.RutaCorrecta(path)
				pathInit = true

			} else {
				logger.LogError("ERROR [ MKFILE]: error en ruta")
				paramCorrectos = false
				break
			}

		case "r":
			if len(tknParam) != 1 {
				fmt.Println("ERROR [ MKFILE ]: Valor desconocido del parametro ", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}
			r = true

		case "size":
			if len(tknParam) != 2 {
				logger.LogError("ERROR [ MKFILE ]: Valor desconocido del parametro, se acepta solo 1 valor para: %s", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}

			var err error // variable para el error posible

			size, err = strconv.Atoi(tknParam[1]) // string a int

			if err != nil {
				logger.LogError("ERROR [ MKFILE ]: size debe ser un valor numerico. Se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			} else if size < 0 {
				logger.LogError("ERROR [ MKFILE ]: size no puede tener un valor negativo. se leyo: %s", tknParam[1])
				paramCorrectos = false
				break
			}

		case "cont":
			if len(tknParam) != 2 {
				logger.LogError("ERROR [ MKFILE ]: Valor desconocido del parametro, se acepta solo 1 valor para: %s", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}

			pathReal = tknParam[1]

			if pathReal != "" {

				pathReal = strings.Trim(pathReal, `"`) // Elimina comillas si están presentes
				pathReal = Acciones.RutaCorrecta(pathReal)

				// nombre del disco
				//path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				ruta := strings.Split(pathReal, "/")
				nombreDisco := ruta[len(ruta)-1] // el ultimo valor de la ruta

				_, err := os.Stat(pathReal)
				if os.IsNotExist(err) {
					logger.LogError("ERROR [ MKFILE ]: El archivo %s no existe", nombreDisco)
					paramCorrectos = false
					break // Terminar el bucle porque encontramos un nombre único
				}
				pathRealInit = true
			} else {
				logger.LogError("ERROR [ MKFILE ]: error en ruta")
				paramCorrectos = false
				break
			}

		default:
			logger.LogError("ERROR [ MKFILE ]: parametro desconocido: %s", tknParam[0])
			paramCorrectos = false
			break
		}
	}

	// 3) validar logica
	if paramCorrectos && pathInit {
		// lógica para el comando

		// CARGA DE INFORMACIÓN NECESARIA PARA EL COMANDO
		// Cargar disco
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ MKFILE ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}

		defer disco.Close()

		var mbr Estructuras.MBR
		// Read object from bin file
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ MKFILE ]: Error al leer el MBR del disco")
			return logger.GetErrors()
		}

		// Buscar partición con id actual
		buscar := false
		part := -1
		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == usuario.Id {
				buscar = true
				part = i
				break
			}
		}

		if !buscar {
			logger.LogError("ERROR [ MKFILE ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		var superBloque SystemFileExt2.Superblock
		err = Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[part].Part_start))
		if err != nil {
			logger.LogError("ERROR [ MKFILE ]: Partición sin formato")
			return logger.GetErrors()
		}

		// Preparar las rutas del DISCO
		if !strings.HasPrefix(pathDisco, "/") {
			pathDisco = "/" + pathDisco
		}

		// Verificar si existe el archivo
		stepPath := strings.Split(pathDisco, "/")
		nombreArchivo := stepPath[len(stepPath)-1]
		rutaPadre := strings.Join(stepPath[:len(stepPath)-1], "/")
		if rutaPadre == "" {
			rutaPadre = "/"
		}

		// Buscar el inodo de la carpeta padre
		idInodoPadre := int32(0) // Empezamos desde la raíz

		// Si la ruta no es la raíz, buscar el inodo de la carpeta padre
		if rutaPadre != "/" {
			// Primero intentamos obtener el inodo correspondiente a la ruta padre
			idInodoPadre = SystemFileExt2.BuscarInodo(0, rutaPadre, superBloque, disco)

			fmt.Println("Debug - MKFILE: rutaPadre =", rutaPadre, "idInodoPadre =", idInodoPadre)

			// Si no existe y tenemos el flag -r, creamos todas las carpetas necesarias
			if (idInodoPadre == 0 || idInodoPadre == -1) && r {
				logger.LogInfo("Creando carpetas padres recursivamente...")

				// Iniciar desde la raíz
				idTemp := int32(0)
				rutaAcumulada := ""

				// Para cada componente de la ruta (excepto el nombre del archivo)
				for _, carpeta := range stepPath[1 : len(stepPath)-1] {
					if carpeta == "" {
						continue // Saltar componentes vacíos
					}

					rutaAcumulada += "/" + carpeta
					fmt.Println("Debug - MKFILE: Procesando carpeta:", carpeta, "Ruta acumulada:", rutaAcumulada)

					// Intentar buscar la carpeta primero
					idEncontrado := SystemFileExt2.BuscarInodo(idTemp, "/"+carpeta, superBloque, disco)
					fmt.Println("Debug - MKFILE: idEncontrado =", idEncontrado, "idTemp =", idTemp)

					if idEncontrado == idTemp {
						// La carpeta no existe, hay que crearla
						fmt.Println("Debug - MKFILE: Creando carpeta:", carpeta)
						idTemp = accionComando.CrearCarpeta(idTemp, carpeta, int64(mbr.Mbr_partitions[part].Part_start), disco, logger)
						if idTemp <= 0 {
							logger.LogError("ERROR [ MKFILE ]: No se pudo crear la carpeta padre '%s'", carpeta)
							return logger.GetErrors()
						}
					} else {
						// La carpeta ya existe, seguimos desde ahí
						idTemp = idEncontrado
					}

					fmt.Println("Debug - MKFILE: Después de procesar carpeta:", carpeta, "idTemp =", idTemp)
				}

				// Al final, idTemp contiene el inodo de la última carpeta creada
				idInodoPadre = idTemp
				fmt.Println("Debug - MKFILE: idInodoPadre final =", idInodoPadre)
			} else if idInodoPadre == 0 || idInodoPadre == -1 {
				logger.LogError("ERROR [ MKFILE ]: La carpeta padre no existe y no se especificó -r para crearla")
				return logger.GetErrors()
			}
		}

		// Determinar el contenido del archivo
		var contenido string
		// Prioridad: -cont > -size
		if pathRealInit {
			// Leer contenido del archivo real
			archivoReal, err := os.Open(pathReal)
			if err != nil {
				logger.LogError("ERROR [ MKFILE ]: No se pudo abrir el archivo %s", pathReal)
				return logger.GetErrors()
			}
			defer archivoReal.Close()

			contenidoBytes, err := ioutil.ReadAll(archivoReal)
			if err != nil {
				logger.LogError("ERROR [ MKFILE ]: Error al leer el archivo %s", pathReal)
				return logger.GetErrors()
			}

			contenido = string(contenidoBytes)
			logger.LogInfo("Contenido leído del archivo %s, tamaño: %d bytes", pathReal, len(contenido))
		} else if size > 0 {
			// Generar contenido numérico
			contenido = accionComando.GenerarContenidoNumerico(size)
			logger.LogInfo("Contenido numérico generado, tamaño: %d bytes", len(contenido))
		} else {
			// Archivo vacío
			contenido = ""
			logger.LogInfo("Se creará un archivo vacío")
		}

		// Crear el archivo
		idArchivo := accionComando.CrearArchivo(idInodoPadre, nombreArchivo, int64(mbr.Mbr_partitions[part].Part_start), disco, contenido, logger)

		if idArchivo == -1 {
			logger.LogError("ERROR [ MKFILE ]: No se pudo crear el archivo %s", nombreArchivo)
			return logger.GetErrors()
		}

		logger.LogInfo("Archivo %s creado exitosamente", nombreArchivo)

	} else {
		logger.LogError("ERROR [ MKFILE ]: Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) validar el return de las mensajes
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
