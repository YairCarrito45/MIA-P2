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

func reporte_inode(path string, id string, logger *utils.Logger) {
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

		// Iniciar la cadena del reporte - Usamos tabla HTML para mayor simplicidad
		cad := "digraph { \n"
		cad += "  node [ shape=none ] \n"
		cad += "  rankdir=TB;\n"
		cad += "  TablaReportNodo [ label = < <table border=\"1\"> \n"
		cad += "    <tr>\n"
		cad += "      <td bgcolor='SlateBlue' COLSPAN=\"2\"> Reporte de Inodos </td> \n"
		cad += "    </tr> \n"

		// Leer el bitmap de inodos para saber cuáles están en uso
		var inodosEnUso []int32

		// Leer el bitmap de inodos completo
		for i := int32(0); i < superBloque.S_inodes_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_inode_start+i))

			if bite.Val[0] == 1 {
				// Este inodo está en uso, lo agregamos a la lista
				inodosEnUso = append(inodosEnUso, i)
			}
		}

		logger.LogInfo("Se encontraron %d inodos en uso", len(inodosEnUso))

		// Generar el reporte para cada inodo en uso
		for _, idInodo := range inodosEnUso {
			// Leer el inodo
			var inodo strExt2.Inode
			Acciones.ReadObject(file, &inodo, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(strExt2.Inode{})))))

			// Determinar el tipo de inodo (carpeta o archivo)
			tipoInodo := "Carpeta"
			if string(inodo.I_type[:]) == "1" {
				tipoInodo = "Archivo"
			}

			// Añadir encabezado para este inodo
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#4682B4' COLSPAN=\"2\"> Inodo %d (%s) </td> \n", idInodo, tipoInodo)
			cad += fmt.Sprintf("    </tr> \n")

			// Información del propietario
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> i_uid </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> %d </td> \n", inodo.I_uid)
			cad += fmt.Sprintf("    </tr> \n")

			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> i_gid </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> %d </td> \n", inodo.I_gid)
			cad += fmt.Sprintf("    </tr> \n")

			// Tamaño
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> i_size </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> %d </td> \n", inodo.I_size)
			cad += fmt.Sprintf("    </tr> \n")

			// Fechas (limpiando bytes nulos)
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> i_atime </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> %s </td> \n", strings.Trim(string(inodo.I_atime[:]), "\x00"))
			cad += fmt.Sprintf("    </tr> \n")

			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> i_ctime </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> %s </td> \n", strings.Trim(string(inodo.I_ctime[:]), "\x00"))
			cad += fmt.Sprintf("    </tr> \n")

			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> i_mtime </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> %s </td> \n", strings.Trim(string(inodo.I_mtime[:]), "\x00"))
			cad += fmt.Sprintf("    </tr> \n")

			// Bloques directos
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> i_block </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> [")

			// Mostrar los apuntadores a bloques directos
			for i := 0; i < 12; i++ {
				if inodo.I_block[i] != -1 {
					cad += fmt.Sprintf(" %d", inodo.I_block[i])
				}
			}

			cad += fmt.Sprintf(" ] </td> \n")
			cad += fmt.Sprintf("    </tr> \n")

			// Bloques indirectos si existen
			if inodo.I_block[12] != -1 || inodo.I_block[13] != -1 || inodo.I_block[14] != -1 {
				cad += fmt.Sprintf("    <tr>\n")
				cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> i_block_indirect </td> \n")
				cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> [")

				if inodo.I_block[12] != -1 {
					cad += fmt.Sprintf(" Simple: %d", inodo.I_block[12])
				}

				if inodo.I_block[13] != -1 {
					cad += fmt.Sprintf(" Doble: %d", inodo.I_block[13])
				}

				if inodo.I_block[14] != -1 {
					cad += fmt.Sprintf(" Triple: %d", inodo.I_block[14])
				}

				cad += fmt.Sprintf(" ] </td> \n")
				cad += fmt.Sprintf("    </tr> \n")
			}

			// Tipo y permisos
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> i_type </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#E6E6FA'> %s </td> \n", tipoInodo)
			cad += fmt.Sprintf("    </tr> \n")

			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> i_perm </td> \n")
			cad += fmt.Sprintf("      <td bgcolor='#F0F8FF'> %s </td> \n", strings.Trim(string(inodo.I_perm[:]), "\x00"))
			cad += fmt.Sprintf("    </tr> \n")

			// Separador entre inodos
			cad += fmt.Sprintf("    <tr>\n")
			cad += fmt.Sprintf("      <td bgcolor='#D3D3D3' COLSPAN=\"2\"> </td> \n")
			cad += fmt.Sprintf("    </tr> \n")
		}

		// Cerrar la estructura del reporte
		cad += "  </table> > ]\n"
		cad += "}\n"

		// Generar el reporte
		carpeta := filepath.Dir(path)
		rutaReporte := "." + carpeta + "/" + nombreReporte + ".dot"

		// Escribir directamente el archivo DOT para evitar errores en RepGraphizMBR
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

		logger.LogInfo(" Reporte INODE del disco %s creado exitosamente", disco)
	} else {
		logger.LogError("REP Error: Id no existe")
	}
}
