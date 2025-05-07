package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func reporte_tree(path string, id string, logger *utils.Logger) {
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
		cad += "  node [shape=none fontname=\"Arial\"];\n" // Usar shape=none en lugar de record
		cad += "  edge [fontname=\"Arial\", fontsize=10];\n"
		cad += "  rankdir=TB;\n"  // Top to Bottom para visualizar el árbol
		cad += "  ranksep=0.6;\n" // Separación entre niveles
		cad += "  nodesep=0.4;\n" // Separación entre nodos
		cad += fmt.Sprintf("  label=\"Árbol del Sistema de Archivos: %s\";\n", disco)

		// Identificar todos los inodos y bloques en uso
		fmt.Println("DEBUG - Tree Report: Identificando inodos en uso...")
		var inodosEnUso []int32
		var bloquesEnUso []int32

		// Leer el bitmap de inodos
		for i := int32(0); i < superBloque.S_inodes_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_inode_start+i))

			if bite.Val[0] == 1 {
				inodosEnUso = append(inodosEnUso, i)
			}
		}

		// Leer el bitmap de bloques
		for i := int32(0); i < superBloque.S_blocks_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_block_start+i))

			if bite.Val[0] == 1 {
				bloquesEnUso = append(bloquesEnUso, i)
			}
		}

		logger.LogInfo("Se encontraron %d inodos y %d bloques en uso", len(inodosEnUso), len(bloquesEnUso))
		fmt.Println("DEBUG - Tree Report:", len(inodosEnUso), "inodos y", len(bloquesEnUso), "bloques en uso")

		// Mapas para almacenar información de inodos y bloques
		inodosInfo := make(map[int32]strExt2.Inode)
		bloquesCarpeta := make(map[int32]strExt2.Folderblock)
		bloquesArchivo := make(map[int32]strExt2.Fileblock)
		bloquesApuntador := make(map[int32]strExt2.Pointerblock)
		tipoBloque := make(map[int32]string) // "carpeta", "archivo", "apuntador"

		// Cargar información de todos los inodos
		for _, idInodo := range inodosEnUso {
			var inodo strExt2.Inode
			Acciones.ReadObject(file, &inodo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))
			inodosInfo[idInodo] = inodo

			// Determinar tipo de bloques a los que apunta
			tipoInodo := string(inodo.I_type[:])

			// Procesar bloques directos
			for i := 0; i < 12; i++ {
				idBloque := inodo.I_block[i]
				if idBloque == -1 {
					continue
				}

				if tipoInodo == "0" { // Carpeta
					tipoBloque[idBloque] = "carpeta"
					var folderBlock strExt2.Folderblock
					Acciones.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))
					bloquesCarpeta[idBloque] = folderBlock
				} else { // Archivo
					tipoBloque[idBloque] = "archivo"
					var fileBlock strExt2.Fileblock
					Acciones.ReadObject(file, &fileBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Fileblock{})))))
					bloquesArchivo[idBloque] = fileBlock
				}
			}

			// Procesar bloques indirectos
			for i := 12; i < 15; i++ {
				idBloque := inodo.I_block[i]
				if idBloque == -1 {
					continue
				}

				tipoBloque[idBloque] = "apuntador"
				var pointerBlock strExt2.Pointerblock
				Acciones.ReadObject(file, &pointerBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Pointerblock{})))))
				bloquesApuntador[idBloque] = pointerBlock
			}
		}

		// Generar nodos de inodos usando HTML-like labels
		for _, idInodo := range inodosEnUso {
			inodo := inodosInfo[idInodo]

			// Determinar el tipo de inodo
			tipoInodo := "Carpeta"
			fillColor := "#FFFACD" // Amarillo claro para carpetas

			if string(inodo.I_type[:]) == "1" {
				tipoInodo = "Archivo"
				fillColor = "#E6F5FF" // Azul claro para archivos
			}

			// Crear nodo para el inodo con etiqueta HTML
			cad += fmt.Sprintf("  inodo%d [label=<\n", idInodo)
			cad += "    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
			cad += fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#4682B4\"><font color=\"white\">Inodo %d</font></td></tr>\n", idInodo)
			cad += fmt.Sprintf("      <tr><td>ID</td><td>%d</td></tr>\n", idInodo)
			cad += fmt.Sprintf("      <tr><td>UID</td><td>%d</td></tr>\n", inodo.I_uid)
			cad += fmt.Sprintf("      <tr><td>GID</td><td>%d</td></tr>\n", inodo.I_gid)
			cad += fmt.Sprintf("      <tr><td>Tipo</td><td>%s</td></tr>\n", tipoInodo)
			cad += fmt.Sprintf("      <tr><td>Size</td><td>%d bytes</td></tr>\n", inodo.I_size)

			// Bloques directos
			cad += "      <tr><td>AD</td><td>"
			hasBlocks := false
			for i := 0; i < 12; i++ {
				if inodo.I_block[i] != -1 {
					if hasBlocks {
						cad += ", "
					}
					cad += fmt.Sprintf("%d", inodo.I_block[i])
					hasBlocks = true
				}
			}
			cad += "</td></tr>\n"

			// Bloques indirectos
			if inodo.I_block[12] != -1 || inodo.I_block[13] != -1 || inodo.I_block[14] != -1 {
				cad += "      <tr><td>AI</td><td>"
				hasIndirect := false
				if inodo.I_block[12] != -1 {
					cad += fmt.Sprintf("S:%d", inodo.I_block[12])
					hasIndirect = true
				}
				if inodo.I_block[13] != -1 {
					if hasIndirect {
						cad += ", "
					}
					cad += fmt.Sprintf("D:%d", inodo.I_block[13])
					hasIndirect = true
				}
				if inodo.I_block[14] != -1 {
					if hasIndirect {
						cad += ", "
					}
					cad += fmt.Sprintf("T:%d", inodo.I_block[14])
				}
				cad += "</td></tr>\n"
			}

			cad += "    </table>\n"
			cad += fmt.Sprintf("  >, style=filled, fillcolor=\"%s\"];\n", fillColor)
		}

		// Generar nodos de bloques usando HTML-like labels
		for _, idBloque := range bloquesEnUso {
			tipo, existe := tipoBloque[idBloque]
			if !existe {
				continue // Si no conocemos el tipo, lo saltamos
			}

			switch tipo {
			case "carpeta":
				folderBlock := bloquesCarpeta[idBloque]
				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td colspan=\"2\" bgcolor=\"#F0E68C\">Bloque Carpeta %d</td></tr>\n", idBloque)

				// Añadir las entradas de la carpeta
				for j := 0; j < 4; j++ {
					if folderBlock.B_content[j].B_inodo != -1 {
						nombreEntrada := strExt2.GetB_name(string(folderBlock.B_content[j].B_name[:]))
						cad += fmt.Sprintf("      <tr><td>%s</td><td>%d</td></tr>\n",
							nombreEntrada, folderBlock.B_content[j].B_inodo)
					}
				}

				cad += "    </table>\n"
				cad += "  >, style=filled, fillcolor=\"#FFFFCC\"];\n"

			case "archivo":
				fileBlock := bloquesArchivo[idBloque]
				contenido := strExt2.GetB_content(string(fileBlock.B_content[:]))
				if len(contenido) > 20 {
					contenido = contenido[:20] + "..."
				}

				// Escapar caracteres especiales
				contenido = strings.ReplaceAll(contenido, "\"", "\\\"")
				contenido = strings.ReplaceAll(contenido, "<", "&lt;")
				contenido = strings.ReplaceAll(contenido, ">", "&gt;")

				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td bgcolor=\"#87CEFA\">Bloque Archivo %d</td></tr>\n", idBloque)
				cad += fmt.Sprintf("      <tr><td>%s</td></tr>\n", contenido)
				cad += "    </table>\n"
				cad += "  >, style=filled, fillcolor=\"#E6F5FF\"];\n"

			case "apuntador":
				pointerBlock := bloquesApuntador[idBloque]
				cad += fmt.Sprintf("  bloque%d [label=<\n", idBloque)
				cad += "    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\">\n"
				cad += fmt.Sprintf("      <tr><td bgcolor=\"#C8A2C8\">Bloque Apuntadores %d</td></tr>\n", idBloque)
				cad += "      <tr><td>Pointers: "

				hasPointers := false
				for j := 0; j < 16; j++ {
					if pointerBlock.B_pointers[j] != -1 {
						if hasPointers {
							cad += ", "
						}
						cad += fmt.Sprintf("%d", pointerBlock.B_pointers[j])
						hasPointers = true
					}
				}

				cad += "</td></tr></table>\n"
				cad += "  >, style=filled, fillcolor=\"#D8BFD8\"];\n"
			}
		}

		// Generar relaciones entre inodos y bloques
		fmt.Println("DEBUG - Tree Report: Generando relaciones entre nodos...")

		// Relaciones de inodos a bloques
		for _, idInodo := range inodosEnUso {
			inodo := inodosInfo[idInodo]

			// Bloques directos
			for i := 0; i < 12; i++ {
				idBloque := inodo.I_block[i]
				if idBloque != -1 {
					cad += fmt.Sprintf("  inodo%d -> bloque%d [label=\"AD[%d]\"];\n", idInodo, idBloque, i)
				}
			}

			// Bloques indirectos
			for i := 12; i < 15; i++ {
				idBloque := inodo.I_block[i]
				if idBloque != -1 {
					tipoIndirecto := ""
					switch i {
					case 12:
						tipoIndirecto = "S"
					case 13:
						tipoIndirecto = "D"
					case 14:
						tipoIndirecto = "T"
					}

					cad += fmt.Sprintf("  inodo%d -> bloque%d [label=\"AI_%s\"];\n", idInodo, idBloque, tipoIndirecto)
				}
			}
		}

		// Relaciones de bloques de carpeta a inodos
		for idBloque, folderBlock := range bloquesCarpeta {
			for j := 0; j < 4; j++ {
				if folderBlock.B_content[j].B_inodo != -1 {
					nombreEntrada := strExt2.GetB_name(string(folderBlock.B_content[j].B_name[:]))

					// Solo crear relación si no es . o ..
					if nombreEntrada != "." && nombreEntrada != ".." {
						cad += fmt.Sprintf("  bloque%d -> inodo%d [label=\"%s\"];\n",
							idBloque, folderBlock.B_content[j].B_inodo, nombreEntrada)
					}
				}
			}
		}

		// Relaciones de bloques de apuntadores a otros bloques
		for idBloque, pointerBlock := range bloquesApuntador {
			for j := 0; j < 16; j++ {
				if pointerBlock.B_pointers[j] != -1 {
					cad += fmt.Sprintf("  bloque%d -> bloque%d [label=\"[%d]\"];\n",
						idBloque, pointerBlock.B_pointers[j], j)
				}
			}
		}

		// Organizar los nodos en una estructura de árbol más clara
		// Definir los rangos para cada nivel en el árbol
		cad += "  { rank=min; inodo0; }\n" // La raíz siempre en la parte superior

		// Agrupar nodos relacionados para mejor visualización
		// Aquí usamos cluster para agrupar bloques e inodos relacionados
		for nivel := 1; nivel <= 6; nivel++ {
			cad += fmt.Sprintf("  subgraph nivel_%d {\n", nivel)
			cad += "    rank=same;\n"
			// Aquí se pueden añadir nodos específicos a cada nivel si se desea
			cad += "  }\n"
		}

		cad += "}\n"

		// Generar el reporte
		fmt.Println("DEBUG - Tree Report: Generando archivos del reporte...")
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

		// Utilizar un enfoque más robusto para generar la imagen
		// Intentar con diferentes opciones de Graphviz
		dotOptions := []string{
			"-Tpng",
			"-Gdpi=100",     // Resolución DPI
			"-Gsize=11,11",  // Tamaño en pulgadas
			"-Gratio=auto",  // Relación de aspecto automática
			"-Nfontsize=10", // Tamaño de fuente para nodos
			"-Efontsize=9",  // Tamaño de fuente para aristas
			rutaReporte,
			"-o", rutaPNG,
		}

		cmd := exec.Command("dot", dotOptions...)

		// Capturar cualquier error de salida
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		// Ejecutar el comando
		fmt.Println("DEBUG - Tree Report: Ejecutando comando dot para generar PNG...")
		err = cmd.Run()
		if err != nil {
			// Si falla, intentar una alternativa más sencilla
			fmt.Println("DEBUG - Tree Report: Primer intento fallido, probando alternativa...")
			cmdAlt := exec.Command("dot", "-Tpng", rutaReporte, "-o", rutaPNG)
			errAlt := cmdAlt.Run()

			if errAlt != nil {
				// Si ambos intentos fallan, registrar el error
				logger.LogError("REP Error: No se pudo generar la imagen PNG: %v - %s", err, stderr.String())
				return
			}
		}

		logger.LogInfo(" Reporte TREE del disco %s creado exitosamente", disco)
		fmt.Println("DEBUG - Tree Report: Reporte generado con éxito:", rutaPNG)
	} else {
		logger.LogError("REP Error: Id no existe")
	}
}
