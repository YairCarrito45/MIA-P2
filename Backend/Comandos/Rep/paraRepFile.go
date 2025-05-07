package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
)

func reporte_file(path string, id string, ruta_file string, logger *utils.Logger) {
	var pathDisco string
	existe := false

	// Busca en struct de particiones montadas el id ingresado
	for _, montado := range Estructuras.Montadas {
		if montado.Id == id {
			pathDisco = montado.PathM
			existe = true
			break
		}
	}

	if existe {
		// Obtener nombre del reporte y del disco
		tmp := strings.Split(path, "/")
		nombreReporte := strings.Split(tmp[len(tmp)-1], ".")[0]

		logger.LogInfo("%s", nombreReporte)

		// Disco a reportar
		tmp = strings.Split(pathDisco, "/")
		disco := strings.Split(tmp[len(tmp)-1], ".")[0]

		// Abrir el archivo del disco
		file, err := Acciones.OpenFile(pathDisco)
		if err != nil {
			logger.LogError("REP ERROR: No se pudo abrir el disco %s", pathDisco)
			return
		}
		defer file.Close()

		// Leer el MBR para encontrar la partición con el ID especificado
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(file, &mbr, 0); err != nil {
			logger.LogError("REP ERROR: No se pudo leer el MBR del disco %s", pathDisco)
			return
		}

		// Buscar la partición con el ID especificado
		var particion Estructuras.Partition
		encontrada := false

		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == id {
				particion = mbr.Mbr_partitions[i]
				encontrada = true
				break
			}
		}

		if !encontrada {
			logger.LogError("REP ERROR: No se encontró la partición con ID %s", id)
			return
		}

		// Leer el superbloque
		var superBloque strExt2.Superblock
		if err := Acciones.ReadObject(file, &superBloque, int64(particion.Part_start)); err != nil {
			logger.LogError("REP ERROR: No se pudo leer el superbloque de la partición con ID %s", id)
			return
		}

		// Asegurar que la ruta comienza con /
		if !strings.HasPrefix(ruta_file, "/") {
			ruta_file = "/" + ruta_file
		}

		// Caso especial para users.txt que sabemos está en la raíz con inodo 1
		var contenidoArchivo string
		if strings.ToLower(ruta_file) == "/users.txt" {
			// Leer el inodo del archivo users.txt (sabemos que es el inodo 1)
			var inodoUsers strExt2.Inode
			Acciones.ReadObject(file, &inodoUsers, int64(superBloque.S_inode_start+int32(binary.Size(strExt2.Inode{}))))

			// Leer el contenido del archivo
			contenidoArchivo = leerContenidoArchivo(file, inodoUsers, superBloque)
		} else {
			// Para cualquier otro archivo, necesitamos buscar su inodo siguiendo la ruta
			// Comenzamos desde el inodo raíz (inodo 0)
			idInodo := int32(0)

			// Dividir la ruta en componentes
			componentes := strings.Split(strings.Trim(ruta_file, "/"), "/")

			// Buscar el inodo del archivo recorriendo la ruta
			for i, componente := range componentes {
				esUltimoComponente := i == len(componentes)-1

				// Leer el inodo actual
				var inodoActual strExt2.Inode
				Acciones.ReadObject(file, &inodoActual, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))

				// Verificar que sea un directorio (excepto el último componente que puede ser archivo)
				if !esUltimoComponente && string(inodoActual.I_type[:]) != "0" {
					logger.LogError("REP ERROR: Componente '%s' en la ruta no es un directorio", componente)
					return
				}

				// Buscar el componente en los bloques de carpeta
				encontrado := false
				nuevoIdInodo := int32(-1)

				// Recorrer los bloques directos del inodo actual
				for _, idBloque := range inodoActual.I_block {
					if idBloque != -1 {
						// Leer el bloque de carpeta
						var folderBlock strExt2.Folderblock
						Acciones.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

						// Buscar el componente en las entradas del directorio
						for j := 0; j < 4; j++ {
							if folderBlock.B_content[j].B_inodo != -1 {
								nombreEntry := strings.TrimRight(string(folderBlock.B_content[j].B_name[:]), "\x00")

								// CORRECCIÓN: Comparar sin distinguir mayúsculas/minúsculas
								if strings.EqualFold(nombreEntry, componente) {
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
					logger.LogError("REP ERROR: No se encontró el componente '%s' en la ruta", componente)
					return
				}

				// Actualizar el idInodo para el siguiente componente
				idInodo = nuevoIdInodo

				// Si es el último componente (archivo), verificar que sea un archivo y leer su contenido
				if esUltimoComponente {
					// Leer el inodo del archivo
					var inodoArchivo strExt2.Inode
					Acciones.ReadObject(file, &inodoArchivo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))

					// Verificar que sea un archivo
					if string(inodoArchivo.I_type[:]) != "1" {
						logger.LogError("REP ERROR: '%s' no es un archivo", componente)
						return
					}

					// Leer el contenido del archivo
					contenidoArchivo = leerContenidoArchivo(file, inodoArchivo, superBloque)
				}
			}
		}

		// Crear contenido del reporte (archivo de texto plano)
		var contenidoReporte strings.Builder
		contenidoReporte.WriteString("REPORTE DE ARCHIVO\n")
		contenidoReporte.WriteString("=================\n\n")
		contenidoReporte.WriteString("Disco: " + disco + "\n")
		contenidoReporte.WriteString("Partición: " + id + "\n")
		contenidoReporte.WriteString("Archivo: " + ruta_file + "\n\n")
		contenidoReporte.WriteString("CONTENIDO:\n")
		contenidoReporte.WriteString("----------\n\n")
		contenidoReporte.WriteString(contenidoArchivo)

		// Crear el directorio si no existe
		carpeta := filepath.Dir(path)
		if err := os.MkdirAll("."+carpeta, os.ModePerm); err != nil {
			logger.LogError("REP ERROR: No se pudo crear el directorio para el reporte: %v", err)
			return
		}

		// Escribir el archivo de reporte directamente
		rutaReporte := "." + path
		err = os.WriteFile(rutaReporte, []byte(contenidoReporte.String()), 0644)
		if err != nil {
			logger.LogError("REP ERROR: No se pudo crear el archivo de reporte: %v", err)
			return
		}

		logger.LogInfo("Reporte FILE del archivo %s en el disco %s creado exitosamente en %s",
			ruta_file, disco, rutaReporte)
	} else {
		logger.LogError("REP ERROR: La partición con ID %s no está montada", id)
	}
}

// Función auxiliar para leer el contenido de un archivo a partir de su inodo
func leerContenidoArchivo(file *os.File, inodo strExt2.Inode, superBloque strExt2.Superblock) string {
	var contenido strings.Builder
	var fileBlock strExt2.Fileblock

	// Recorrer todos los bloques directos que conforman el archivo
	for i := 0; i < 12; i++ {
		idBloque := inodo.I_block[i]
		if idBloque != -1 { // Si el bloque está en uso
			// Leer el bloque
			Acciones.ReadObject(file, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Fileblock{})))))
			// Añadir contenido eliminando bytes nulos al final
			blockContent := strings.TrimRight(string(fileBlock.B_content[:]), "\x00")
			contenido.WriteString(blockContent)
		}
	}

	// TODO: Implementar manejo de bloques indirectos si es necesario

	return contenido.String()
}
