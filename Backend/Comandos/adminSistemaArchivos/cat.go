package AdminSistemaArchivos

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

func Cat(parametros []string) string {
	// 1) estructura para devolver las respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("cat")
	// Encabezado
	logger.LogInfo("[ CAT ]")

	// validar parametros
	var rutaArchivo string //obligatorio
	var listaRutas []string
	paramCorrectos := true

	usuario := Estructuras.UsuarioActual

	// validar que este un inicio de sesion activo
	if !usuario.Status {
		logger.LogError("ERROR [ CAT ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	}

	for i := 0; i < len(parametros[1:]); i++ {
		parametro := parametros[i+1]
		fmt.Println(" -> Parametro: ", parametro)

		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ CAT ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		if tknParam[1] == "" {
			logger.LogError("ERROR [ CAT ]: Ningun valor del parametro fileN puede esta vacio")
			paramCorrectos = false
			return logger.GetErrors()
		}

		param := "file" + strconv.Itoa(i+1)

		switch strings.ToLower(tknParam[0]) {
		case param: // nombre del usuario

			rutaArchivo = tknParam[1]
			rutaArchivo = strings.Trim(rutaArchivo, `"`) // Elimina comillas si están presentes
			listaRutas = append(listaRutas, rutaArchivo)
			fmt.Println("-> parametro ingresado: ", param)
			fmt.Println("-> ruta del archivo en disco .mia: ", rutaArchivo)

		default:
			logger.LogError("ERROR [ CAT ]: Parametro desconocido: '%s' :", string(tknParam[0]))
			paramCorrectos = false
			break

		}

	}

	if paramCorrectos {
		// aplicar logica para el comando:
		// Verificar que hay al menos una ruta de archivo
		if len(listaRutas) == 0 {
			logger.LogError("ERROR [ CAT ]: No se especificaron archivos para leer")
			return logger.GetErrors()
		}

		// Abrir el disco donde está la partición montada
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ CAT ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}
		defer disco.Close() // Asegurar que el disco se cierra al finalizar

		// Cargar el MBR para encontrar la partición montada
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ CAT ]: Error al leer el MBR del disco")
			return logger.GetErrors()
		}

		// Buscar la partición por su ID
		partitionIndex := -1
		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == usuario.Id {
				partitionIndex = i
				break
			}
		}

		if partitionIndex == -1 {
			logger.LogError("ERROR [ CAT ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		// Cargar el superbloque de la partición
		var superBloque SystemFileExt2.Superblock
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ CAT ]: No se pudo leer el superbloque de la partición")
			return logger.GetErrors()
		}

		// Para cada ruta en la lista, leer el contenido del archivo
		var contenidoCompleto strings.Builder

		for i, ruta := range listaRutas {
			// Agregar separador entre archivos si no es el primero
			if i > 0 {
				contenidoCompleto.WriteString("\n")
			}

			// Procesar la ruta: si empieza con '/', es una ruta absoluta desde la raíz de la partición
			if !strings.HasPrefix(ruta, "/") {
				ruta = "/" + ruta // Convertir a ruta absoluta
			}

			// Caso especial para users.txt que sabemos está en la raíz con inodo 1
			if ruta == "/users.txt" {
				// Comprobar permisos (para users.txt, solo root puede leerlo)
				fmt.Println("=====> ", usuario.Nombre)
				if usuario.Nombre != "root" {
					logger.LogError("ERROR [ CAT ]: No tiene permisos para leer el archivo %s", ruta)
					continue
				}

				// Leer el inodo del archivo users.txt (sabemos que es el inodo 1)
				var inodoUsers SystemFileExt2.Inode
				Acciones.ReadObject(disco, &inodoUsers, int64(superBloque.S_inode_start+int32(binary.Size(SystemFileExt2.Inode{}))))

				// Leer el contenido del archivo
				var contenidoArchivo string
				var fileBlock SystemFileExt2.Fileblock

				// Recorrer todos los bloques que conforman el archivo
				for _, idBloque := range inodoUsers.I_block {
					if idBloque != -1 { // Si el bloque está en uso
						// Leer el bloque
						Acciones.ReadObject(disco, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(SystemFileExt2.Fileblock{})))))
						// Añadir contenido eliminando bytes nulos al final
						blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
						contenidoArchivo += blockContent
					}
				}

				contenidoCompleto.WriteString(contenidoArchivo)
				logger.LogInfo("Leyendo contenido del archivo %s:", ruta)
			} else {
				// Para cualquier otro archivo, necesitamos buscar su inodo siguiendo la ruta
				// Comenzamos desde el inodo raíz (inodo 0)
				idInodo := int32(0)

				// Dividir la ruta en componentes
				componentes := strings.Split(strings.Trim(ruta, "/"), "/")

				// Buscar el inodo del archivo recorriendo la ruta
				for i, componente := range componentes {
					esUltimoComponente := i == len(componentes)-1

					// Leer el inodo actual
					var inodoActual SystemFileExt2.Inode
					Acciones.ReadObject(disco, &inodoActual, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(SystemFileExt2.Inode{})))))

					// Verificar que sea un directorio (excepto el último componente que puede ser archivo)
					if !esUltimoComponente && string(inodoActual.I_type[:]) != "0" {
						logger.LogError("ERROR [ CAT ]: Componente '%s' en la ruta no es un directorio", componente)
						return logger.GetErrors()
					}

					// Buscar el componente en los bloques de carpeta
					encontrado := false
					nuevoIdInodo := int32(-1)

					// Recorrer los bloques directos del inodo actual
					for _, idBloque := range inodoActual.I_block {
						if idBloque != -1 {
							// Leer el bloque de carpeta
							var folderBlock SystemFileExt2.Folderblock
							Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(SystemFileExt2.Folderblock{})))))

							// Buscar el componente en las entradas del directorio
							for j := 0; j < 4; j++ {
								if folderBlock.B_content[j].B_inodo != -1 {
									nombreEntry := strings.TrimRight(string(folderBlock.B_content[j].B_name[:]), "\x00")
									if nombreEntry == componente {
										nuevoIdInodo = folderBlock.B_content[j].B_inodo
										encontrado = true
										break
									}
								}
							}

							if encontrado {
								break
							}
						}
					}

					// Si no se encontró el componente
					if !encontrado {
						logger.LogError("ERROR [ CAT ]: No se encontró el componente '%s' en la ruta", componente)
						return logger.GetErrors()
					}

					// Actualizar el idInodo para el siguiente componente
					idInodo = nuevoIdInodo

					// Si es el último componente (archivo), verificar permisos y leer su contenido
					if esUltimoComponente {
						// Leer el inodo del archivo
						var inodoArchivo SystemFileExt2.Inode
						Acciones.ReadObject(disco, &inodoArchivo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(SystemFileExt2.Inode{})))))

						// Verificar que sea un archivo
						if string(inodoArchivo.I_type[:]) != "1" {
							logger.LogError("ERROR [ CAT ]: '%s' no es un archivo", componente)
							return logger.GetErrors()
						}

						// // Verificar permisos de lectura para el usuario actual
						// permisoStr := string(inodoArchivo.I_perm[:])
						// permisos, _ := strconv.Atoi(permisoStr)
						// permisoDueño := (permisos / 100) % 10
						// permisoGrupo := (permisos / 10) % 10
						// permisoOtros := permisos % 10

						// // Verificar si el usuario tiene permisos para leer el archivo
						// tienePermiso := false

						// // Si el usuario es el dueño del archivo
						// if inodoArchivo.I_uid == usuario.IdUsr {
						// 	tienePermiso = (permisoDueño & 4) > 0 // 4 es permiso de lectura
						// } else if inodoArchivo.I_gid == usuario.IdGrp {
						// 	// Si el usuario pertenece al grupo del archivo
						// 	tienePermiso = (permisoGrupo & 4) > 0
						// } else {
						// 	// Para otros usuarios
						// 	tienePermiso = (permisoOtros & 4) > 0
						// }

						// if !tienePermiso {
						// 	logger.LogError("ERROR [ CAT ]: No tiene permisos para leer el archivo %s", ruta)
						// 	continue
						// }

						// Leer el contenido del archivo
						var contenidoArchivo string
						var fileBlock SystemFileExt2.Fileblock

						// Recorrer todos los bloques que conforman el archivo
						for _, idBloque := range inodoArchivo.I_block {
							if idBloque != -1 { // Si el bloque está en uso
								// Leer el bloque
								Acciones.ReadObject(disco, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(SystemFileExt2.Fileblock{})))))
								// Añadir contenido eliminando bytes nulos al final
								blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
								contenidoArchivo += blockContent
							}
						}

						contenidoCompleto.WriteString(contenidoArchivo)
						logger.LogInfo("Leyendo contenido del archivo %s:", ruta)
					}
				}
			}
		}

		// Si llegamos aquí, hemos leído todos los archivos correctamente
		logger.LogInfo("Contenido completo de los archivos:\n%s", contenidoCompleto.String())

	} else {
		logger.LogError("ERROR [ CAT ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
