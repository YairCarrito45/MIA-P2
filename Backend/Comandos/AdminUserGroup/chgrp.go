package AdminUserGroup

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func Chgrp(parametros []string) string {
	// 1) estructura para devolver respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("chgrp")
	// Encabezado
	logger.LogInfo("[ CHGRP ]")

	// 2) validar parametros
	usuario := Estructuras.UsuarioActual
	var grupo string
	var paramUser string

	// validar que este un inicio de sesion activo
	if !usuario.Status {
		logger.LogError("ERROR [ CHGRP ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	}

	// validar que lo este ejecutando el usuario ROOT
	if usuario.Nombre != "root" {
		logger.LogError("ERROR [ CHGRP ]: Este comando solo lo puede ejecutar el usuario root.\nEl usuario %s no tiene los permisos necesarios.", usuario.Nombre)
		return logger.GetErrors()
	}

	paramCorrectos := true
	userInit := false
	grupoInit := false

	for _, parametro := range parametros[1:] {
		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ CHGRP ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		switch strings.ToLower(tknParam[0]) {
		case "user": // nombre del usuario
			paramUser = tknParam[1]

			paramUser = strings.Trim(paramUser, `"`) // Elimina comillas si están presentes

			if paramUser == "" {
				logger.LogError("ERROR [ CHGRP ]: el parametro Name no puede esta vacio")
				paramCorrectos = false
			}
			userInit = true

		case "grp": // nombre del grupo a pertenecer
			grupo = tknParam[1]

			grupo = strings.Trim(grupo, `"`) // Elimina comillas si están presentes

			if grupo == "" {
				logger.LogError("ERROR [ CHGRP ]: el parametro %s no puede esta vacio", tknParam[0])
				paramCorrectos = false
			}
			grupoInit = true
		default:
			logger.LogError("ERROR [ CHGRP ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break

		}
	}

	if paramCorrectos && userInit && grupoInit {
		// No se puede cambiar el grupo del usuario root
		if paramUser == "root" {
			logger.LogError("ERROR [ CHGRP ]: No se puede cambiar el grupo del usuario 'root'")
			return logger.GetErrors()
		}

		// Abrir el disco donde está la partición montada
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ CHGRP ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}
		defer disco.Close() // Asegurar que el disco se cierra al finalizar

		// Cargar el MBR para encontrar la partición montada
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ CHGRP ]: Error al leer el MBR del disco")
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
			logger.LogError("ERROR [ CHGRP ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		// Cargar el superbloque de la partición
		var superBloque SystemFileExt2.Superblock
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ CHGRP ]: No se pudo leer el superbloque de la partición")
			return logger.GetErrors()
		}

		// Sabemos que users.txt está en el inodo 1
		var inodoUsers SystemFileExt2.Inode
		Acciones.ReadObject(disco, &inodoUsers, int64(superBloque.S_inode_start+int32(binary.Size(SystemFileExt2.Inode{}))))

		// Leer el contenido actual del archivo users.txt
		var contenidoActual string
		var fileBlock SystemFileExt2.Fileblock

		// Recorrer todos los bloques que conforman el archivo users.txt
		for _, idBloque := range inodoUsers.I_block {
			if idBloque != -1 { // Si el bloque está en uso
				// Leer el bloque
				Acciones.ReadObject(disco, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(SystemFileExt2.Fileblock{})))))
				// Añadir contenido eliminando bytes nulos al final
				blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
				contenidoActual += blockContent
			}
		}

		// Dividir el contenido por líneas para analizar cada entrada
		lineas := strings.Split(contenidoActual, "\n")

		// Variables para verificar si el usuario y el grupo existen
		usuarioEncontrado := false
		grupoEncontrado := false

		// Primero verificar si el grupo existe
		for _, linea := range lineas {
			linea = strings.TrimSpace(linea)
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")
			// Verificar si es un registro de grupo
			if len(campos) >= 3 && strings.TrimSpace(campos[1]) == "G" {
				idActivo := strings.TrimSpace(campos[0])
				nombreRegistro := strings.TrimSpace(campos[2])

				// Verificar si es un grupo activo con el nombre buscado
				if idActivo != "0" && nombreRegistro == grupo {
					grupoEncontrado = true
					break
				}
			}
		}

		// Si no se encontró el grupo, mostrar error
		if !grupoEncontrado {
			logger.LogError("ERROR [ CHGRP ]: El grupo '%s' no existe o está eliminado", grupo)
			return logger.GetErrors()
		}

		// Construir el nuevo contenido línea por línea
		var nuevoContenido strings.Builder

		// Procesar cada línea para encontrar y modificar el usuario
		for _, linea := range lineas {
			linea = strings.TrimSpace(linea)
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")

			// Verificar si es un registro de usuario
			if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" {
				idUsuario := strings.TrimSpace(campos[0])
				tipoRegistro := strings.TrimSpace(campos[1])
				grupoUsuario := strings.TrimSpace(campos[2])
				nombreUsuario := strings.TrimSpace(campos[3])
				passwordUsuario := strings.TrimSpace(campos[4])

				// Verificar si es el usuario que buscamos y está activo
				if nombreUsuario == paramUser && idUsuario != "0" {
					usuarioEncontrado = true

					// Si el usuario ya pertenece al grupo indicado
					if grupoUsuario == grupo {
						logger.LogError("ERROR [ CHGRP ]: El usuario '%s' ya pertenece al grupo '%s'", paramUser, grupo)
						return logger.GetErrors()
					}

					// Cambiar el grupo del usuario
					nuevoContenido.WriteString(idUsuario + "," + tipoRegistro + "," + grupo + "," + nombreUsuario + "," + passwordUsuario + "\n")
					logger.LogInfo("Grupo del usuario '%s' cambiado de '%s' a '%s'", paramUser, grupoUsuario, grupo)
				} else {
					// Mantener la línea sin cambios
					nuevoContenido.WriteString(linea + "\n")
				}
			} else {
				// Mantener la línea sin cambios (incluidos los grupos)
				nuevoContenido.WriteString(linea + "\n")
			}
		}

		// Si no se encontró el usuario, mostrar error
		if !usuarioEncontrado {
			logger.LogError("ERROR [ CHGRP ]: El usuario '%s' no existe o está eliminado", paramUser)
			return logger.GetErrors()
		}

		// Actualizar el contenido final
		contenidoFinal := nuevoContenido.String()

		// Actualizar la fecha de modificación del inodo
		ahora := time.Now()
		copy(inodoUsers.I_mtime[:], ahora.Format("02/01/2006 15:04"))

		// Actualizar el tamaño del archivo en el inodo
		bytesNuevoContenido := len([]byte(contenidoFinal))
		inodoUsers.I_size = int32(bytesNuevoContenido)

		// Escribir el nuevo contenido en los bloques
		bytesEscritos := 0
		bytesContenido := []byte(contenidoFinal)

		for i := 0; i < 12; i++ {
			idBloque := inodoUsers.I_block[i]
			if idBloque != -1 {
				// Crear un nuevo bloque para escribir
				var nuevoFileBlock SystemFileExt2.Fileblock

				// Calcular cuántos bytes quedan por escribir
				bytesPorEscribir := len(bytesContenido) - bytesEscritos
				if bytesPorEscribir <= 0 {
					break
				}

				// Si quedan menos de 64 bytes, escribir solo esos
				if bytesPorEscribir < 64 {
					copy(nuevoFileBlock.B_content[:bytesPorEscribir], bytesContenido[bytesEscritos:])
					// Rellenar el resto con bytes nulos para limpiar contenido anterior
					for j := bytesPorEscribir; j < 64; j++ {
						nuevoFileBlock.B_content[j] = 0
					}
				} else {
					// Si quedan más de 64 bytes, escribir un bloque completo
					copy(nuevoFileBlock.B_content[:], bytesContenido[bytesEscritos:bytesEscritos+64])
				}

				// Escribir el bloque en el disco
				Acciones.WriteObject(disco, nuevoFileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(SystemFileExt2.Fileblock{})))))

				// Actualizar el contador de bytes escritos
				bytesEscritos += 64
			}
		}

		// Escribir el inodo actualizado
		Acciones.WriteObject(disco, inodoUsers, int64(superBloque.S_inode_start+int32(binary.Size(SystemFileExt2.Inode{}))))

		logger.LogInfo("Grupo del usuario '%s' cambiado exitosamente a '%s'", paramUser, grupo)

	} else {
		logger.LogError("ERROR [ CHGRP ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) Retornar salidas
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
