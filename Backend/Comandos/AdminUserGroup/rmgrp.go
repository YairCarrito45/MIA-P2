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

/*
eliminará un grupo para los usuarios de la particion y se guradará en el archivo user.txt
y asu vez eliminará a todos los usuarios pertenecientes a ese grupo

RMGRP

	-name (Obligatorio) Indicará el nombre del grupo a eliminar.


	1. Abre el disco y carga el MBR para ubicar la partición montada
	2. Carga el superbloque de la partición
	3. Lee el inodo 1 que contiene el archivo users.txt
	4. Lee todos los bloques que conforman el archivo users.txt
	5. Realiza un primer análisis del contenido para encontrar el grupo a eliminar:
		5.1 Si el grupo no existe o ya está eliminado (ID=0), muestra un error
		5.2 No permite eliminar el grupo "root"
		5.3 Si encuentra el grupo, cambia su ID a 0
	6. Realiza un segundo análisis para encontrar usuarios que pertenecen al grupo:
		6.1 Para cada usuario del grupo, cambia su ID a 0
	7. Escribe el nuevo contenido en los bloques del archivo users.txt
	8. Actualiza el inodo con la nueva fecha de modificación y tamaño
*/
func Rmgrp(parametros []string) string {
	// 1) estructura para devolver respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("rmgrp")
	// Encabezado
	logger.LogInfo("[ RMGRP ]")

	// 2) validar parametros
	usuario := Estructuras.UsuarioActual
	var borrarGrupo string

	// validar que este un inicio de sesion activo
	if !usuario.Status {
		logger.LogError("ERROR [ RMGRP ]: Actualmente no hay ninguna sesion abierta")
		return logger.GetErrors()
	}

	// validar que lo este ejecutando el usuario ROOT
	if usuario.Nombre != "root" {
		logger.LogError("ERROR [ RMGRP ]: Este comando solo lo puede ejecutar el usuario root.\nEl usuario %s no tiene los permisos necesarios.", usuario.Nombre)
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
			logger.LogError("ERROR [ RMGRP ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		switch strings.ToLower(tknParam[0]) {
		case "name": // nombre del usuario
			borrarGrupo = tknParam[1]

			borrarGrupo = strings.Trim(borrarGrupo, `"`) // Elimina comillas si están presentes

			if borrarGrupo == "" {
				logger.LogError("ERROR [ RMGRP ]: el parametro Name no puede estar vacio")
				paramCorrectos = false
			}
			nameInit = true
		default:
			logger.LogError("ERROR [ RMGRP ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break

		}
	}

	// 3) validar logica para MKGRP
	if paramCorrectos && nameInit {
		// logica para validar el comando RMGRP
		// Abrir el disco donde está la partición montada
		disco, err := Acciones.OpenFile(usuario.PathD)
		if err != nil {
			logger.LogError("ERROR [ RMGRP ]: No se pudo abrir el disco %s", usuario.PathD)
			return logger.GetErrors()
		}
		defer disco.Close() // Asegurar que el disco se cierra al finalizar

		// Cargar el MBR para encontrar la partición montada
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ RMGRP ]: Error al leer el MBR del disco")
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
			logger.LogError("ERROR [ RMGRP ]: No se encontró la partición con ID %s", usuario.Id)
			return logger.GetErrors()
		}

		// Cargar el superbloque de la partición
		var superBloque SystemFileExt2.Superblock
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[partitionIndex].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ RMGRP ]: No se pudo leer el superbloque de la partición")
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
				contenidoActual += string(fileBlock.B_content[:])
			}
		}

		// Dividir el contenido por líneas para analizar cada entrada
		lineas := strings.Split(contenidoActual, "\n")

		// Variables para verificar si el grupo existe y obtener su ID
		grupoEncontrado := false
		var grupoID string

		// Construir el nuevo contenido línea por línea
		var nuevoContenido strings.Builder

		// Primera pasada: encontrar el grupo y cambiarlo a estado 0
		for _, linea := range lineas {
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")
			// Asegurarse de que haya suficientes campos para un registro de grupo
			if len(campos) >= 3 {
				// Eliminar espacios en blanco
				idActivo := strings.TrimSpace(campos[0])
				tipoRegistro := strings.TrimSpace(campos[1])
				nombreRegistro := strings.TrimSpace(campos[2])

				// Verificar si es un grupo activo con el nombre buscado
				if idActivo != "0" && tipoRegistro == "G" && nombreRegistro == borrarGrupo {
					// No podemos borrar el grupo root
					if borrarGrupo == "root" {
						logger.LogError("ERROR [ RMGRP ]: No se puede eliminar el grupo 'root'")
						return logger.GetErrors()
					}

					// Encontramos el grupo, lo "borramos" cambiando su ID a 0
					grupoEncontrado = true
					grupoID = idActivo // Guardar el ID para buscar usuarios de este grupo
					fmt.Println("ID del grupo: ", grupoID)
					// Agregar línea modificada (ID cambiado a 0)
					nuevoContenido.WriteString("0,G," + nombreRegistro + "\n")
					logger.LogInfo("Grupo '%s' marcado para eliminación", borrarGrupo)
				} else {
					// Mantener la línea sin cambios
					nuevoContenido.WriteString(linea + "\n")
				}
			} else {
				// Mantener la línea sin cambios (aunque no sea un formato válido)
				nuevoContenido.WriteString(linea + "\n")
			}
		}

		// Si no se encontró el grupo, mostrar error
		if !grupoEncontrado {
			logger.LogError("ERROR [ RMGRP ]: El grupo '%s' no existe o ya fue eliminado", borrarGrupo)
			return logger.GetErrors()
		}

		// Segunda pasada: buscar usuarios de este grupo y marcarlos como eliminados
		// Necesitamos volver a procesar el contenido original
		nuevoContenidoConUsuarios := strings.Builder{}

		for _, linea := range lineas {
			if linea == "" {
				continue // Ignorar líneas vacías
			}

			campos := strings.Split(linea, ",")

			// Verificar si es un registro de usuario (debería tener 5 campos)
			if len(campos) >= 5 && strings.TrimSpace(campos[1]) == "U" {
				idUsuario := strings.TrimSpace(campos[0])
				grupoUsuario := strings.TrimSpace(campos[2])
				nombreUsuario := strings.TrimSpace(campos[3])
				passwordUsuario := strings.TrimSpace(campos[4])

				// Si el usuario pertenece al grupo que estamos eliminando y está activo
				if grupoUsuario == borrarGrupo && idUsuario != "0" {
					// Marcar el usuario como eliminado
					nuevoContenidoConUsuarios.WriteString("0,U," + grupoUsuario + "," + nombreUsuario + "," + passwordUsuario + "\n")
					logger.LogInfo("Usuario '%s' del grupo '%s' marcado para eliminación", nombreUsuario, borrarGrupo)
				} else {
					// Mantener la línea sin cambios
					nuevoContenidoConUsuarios.WriteString(linea + "\n")
				}
			} else if len(campos) >= 3 && strings.TrimSpace(campos[1]) == "G" {
				idGrupo := strings.TrimSpace(campos[0])
				nombreGrupo := strings.TrimSpace(campos[2])

				// Si es el grupo que estamos eliminando
				if nombreGrupo == borrarGrupo && idGrupo != "0" {
					// Ya lo marcamos como eliminado en la primera pasada
					nuevoContenidoConUsuarios.WriteString("0,G," + nombreGrupo + "\n")
				} else {
					// Mantener la línea sin cambios
					nuevoContenidoConUsuarios.WriteString(linea + "\n")
				}
			} else {
				// Mantener la línea sin cambios
				nuevoContenidoConUsuarios.WriteString(linea + "\n")
			}
		}

		// Actualizar el contenido final
		contenidoFinal := nuevoContenidoConUsuarios.String()

		// Actualizar la fecha de modificación del inodo
		ahora := time.Now()
		copy(inodoUsers.I_mtime[:], ahora.Format("02/01/2006 15:04"))

		// Actualizar el tamaño del archivo en el inodo si es necesario
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

		logger.LogInfo("Grupo '%s' eliminado exitosamente", borrarGrupo)

	} else {
		logger.LogError("ERROR [ RMGRP ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) Retornar salidas
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
