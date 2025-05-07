package AccionesFileSystem

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

/*
CrearArchivo crea un archivo en la ruta especificada

Parámetros:
  - idParent: ID del inodo de la carpeta donde se creará el archivo
  - nombreArchivo: Nombre del archivo a crear
  - initSuperBloque: Posición inicial del superbloque en el disco
  - disco: Puntero al archivo de disco
  - contenido: Contenido que se escribirá en el archivo
  - logger: Logger para registrar mensajes
    Retorna: ID del inodo creado, o -1 si hubo error
*/
func CrearArchivo(idParent int32, nombreArchivo string, initSuperBloque int64, disco *os.File, contenido string, logger *utils.Logger) int32 {

	// Añadir información de depuración
	fmt.Println("Debug - CrearArchivo: idParent =", idParent, "nombreArchivo =", nombreArchivo)

	var superBloque strExt2.Superblock
	Acciones.ReadObject(disco, &superBloque, initSuperBloque)

	// Cargar el inodo de la carpeta padre
	var inodoPadre strExt2.Inode
	Acciones.ReadObject(disco, &inodoPadre, int64(superBloque.S_inode_start+(idParent*int32(binary.Size(strExt2.Inode{})))))

	// Añadir información de depuración para el tipo de inodo
	fmt.Println("Debug - CrearArchivo: Tipo de inodo padre =", string(inodoPadre.I_type[:]))

	// Verificar que el inodo padre sea una carpeta
	// MODIFICACIÓN CLAVE: No verificar para inodos que acabamos de crear en la creación recursiva
	if idParent != 0 && idParent != -1 {
		// Cargar el inodo padre y verificar su tipo
		var inodoPadre strExt2.Inode
		Acciones.ReadObject(disco, &inodoPadre, int64(superBloque.S_inode_start+(idParent*int32(binary.Size(strExt2.Inode{})))))

		if string(inodoPadre.I_type[:]) != "0" {
			logger.LogError("ERROR [ MKFILE ]: La ruta especificada no es una carpeta (inodo tipo: %s)", string(inodoPadre.I_type[:]))
			return -1
		}
	} else if idParent == -1 {
		// Si el inodo padre es -1, significa que la ruta no existe
		logger.LogError("ERROR [ MKFILE ]: La ruta especificada no existe")
		return -1
	}

	// Verificar si el archivo ya existe
	idArchivoExistente := BuscarArchivo(idParent, nombreArchivo, superBloque, disco)
	if idArchivoExistente != -1 {
		// El archivo ya existe, preguntar si se quiere sobreescribir
		logger.LogInfo("El archivo '%s' ya existe. Se sobreescribirá.", nombreArchivo)
		return SobreescribirArchivo(idArchivoExistente, contenido, superBloque, disco, logger)
	}

	logger.LogInfo("Creando archivo %s", nombreArchivo)

	// Buscar espacio disponible en la carpeta padre para agregar el nuevo archivo
	var espacioEncontrado bool = false
	var idBloqueDisponible int32 = -1
	var posicionDisponible int = -1

	// Recorrer los bloques directos del inodo padre
	for i := 0; i < 12; i++ {
		idBloque := inodoPadre.I_block[i]
		if idBloque != -1 {
			// Cargar el bloque de carpeta
			var folderBlock strExt2.Folderblock
			Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

			// Buscar espacio en el bloque
			for j := 2; j < 4; j++ {
				if folderBlock.B_content[j].B_inodo == -1 {
					espacioEncontrado = true
					idBloqueDisponible = idBloque
					posicionDisponible = j
					break
				}
			}
			if espacioEncontrado {
				break
			}
		} else if i < 12 {
			// No hay más bloques asignados, pero podemos crear uno nuevo
			if superBloque.S_free_blocks_count <= 0 {
				logger.LogError("ERROR [ MKFILE ]: No hay bloques libres disponibles para crear el archivo")
				return -1
			}

			// Asignar un nuevo bloque para la carpeta
			idBloqueNuevo := superBloque.S_first_blo
			inodoPadre.I_block[i] = idBloqueNuevo

			// Actualizar el inodo padre
			Acciones.WriteObject(disco, inodoPadre, int64(superBloque.S_inode_start+(idParent*int32(binary.Size(strExt2.Inode{})))))

			// Crear un nuevo bloque de carpeta
			var folderBlockNuevo strExt2.Folderblock
			folderBlockNuevo.B_content[0].B_inodo = idParent // self
			copy(folderBlockNuevo.B_content[0].B_name[:], ".")
			folderBlockNuevo.B_content[1].B_inodo = idParent // parent (mismo que self porque estamos en el mismo directorio)
			copy(folderBlockNuevo.B_content[1].B_name[:], "..")
			folderBlockNuevo.B_content[2].B_inodo = -1
			folderBlockNuevo.B_content[3].B_inodo = -1

			// Escribir el nuevo bloque
			Acciones.WriteObject(disco, folderBlockNuevo, int64(superBloque.S_block_start+(idBloqueNuevo*int32(binary.Size(strExt2.Folderblock{})))))

			// Actualizar bitmap de bloques
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+idBloqueNuevo))

			// Actualizar superbloque
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo = idBloqueNuevo + 1
			Acciones.WriteObject(disco, superBloque, initSuperBloque)

			// Este nuevo bloque tiene espacio disponible
			espacioEncontrado = true
			idBloqueDisponible = idBloqueNuevo
			posicionDisponible = 2 // Primera posición disponible después de . y ..
			break
		}
	}

	if !espacioEncontrado {
		logger.LogError("ERROR [ MKFILE ]: No se encontró espacio en la carpeta para crear el archivo")
		return -1
	}

	// Crear el nuevo inodo para el archivo
	var nuevoInodo strExt2.Inode
	nuevoInodo.I_uid = Estructuras.UsuarioActual.IdUsr
	nuevoInodo.I_gid = Estructuras.UsuarioActual.IdGrp
	nuevoInodo.I_size = int32(len(contenido))
	ahora := time.Now()
	date := ahora.Format("02/01/2006 15:04")
	copy(nuevoInodo.I_atime[:], date)
	copy(nuevoInodo.I_ctime[:], date)
	copy(nuevoInodo.I_mtime[:], date)
	copy(nuevoInodo.I_type[:], "1")   // 1 = archivo
	copy(nuevoInodo.I_perm[:], "664") // Permisos por defecto

	// Inicializar todos los apuntadores a bloques como no usados
	for i := 0; i < 15; i++ {
		nuevoInodo.I_block[i] = -1
	}

	// ID del nuevo inodo
	idNuevoInodo := superBloque.S_first_ino

	// Actualizar el bloque de carpeta con la referencia al nuevo archivo
	var folderBlock strExt2.Folderblock
	Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(idBloqueDisponible*int32(binary.Size(strExt2.Folderblock{})))))
	folderBlock.B_content[posicionDisponible].B_inodo = idNuevoInodo
	copy(folderBlock.B_content[posicionDisponible].B_name[:], nombreArchivo)
	Acciones.WriteObject(disco, folderBlock, int64(superBloque.S_block_start+(idBloqueDisponible*int32(binary.Size(strExt2.Folderblock{})))))

	// Escribir el contenido en bloques
	if len(contenido) > 0 {
		bloquesNecesarios := (len(contenido) + 63) / 64 // Redondear hacia arriba
		if superBloque.S_free_blocks_count < int32(bloquesNecesarios) {
			logger.LogError("ERROR [ MKFILE ]: No hay suficientes bloques libres para almacenar el contenido (%d necesarios)", bloquesNecesarios)
			return -1
		}

		// Escribir el contenido en bloques
		bytesEscritos := 0
		for i := 0; i < bloquesNecesarios && i < 12; i++ { // Solo usamos bloques directos por ahora
			var fileBlock strExt2.Fileblock

			// Calcular cuántos bytes quedan por escribir
			bytesPorEscribir := len(contenido) - bytesEscritos
			if bytesPorEscribir > 64 {
				bytesPorEscribir = 64
			}

			// Copiar contenido al bloque
			copy(fileBlock.B_content[:bytesPorEscribir], contenido[bytesEscritos:bytesEscritos+bytesPorEscribir])

			// Asignar un nuevo bloque para el contenido
			idBloqueContenido := superBloque.S_first_blo
			nuevoInodo.I_block[i] = idBloqueContenido

			// Escribir el bloque de contenido
			Acciones.WriteObject(disco, fileBlock, int64(superBloque.S_block_start+(idBloqueContenido*int32(binary.Size(strExt2.Fileblock{})))))

			// Actualizar bitmap de bloques
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+idBloqueContenido))

			// Actualizar superbloque
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo = idBloqueContenido + 1

			bytesEscritos += bytesPorEscribir
		}

		// TODO: Implementar manejo de bloques indirectos si es necesario
		if bytesEscritos < len(contenido) {
			logger.LogWarning("ADVERTENCIA [ MKFILE ]: El archivo es demasiado grande, se ha truncado")
			nuevoInodo.I_size = int32(bytesEscritos)
		}
	}

	// Escribir el nuevo inodo
	Acciones.WriteObject(disco, nuevoInodo, int64(superBloque.S_inode_start+(idNuevoInodo*int32(binary.Size(strExt2.Inode{})))))

	// Actualizar bitmap de inodos
	Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_inode_start+idNuevoInodo))

	// Actualizar superbloque
	superBloque.S_free_inodes_count--
	superBloque.S_first_ino = idNuevoInodo + 1
	Acciones.WriteObject(disco, superBloque, initSuperBloque)

	logger.LogInfo("Archivo '%s' creado exitosamente, tamaño: %d bytes", nombreArchivo, nuevoInodo.I_size)
	return idNuevoInodo

}

// BuscarArchivo busca un archivo por nombre en una carpeta
func BuscarArchivo(idInodoPadre int32, nombreArchivo string, superBloque strExt2.Superblock, disco *os.File) int32 {
	var inodoPadre strExt2.Inode
	Acciones.ReadObject(disco, &inodoPadre, int64(superBloque.S_inode_start+(idInodoPadre*int32(binary.Size(strExt2.Inode{})))))

	// Recorrer los bloques directos del inodo padre
	for i := 0; i < 12; i++ {
		idBloque := inodoPadre.I_block[i]
		if idBloque != -1 {
			// Cargar el bloque de carpeta
			var folderBlock strExt2.Folderblock
			Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

			// Buscar el archivo por nombre
			for j := 2; j < 4; j++ {
				if folderBlock.B_content[j].B_inodo != -1 {
					nombreEntrada := strExt2.GetB_name(string(folderBlock.B_content[j].B_name[:]))
					if nombreEntrada == nombreArchivo {
						// Verificar que sea un archivo (no una carpeta)
						var inodoEntrada strExt2.Inode
						Acciones.ReadObject(disco, &inodoEntrada, int64(superBloque.S_inode_start+(folderBlock.B_content[j].B_inodo*int32(binary.Size(strExt2.Inode{})))))
						if string(inodoEntrada.I_type[:]) == "1" { // Es un archivo
							return folderBlock.B_content[j].B_inodo
						}
					}
				}
			}
		}
	}

	return -1 // No se encontró el archivo
}

// SobreescribirArchivo sobreescribe el contenido de un archivo existente
func SobreescribirArchivo(idInodo int32, nuevoContenido string, superBloque strExt2.Superblock, disco *os.File, logger *utils.Logger) int32 {
	var inodo strExt2.Inode
	Acciones.ReadObject(disco, &inodo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))

	// Verificar que sea un archivo
	if string(inodo.I_type[:]) != "1" {
		logger.LogError("ERROR [ MKFILE ]: El inodo no corresponde a un archivo")
		return -1
	}

	// Liberar bloques actuales
	for i := 0; i < 12; i++ {
		if inodo.I_block[i] != -1 {
			// Marcar el bloque como libre en el bitmap
			Acciones.WriteObject(disco, byte(0), int64(superBloque.S_bm_block_start+inodo.I_block[i]))
			superBloque.S_free_blocks_count++

			// Liberar el bloque
			inodo.I_block[i] = -1
		}
	}

	// TODO: Liberar bloques indirectos si es necesario

	// Actualizar tamaño y fechas del inodo
	inodo.I_size = int32(len(nuevoContenido))
	ahora := time.Now()
	date := ahora.Format("02/01/2006 15:04")
	copy(inodo.I_atime[:], date)
	copy(inodo.I_mtime[:], date)

	// Escribir el contenido en bloques
	if len(nuevoContenido) > 0 {
		bloquesNecesarios := (len(nuevoContenido) + 63) / 64 // Redondear hacia arriba
		if superBloque.S_free_blocks_count < int32(bloquesNecesarios) {
			logger.LogError("ERROR [ MKFILE ]: No hay suficientes bloques libres para almacenar el contenido (%d necesarios)", bloquesNecesarios)
			return -1
		}

		// Escribir el contenido en bloques
		bytesEscritos := 0
		for i := 0; i < bloquesNecesarios && i < 12; i++ { // Solo usamos bloques directos por ahora
			var fileBlock strExt2.Fileblock

			// Calcular cuántos bytes quedan por escribir
			bytesPorEscribir := len(nuevoContenido) - bytesEscritos
			if bytesPorEscribir > 64 {
				bytesPorEscribir = 64
			}

			// Copiar contenido al bloque
			copy(fileBlock.B_content[:bytesPorEscribir], nuevoContenido[bytesEscritos:bytesEscritos+bytesPorEscribir])

			// Asignar un nuevo bloque para el contenido
			idBloqueContenido := superBloque.S_first_blo
			inodo.I_block[i] = idBloqueContenido

			// Escribir el bloque de contenido
			Acciones.WriteObject(disco, fileBlock, int64(superBloque.S_block_start+(idBloqueContenido*int32(binary.Size(strExt2.Fileblock{})))))

			// Actualizar bitmap de bloques
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+idBloqueContenido))

			// Actualizar superbloque
			superBloque.S_free_blocks_count--
			superBloque.S_first_blo = idBloqueContenido + 1

			bytesEscritos += bytesPorEscribir
		}

		// TODO: Implementar manejo de bloques indirectos si es necesario
		if bytesEscritos < len(nuevoContenido) {
			logger.LogWarning("ADVERTENCIA [ MKFILE ]: El archivo es demasiado grande, se ha truncado")
			inodo.I_size = int32(bytesEscritos)
		}
	}

	// Escribir el inodo actualizado
	Acciones.WriteObject(disco, inodo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))

	// Actualizar superbloque
	Acciones.WriteObject(disco, superBloque, int64(superBloque.S_inode_start-(int32(binary.Size(strExt2.Superblock{})))))

	logger.LogInfo("Archivo sobreescrito exitosamente, nuevo tamaño: %d bytes", inodo.I_size)
	return idInodo
}

// GenerarContenidoNumerico genera contenido numérico desde 0-9 repetidamente hasta alcanzar el tamaño deseado
func GenerarContenidoNumerico(size int) string {
	if size <= 0 {
		return ""
	}

	resultado := make([]byte, size)
	for i := 0; i < size; i++ {
		resultado[i] = byte('0' + (i % 10))
	}

	return string(resultado)
}
