package AdminFiles

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"fmt"
	"strings"

	accionComando "Gestor/Comandos/AdminFiles/AccionesFileSystem"
)

/*
Este comando permitirá crear un folder, el propietario será el usuario que actualmente ha iniciado sesión. Tendrá los permisos 664.

	mkdir

		-path (Obligatorio)  -Este parámetro será la ruta de la carpeta que se creará.
		-p    (Opcional)     -Si se utiliza este parámetro y las carpetas padres en el parámetro path no existen, entonces deben crearse.

Si no existen las carpetas padres, debe mostrar
error, a menos que se utilice el parámetro p.
*/
func Mkdir(parametros []string) string {
	// 1) estructura para devolver respuesta
	logger := utils.NewLogger("mkdir")
	// Encabezado
	logger.LogInfo("[ MKDIR ]")

	// 2) validacion de parametros
	var path string
	usuario := Estructuras.UsuarioActual
	r := false

	paramCorrectos := true
	pathInit := false

	// Para utilizar este comando es obligatorio que un usuario tenga una sesion abierta
	//validar que haya un usuario logeado
	if !usuario.Status {
		logger.LogError("ERROR [ MKDIR ]: Actualmente no hay una sesion iniciada")
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
				logger.LogError("ERROR [ MKDIR ]: Valor desconocido del parametro, se acepta solo 1 valor para: %s", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}

			path = tknParam[1] // es el path donde se hara la carpeta. por ejemplo:
			// -path=/home/archivos/user/docs/usac --> hacer la carpeta usac

			if path != "" {

				path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				//path = Acciones.RutaCorrecta(path)
				pathInit = true

			} else {
				logger.LogError("ERROR [ MKDIR]: error en ruta")
				paramCorrectos = false
				break
			}

		case "p":
			if len(tknParam) != 1 {
				logger.LogError("ERROR [ MKDIR ]: Valor desconocido del parametro ", tknParam[0])
				paramCorrectos = false
				return logger.GetErrors()
			}
			r = true

		default:
			logger.LogError("ERROR [ MKDIR ]: parametro desconocido: %s", tknParam[0])
			paramCorrectos = false
			break
		}
	}

	// 3) validacion de logica para mkdir
	if paramCorrectos && pathInit {
		//CARGA DE INFORMACION NECESARIA PARA EL COMANDO
		//Cargar disco
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			return logger.GetErrors()
		}

		var mbr Estructuras.MBR
		// Read object from bin file
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			return logger.GetOutput()
		}

		// Close bin file
		defer disco.Close()

		//buscar particion con id actual
		buscar := false
		part := -1
		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == usuario.Id {
				buscar = true
				part = i
				break //para que ya no siga recorriendo si ya encontro la particion independientemente si se pudo o no reducir
			}
		}

		if buscar {
			var superBloque strExt2.Superblock

			err := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[part].Part_start))
			if err != nil {
				logger.LogError("ERROR [ MKDIR ]: Particion sin formato")
			}

			//Validar cada carpeta para ver si existe y crear los padres inexistentes
			stepPath := strings.Split(path, "/")
			idInicial := int32(0)
			idActual := int32(0)
			crear := -1
			for i, itemPath := range stepPath[1:] {
				idActual = strExt2.BuscarInodo(idInicial, "/"+itemPath, superBloque, disco)
				if idInicial != idActual {
					idInicial = idActual
				} else {
					crear = i + 1 //porque estoy iniciando desde 1 e i inicia en 0
					break
				}
			}

			if crear != -1 {
				if crear == len(stepPath)-1 {
					if r {
						accionComando.CrearCarpeta(idInicial, stepPath[crear], int64(mbr.Mbr_partitions[part].Part_start), disco, logger)
					} else {
						logger.LogError("ERROR [ MKDIR ]: Sin permiso de crear carpetas padre")
					}
				} else {
					if r {
						valueR := "TRUE"
						logger.LogInfo("El parametro para crear carpetas padre es %s", valueR)
						for _, item := range stepPath[crear:] {
							idInicial = accionComando.CrearCarpeta(idInicial, item, int64(mbr.Mbr_partitions[part].Part_start), disco, logger)
							if idInicial == 0 {
								logger.LogError("ERROR [ MKDIR ]: No se pudo crear carpeta")
								return logger.GetErrors()
							}
						}
					} else {
						logger.LogError("ERROR [ MKDIR ]: Sin permiso de crear carpetas padre")
					}
				}
			} else {
				logger.LogError("ERROR [ MKDIR ]: Carpeta ya existe")
			}
		}

	} else {
		logger.LogError("ERROR [ MKDIR ]: Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) validar salidas
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
