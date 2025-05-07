package AdminUserGroup

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Mkusr(parametros []string) string {
	// 1) estructura para devolver respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("mkusr")
	// Encabezado
	logger.LogInfo("[ MKUSR ]")

	// 2) validar parametros:
	usuario := Estructuras.UsuarioActual
	var nuevoUser string
	var contrasenia string
	var grupo string

	// validar que este un inicio de sesion activo
	if !usuario.Status {
		logger.LogError("ERROR [ MKUSR ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	}

	// validar que lo este ejecutando el usuario ROOT
	if usuario.Nombre != "root" {
		logger.LogError("ERROR [ MKUSR ]: Este comando solo lo puede ejecutar el usuario root.\nEl usuario %s no tiene los permisos necesarios.", usuario.Nombre)
		return logger.GetErrors()
	}

	paramCorrectos := true
	userInit := false
	passInit := false
	grupoInit := false

	for _, parametro := range parametros[1:] {
		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ MKUSR ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		switch strings.ToLower(tknParam[0]) {
		case "user": // nombre del usuario
			nuevoUser = tknParam[1]

			nuevoUser = strings.Trim(nuevoUser, `"`) // Elimina comillas si están presentes

			if nuevoUser == "" {
				logger.LogError("ERROR [ MKUSR ]: el parametro Name no puede esta vacio")
				paramCorrectos = false
			}
			userInit = true
		case "pass": // nombre del usuario
			contrasenia = tknParam[1]

			contrasenia = strings.Trim(contrasenia, `"`) // Elimina comillas si están presentes

			if contrasenia == "" {
				logger.LogError("ERROR [ MKUSR ]: el parametro %s no puede esta vacio", tknParam[0])
				paramCorrectos = false
			}
			passInit = true
		case "grp": // nombre del grupo a pertenecer
			grupo = tknParam[1]

			grupo = strings.Trim(grupo, `"`) // Elimina comillas si están presentes

			if grupo == "" {
				logger.LogError("ERROR [ MKUSR ]: el parametro %s no puede esta vacio", tknParam[0])
				paramCorrectos = false
			}
			grupoInit = true
		default:
			logger.LogError("ERROR [ MKUSR ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break

		}
	}

	// 3) validar logica para el comando
	if paramCorrectos && userInit && passInit && grupoInit {
		// logica para mkusr
		// Verificar longitudes máximas
		if len(nuevoUser) > 10 {
			logger.LogError("ERROR [ MKUSR ]: El nombre de usuario no puede exceder 10 caracteres")
			return logger.GetErrors()
		}

		if len(contrasenia) > 10 {
			logger.LogError("ERROR [ MKUSR ]: La contraseña no puede exceder 10 caracteres")
			return logger.GetErrors()
		}

		if len(grupo) > 10 {
			logger.LogError("ERROR [ MKUSR ]: El nombre del grupo no puede exceder 10 caracteres")
			return logger.GetErrors()
		}

		// Abrir el disco donde está la partición montada
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ MKUSR ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}
		defer disco.Close() // Asegurar que el disco se cierra al finalizar

		// Cargar el MBR para encontrar la partición montada
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ MKUSR ]: Error al leer el MBR del disco")
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
			logger.LogError("ERROR [ MKUSR ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		// Cargar el superbloque de la partición
		var superBloque SystemFileExt2.Superblock
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ MKUSR ]: No se pudo leer el superbloque de la partición")
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
				// Añadir el contenido del bloque al contenido acumulado, pero eliminar bytes nulos
				blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
				contenidoActual += blockContent
			}
		}

		// Dividir el contenido por líneas para analizar cada entrada
		lineas := strings.Split(contenidoActual, "\n")

		// Variables para verificar si el grupo existe y si el usuario ya existe
		grupoEncontrado := false
		usuarioExistente := false
		var grupoID string

		// Verificar si el grupo existe y obtener su ID
		for _, linea := range lineas {
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")
			// Verificar si es un registro de grupo
			if len(campos) >= 3 {
				idActivo := strings.TrimSpace(campos[0])
				tipoRegistro := strings.TrimSpace(campos[1])
				nombreRegistro := strings.TrimSpace(campos[2])

				// Verificar si es un grupo activo con el nombre buscado
				if idActivo != "0" && tipoRegistro == "G" && nombreRegistro == grupo {
					grupoEncontrado = true
					grupoID = idActivo
					break
				}
			}
		}

		fmt.Println("ID del grupo: ", grupoID)

		// Si no se encontró el grupo, mostrar error
		if !grupoEncontrado {
			logger.LogError("ERROR [ MKUSR ]: El grupo '%s' no existe o está eliminado", grupo)
			return logger.GetErrors()
		}

		// Verificar si el usuario ya existe (independientemente del grupo)
		for _, linea := range lineas {
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")
			// Verificar si es un registro de usuario
			if len(campos) >= 5 {
				idActivo := strings.TrimSpace(campos[0])
				tipoRegistro := strings.TrimSpace(campos[1])
				nombreUsuario := strings.TrimSpace(campos[3])

				// Verificar si es un usuario activo con el nombre buscado
				if idActivo != "0" && tipoRegistro == "U" && nombreUsuario == nuevoUser {
					usuarioExistente = true
					break
				}
			}
		}

		// Si el usuario ya existe, mostrar error
		if usuarioExistente {
			logger.LogError("ERROR [ MKUSR ]: El usuario '%s' ya existe", nuevoUser)
			return logger.GetErrors()
		}

		// Encontrar el ID más alto actual para usuarios
		maxUID := 1 // Empezamos desde 1 porque el usuario root ya tiene el ID 1
		for _, linea := range lineas {
			if linea == "" {
				continue
			}

			campos := strings.Split(linea, ",")
			if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" {
				// Es un registro de usuario, verificar su ID
				idStr := strings.TrimSpace(campos[0])
				if idStr != "0" { // No está borrado lógicamente
					id, err := strconv.Atoi(idStr)
					if err == nil && id > maxUID {
						maxUID = id
					}
				}
			}
		}

		// El nuevo UID será el máximo + 1
		nuevoUID := maxUID + 1

		// Crear la nueva línea para el usuario
		nuevaLinea := fmt.Sprintf("%d,U,%s,%s,%s\n", nuevoUID, grupo, nuevoUser, contrasenia)

		// Añadir la nueva línea al contenido actual
		nuevoContenido := contenidoActual + nuevaLinea

		// Verificar si el contenido cabe en los bloques actuales o necesitamos más
		bytesNecesarios := len([]byte(nuevoContenido))
		blocksNecesarios := (bytesNecesarios + 63) / 64 // Redondear hacia arriba

		// Contar cuántos bloques se están usando actualmente
		blocksActuales := 0
		for _, idBloque := range inodoUsers.I_block {
			if idBloque != -1 {
				blocksActuales++
			}
		}

		// Si necesitamos más bloques de los que tenemos actualmente
		if blocksNecesarios > blocksActuales {
			// Necesitamos añadir bloques adicionales

			// Asegurarse de que hay bloques libres disponibles
			if superBloque.S_free_blocks_count <= 0 {
				logger.LogError("ERROR [ MKUSR ]: No hay bloques libres disponibles para ampliar users.txt")
				return logger.GetErrors()
			}

			// Buscar un bloque libre en el bitmap de bloques
			var bitValue byte
			nuevoBloque := -1

			// Recorrer el bitmap de bloques desde la posición del primer bloque libre
			for i := superBloque.S_first_blo; i < superBloque.S_blocks_count; i++ {
				posicionBitmap := superBloque.S_bm_block_start + i
				Acciones.ReadObject(disco, &bitValue, int64(posicionBitmap))

				// Si el bit es 0, el bloque está libre
				if bitValue == 0 {
					nuevoBloque = int(i)
					break
				}
			}

			if nuevoBloque == -1 {
				logger.LogError("ERROR [ MKUSR ]: No se encontró un bloque libre en el bitmap")
				return logger.GetErrors()
			}

			// Actualizar el bitmap de bloques para marcar el nuevo bloque como usado
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+int32(nuevoBloque)))

			// Actualizar el superbloque
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo = int32(nuevoBloque) + 1

			// Actualizar el inodo para que apunte al nuevo bloque
			for i := 0; i < 12; i++ {
				if inodoUsers.I_block[i] == -1 {
					inodoUsers.I_block[i] = int32(nuevoBloque)
					break
				}
			}

			// Escribir el superbloque actualizado
			Acciones.WriteObject(disco, superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		}

		// Actualizar el tamaño del archivo en el inodo
		inodoUsers.I_size = int32(bytesNecesarios)

		// Actualizar la fecha de modificación del inodo
		ahora := time.Now()
		copy(inodoUsers.I_mtime[:], ahora.Format("02/01/2006 15:04"))

		// Escribir el nuevo contenido en los bloques
		bytesEscritos := 0
		bytesContenido := []byte(nuevoContenido)

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

		logger.LogInfo("Usuario '%s' creado exitosamente en el grupo '%s' con ID: %d", nuevoUser, grupo, nuevoUID)

	} else {
		logger.LogError("ERROR [ MKUSR ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) Retornar salidas
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
