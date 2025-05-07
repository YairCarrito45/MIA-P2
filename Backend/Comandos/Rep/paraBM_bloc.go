package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"fmt"
	"os"
	"path/filepath"

	"strings"
)

func reporte_bm_block(path string, id string, logger *utils.Logger) {
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
		// Disco a reportar
		tmp := strings.Split(pathDisco, "/")
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

		// Preparar el contenido del reporte
		contenido := fmt.Sprintf("REPORTE BITMAP DE BLOQUES\n")
		contenido += fmt.Sprintf("Disco: %s\n", disco)
		contenido += fmt.Sprintf("ID: %s\n", id)
		contenido += fmt.Sprintf("Total de bloques: %d\n\n", superBloque.S_blocks_count)

		// Leer el bitmap de bloques completo
		contadorLinea := 0
		for i := int32(0); i < superBloque.S_blocks_count; i++ {
			var bite strExt2.Bite
			Acciones.ReadObject(file, &bite, int64(superBloque.S_bm_block_start+i))

			// Añadir bit al reporte (0 o 1)
			contenido += fmt.Sprintf("%d", bite.Val[0])

			contadorLinea++

			// Añadir separación cada 20 registros
			if contadorLinea == 20 {
				contenido += "\n"
				contadorLinea = 0
			}
		}

		// Manejar la ruta de manera similar al reporte de inodos
		carpeta := filepath.Dir(path)
		nombreReporte := filepath.Base(path)

		// Usar ruta relativa para evitar problemas con directorios protegidos
		rutaReporte := "." + carpeta + "/" + nombreReporte

		// Crear el directorio si no existe
		dir := filepath.Dir(rutaReporte)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.LogError("REP Error: No se pudo crear el directorio para el reporte: %v", err)
			return
		}

		// Escribir el archivo de texto
		err = os.WriteFile(rutaReporte, []byte(contenido), 0644)
		if err != nil {
			logger.LogError("REP Error: No se pudo escribir el archivo de reporte: %v", err)
			return
		}

		logger.LogInfo("Reporte BM_BLOCK del disco %s creado exitosamente en %s", disco, path)
	} else {
		logger.LogError("REP Error: Id no existe")
	}
}
