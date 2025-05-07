package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"fmt"
	"path/filepath"
	"strings"
)

func reporte_sb(path string, id string, logger *utils.Logger) {
	logger.LogInfo("reporte sb")
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
		var sb SystemFileExt2.Superblock
		if err := Acciones.ReadObject(file, &sb, int64(particion.Part_start)); err != nil {
			logger.LogError("REP ERROR: No se pudo leer el superbloque de la partición con ID %s", id)
			return
		}

		// Generar el contenido del reporte en formato DOT para Graphviz
		cad := "digraph { \nnode [ shape=none ] \nTablaReportNodo [ label = < <table border=\"1\"> \n"
		cad += " <tr>\n <td bgcolor='SlateBlue' COLSPAN=\"2\"> Reporte Superbloque </td> \n </tr> \n"
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_filesystem_type </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_filesystem_type)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_inodes_count </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_inodes_count)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_blocks_count </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_blocks_count)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_free_blocks_count </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_free_blocks_count)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_free_inodes_count </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_free_inodes_count)

		// Limpiar strings de fechas para evitar bytes nulos
		mtimeStr := strings.Trim(string(sb.S_mtime[:]), "\x00")
		umtimeStr := strings.Trim(string(sb.S_umtime[:]), "\x00")

		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_mtime </td> \n <td bgcolor='#AFA1D1'> %s </td> \n </tr> \n", mtimeStr)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_umtime </td> \n <td bgcolor='Azure'> %s </td> \n </tr> \n", umtimeStr)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_mnt_count </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_mnt_count)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_magic </td> \n <td bgcolor='Azure'> 0x%X </td> \n </tr> \n", sb.S_magic)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_inode_size </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_inode_size)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_block_size </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_block_size)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_first_ino </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_first_ino)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_first_blo </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_first_blo)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_bm_inode_start </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_bm_inode_start)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_bm_block_start </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_bm_block_start)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#AFA1D1'> s_inode_start </td> \n <td bgcolor='#AFA1D1'> %d </td> \n </tr> \n", sb.S_inode_start)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='Azure'> s_block_start </td> \n <td bgcolor='Azure'> %d </td> \n </tr> \n", sb.S_block_start)
		cad += "</table> > ]\n}"

		// Generar el reporte usando la misma estrategia que otros reportes exitosos
		carpeta := filepath.Dir(path)
		rutaReporte := "." + carpeta + "/" + nombreReporte + ".dot"

		// Usar la función RepGraphizMBR que ya funciona correctamente para otros reportes
		err = Acciones.RepGraphizMBR(rutaReporte, cad, nombreReporte)
		if err != nil {
			logger.LogError("REP ERROR: No se pudo crear el reporte del superbloque: %v", err)
			return
		}

		logger.LogInfo("Reporte Superbloque de la partición %s en el disco %s creado exitosamente", id, disco)
	} else {
		logger.LogError("REP ERROR: La partición con ID %s no está montada", id)
	}
}
