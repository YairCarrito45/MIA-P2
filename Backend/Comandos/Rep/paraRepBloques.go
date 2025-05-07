package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	strExt2 "Gestor/Estructuras/SystemFileExt2"
)

func reporte_block(path string, id string, logger *utils.Logger) {
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
		// Reporte
		tmp := strings.Split(path, "/") // /dir1/dir2/reporte
		nombreReporte := strings.Split(tmp[len(tmp)-1], ".")[0]

		// Disco a reportar
		tmp = strings.Split(pathDisco, "/")
		disco := strings.Split(tmp[len(tmp)-1], ".")[0]

		file, err := Acciones.OpenFile(pathDisco)
		if err != nil {
			logger.LogError("REP Error: No se pudo abrir el disco")
			return
		}
		defer file.Close()

		var mbr Estructuras.MBR
		// Leer MBR del disco
		if err := Acciones.ReadObject(file, &mbr, 0); err != nil {
			logger.LogError("REP Error: No se pudo leer el MBR")
			return
		}

		// Buscar la partición con el ID proporcionado
		var particionEncontrada bool = false
		var particion Estructuras.Partition

		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == id {
				particion = mbr.Mbr_partitions[i]
				particionEncontrada = true
				break
			}
		}

		if !particionEncontrada {
			logger.LogError("REP Error: No se encontró la partición con ID %s", id)
			return
		}

		// Leer el superbloque de la partición
		var superBloque strExt2.Superblock
		if err := Acciones.ReadObject(file, &superBloque, int64(particion.Part_start)); err != nil {
			logger.LogError("REP Error: No se pudo leer el superbloque. La partición posiblemente no está formateada.")
			return
		}

		// Verificar que la partición esté formateada como EXT2
		if superBloque.S_filesystem_type != 2 {
			logger.LogError("REP Error: La partición no está formateada como EXT2")
			return
		}

		// Iniciar la cadena del reporte
		cad := "digraph G {\n"
		cad += "  node [shape=plaintext];\n"
		cad += "  rankdir=TB;\n"
		cad += fmt.Sprintf("  label=\"Reporte de Bloques: %s\";\n", disco)

		// Leer el bitmap de bloques para saber cuáles están en uso
		var bloquesEnUso []int32

		// Leer el bitmap de bloques completo
		for i := int32(0); i < superBloque.S_blocks_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_block_start+i))

			if bite.Val[0] == 1 {
				// Este bloque está en uso, lo agregamos a la lista
				bloquesEnUso = append(bloquesEnUso, i)
			}
		}

		logger.LogInfo("Se encontraron %d bloques en uso", len(bloquesEnUso))

		// Determinar el tipo de cada bloque (necesitamos revisar los inodos)
		// Mapa para almacenar el tipo de cada bloque
		bloquesTipo := make(map[int32]string)

		// Leer todos los inodos en uso para identificar tipos de bloques
		for i := int32(0); i < superBloque.S_inodes_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_inode_start+i))

			if bite.Val[0] == 1 {
				// Este inodo está en uso
				var inodo strExt2.Inode
				Acciones.ReadObject(file, &inodo, int64(superBloque.S_inode_start+(i*int32(binary.Size(strExt2.Inode{})))))

				// Determinar qué tipo de inodo es (carpeta o archivo)
				tipoInodo := string(inodo.I_type[:])

				// Recorrer los bloques directos
				for j := 0; j < 12; j++ {
					if inodo.I_block[j] != -1 {
						if tipoInodo == "0" { // Carpeta
							bloquesTipo[inodo.I_block[j]] = "carpeta"
						} else { // Archivo
							bloquesTipo[inodo.I_block[j]] = "archivo"
						}
					}
				}

				// Bloques indirectos
				if inodo.I_block[12] != -1 {
					bloquesTipo[inodo.I_block[12]] = "apuntador"
				}
				if inodo.I_block[13] != -1 {
					bloquesTipo[inodo.I_block[13]] = "apuntador"
				}
				if inodo.I_block[14] != -1 {
					bloquesTipo[inodo.I_block[14]] = "apuntador"
				}
			}
		}

		// Generar el reporte para cada bloque en uso
		for _, idBloque := range bloquesEnUso {
			tipoBloque, existe := bloquesTipo[idBloque]
			if !existe {
				// Si no conocemos el tipo, lo marcamos como desconocido
				tipoBloque = "desconocido"
			}

			// Crear nodo específico para cada tipo de bloque
			if tipoBloque == "carpeta" {
				var folderBlock strExt2.Folderblock
				Acciones.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

				// Crear la tabla para este bloque de carpeta
				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#FFFFCC\">Bloque Carpeta %d</td></tr>\n", idBloque)
				cad += "      <tr><td bgcolor=\"#F0E68C\">b_name</td><td bgcolor=\"#F0E68C\">b_inodo</td></tr>\n"

				// Añadir cada entrada de la carpeta
				for j := 0; j < 4; j++ {
					if folderBlock.B_content[j].B_inodo != -1 {
						nombreEntrada := strExt2.GetB_name(string(folderBlock.B_content[j].B_name[:]))
						cad += fmt.Sprintf("      <tr><td>%s</td><td>%d</td></tr>\n",
							nombreEntrada, folderBlock.B_content[j].B_inodo)
					}
				}

				cad += "    </table>\n"
				cad += "  >];\n"

			} else if tipoBloque == "archivo" {
				var fileBlock strExt2.Fileblock
				Acciones.ReadObject(file, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Fileblock{})))))

				// Obtener el contenido y limitarlo para visualización
				contenido := strExt2.GetB_content(string(fileBlock.B_content[:]))
				if len(contenido) > 30 {
					contenido = contenido[:30] + "..."
				}
				// Escapar caracteres especiales
				contenido = strings.ReplaceAll(contenido, "\"", "\\\"")
				contenido = strings.ReplaceAll(contenido, "<", "&lt;")
				contenido = strings.ReplaceAll(contenido, ">", "&gt;")

				// Crear la tabla para este bloque de archivo
				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td bgcolor=\"#E6F5FF\">Bloque Archivo %d</td></tr>\n", idBloque)
				cad += fmt.Sprintf("      <tr><td>%s</td></tr>\n", contenido)
				cad += "    </table>\n"
				cad += "  >];\n"

			} else if tipoBloque == "apuntador" {
				var pointerBlock strExt2.Pointerblock
				Acciones.ReadObject(file, &pointerBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Pointerblock{})))))

				// Crear la tabla para este bloque de apuntadores
				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"1\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td bgcolor=\"#D8BFD8\">Bloque Apuntadores %d</td></tr>\n", idBloque)

				// Formatear los apuntadores en una sola celda
				cad += "      <tr><td>"

				// Mostrar todos los apuntadores
				var apuntadoresTexto []string
				for j := 0; j < 16; j++ {
					apuntadoresTexto = append(apuntadoresTexto, fmt.Sprintf("%d", pointerBlock.B_pointers[j]))
				}

				cad += strings.Join(apuntadoresTexto, ", ")
				cad += "</td></tr>\n"

				cad += "    </table>\n"
				cad += "  >];\n"
			} else {
				// Bloque de tipo desconocido
				cad += fmt.Sprintf("  bloque%d [label=\"Bloque Desconocido %d\"];\n", idBloque, idBloque)
			}
		}

		cad += "}\n"

		// Generar el reporte
		carpeta := filepath.Dir(path)
		rutaReporte := "." + carpeta + "/" + nombreReporte + ".dot"

		// Escribir directamente el archivo DOT
		dir := filepath.Dir(rutaReporte)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.LogError("REP Error: No se pudo crear el directorio para el reporte: %v", err)
			return
		}

		// Crear el archivo DOT
		dotFile, err := os.Create(rutaReporte)
		if err != nil {
			logger.LogError("REP Error: No se pudo crear el archivo DOT: %v", err)
			return
		}

		// Escribir el contenido
		_, err = dotFile.WriteString(cad)
		if err != nil {
			logger.LogError("REP Error: No se pudo escribir en el archivo DOT: %v", err)
			dotFile.Close()
			return
		}
		dotFile.Close()

		// Generar la imagen PNG
		rutaPNG := dir + "/" + nombreReporte + ".png"
		cmd := exec.Command("dot", "-Tpng", rutaReporte, "-o", rutaPNG)

		// Capturar cualquier error de salida
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		// Ejecutar el comando
		err = cmd.Run()
		if err != nil {
			logger.LogError("REP Error: No se pudo generar la imagen PNG: %v - %s", err, stderr.String())
			return
		}

		logger.LogInfo(" Reporte BLOCK del disco %s creado exitosamente", disco)
	} else {
		logger.LogError("REP Error: Id no existe")
	}
}
