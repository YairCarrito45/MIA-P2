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

/*
	Creara un grupo para los usuarios de la particion y se guradará en el archivo user.txt

mkgrp:

 01. Abre el disco donde está montada la partición
 02. Carga el MBR para ubicar la partición por su ID
 03. Carga el superbloque de la partición
 04. Lee el inodo 1 que contiene el archivo users.txt
 05. Lee todos los bloques que conforman el archivo users.txt
 06. Verifica si el grupo ya existe
 07. Encuentra el ID más alto para asignar el siguiente
 08. Crea la nueva línea para el grupo
 09. Verifica si necesita más bloques para almacenar el contenido
    9.1 Si necesita más bloques, busca un bloque libre en el bitmap
    9.2 Actualiza el bitmap, el superbloque y el inodo
 10. Escribe el nuevo contenido en los bloques
*/
func Mkgrp(parametros []string) string {
	// 1) estructura para devolver respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("mkgrp")
	// Encabezado
	logger.LogInfo("[ MKGRP ]")

	// 2) validar parametros
	usuario := Estructuras.UsuarioActual
	var nuevoGrupo string

	// validar que este un inicio de sesion activo
	if !usuario.Status {
		logger.LogError("ERROR [ MKGRP ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	}

	// validar que lo este ejecutando el usuario ROOT
	if usuario.Nombre != "root" {
		logger.LogError("ERROR [ MKGRP ]: Este comando solo lo puede ejecutar el usuario root.\nEl usuario %s no tiene los permisos necesarios.", usuario.Nombre)
		return logger.GetErrors()
	}

	paramCorrectos := true
	nameInit := false

	for _, parametro := range parametros[1:] {
		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ MKGRP ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		switch strings.ToLower(tknParam[0]) {
		case "name": // nombre del usuario
			nuevoGrupo = tknParam[1]
			nuevoGrupo = strings.Trim(nuevoGrupo, `"`) // Elimina comillas si están presentes

			if nuevoGrupo == "" {
				logger.LogError("ERROR [ MKGRP ]: el parametro Name no puede esta vacio")
				paramCorrectos = false
			}
			nameInit = true
		default:
			logger.LogError("ERROR [ MKGRP ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break

		}
	}

	// 3) validar logica para MKGRP
	if paramCorrectos && nameInit {
		// aplicar logica del comado:
		/*
			Leer el archivo users.txt existente
			Verificar que el grupo no exista ya
			Generar un nuevo ID para el grupo
			Añadir el grupo al archivo users.txt
			Actualizar las estructuras EXT2 correspondientes

		*/

		// Abrir el disco donde está la partición montada
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ MKGRP ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}
		defer disco.Close() // Asegurar que el disco se cierra al finalizar

		// Cargar el MBR para encontrar la partición montada
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ MKGRP ]: Error al leer el MBR del disco")
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
			logger.LogError("ERROR [ MKGRP ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		// Cargar el superbloque de la partición
		var superBloque SystemFileExt2.Superblock
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ MKGRP ]: No se pudo leer el superbloque de la partición")
			return logger.GetErrors()
		}

		// Sabemos que users.txt está en el inodo 1 (por la estructura del sistema)
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

		// Imprimir las líneas para depuración
		fmt.Println("Contenido leído:")
		for i, linea := range lineas {
			fmt.Printf("Línea %d: '%s'\n", i, linea)
		}

		// Verificar si el grupo ya existe y encontrar el ID más alto
		maxGID := 1 // Empezamos desde 1 porque el grupo root ya tiene el ID 1
		for _, linea := range lineas {
			linea = strings.TrimSpace(linea)
			if linea == "" {
				continue
			}

			campos := strings.Split(linea, ",")
			if len(campos) >= 3 { // GID, Tipo, Grupo
				// Eliminar espacios en blanco
				idActivo := strings.TrimSpace(campos[0])
				tipoRegistro := strings.TrimSpace(campos[1])
				nombreRegistro := strings.TrimSpace(campos[2])

				// Verificar si es un grupo (G)
				if tipoRegistro == "G" {
					// Verificar si el nombre coincide (para detectar existencia)
					if idActivo != "0" && nombreRegistro == nuevoGrupo {
						logger.LogError("ERROR [ MKGRP ]: El grupo '%s' ya existe", nuevoGrupo)
						return logger.GetErrors()
					}

					// Independientemente de si coincide el nombre, actualizar maxGID
					if idActivo != "0" { // No está borrado lógicamente
						id, err := strconv.Atoi(idActivo)
						if err == nil && id > maxGID {
							maxGID = id
						}
					}
				}
			}
		}

		// El nuevo GID será el máximo + 1
		nuevoGID := maxGID + 1

		// Crear la nueva línea para el grupo
		nuevaLinea := fmt.Sprintf("%d,G,%s\n", nuevoGID, nuevoGrupo)

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
				logger.LogError("ERROR [ MKGRP ]: No hay bloques libres disponibles para ampliar users.txt")
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
				logger.LogError("ERROR [ MKGRP ]: No se encontró un bloque libre en el bitmap")
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

			// Actualizar el tamaño del archivo en el inodo
			inodoUsers.I_size = int32(bytesNecesarios)

			// Actualizar la fecha de modificación del inodo
			ahora := time.Now()
			copy(inodoUsers.I_mtime[:], ahora.Format("02/01/2006 15:04"))

			// Escribir el superbloque actualizado
			Acciones.WriteObject(disco, superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))

			// Escribir el inodo actualizado
			Acciones.WriteObject(disco, inodoUsers, int64(superBloque.S_inode_start+int32(binary.Size(SystemFileExt2.Inode{}))))
		}

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

		fmt.Println("ID máximo encontrado para grupos:", maxGID)
		fmt.Println("Nuevo ID asignado:", nuevoGID)

		logger.LogInfo("Grupo '%s' creado exitosamente con ID: %d", nuevoGrupo, nuevoGID)

	} else {
		logger.LogError("ERROR [ MKGRP ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) Retornar salidas
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
