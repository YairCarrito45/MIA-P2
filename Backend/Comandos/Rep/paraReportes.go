package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func reporte_disk(path string, id string, logger *utils.Logger) {
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

		// Calcular el tamaño del MBR
		sizeMBR := int32(binary.Size(mbr))

		// Iniciar el reporte gráfico
		cad := "digraph G {\n"
		cad += "  rankdir=LR;\n" // Orientación horizontal (Left to Right)
		cad += "  node [shape=none];\n"
		cad += "  labelloc=\"t\";\n"
		cad += fmt.Sprintf("  label=\"Reporte de Disco: %s\";\n", disco)

		// Crear una tabla HTML para el diagrama de bloques
		cad += "  diskStructure [label=<\n"
		cad += "    <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" width=\"1000\">\n"
		cad += "      <tr>\n"

		// Obtener el porcentaje que representa el MBR del disco
		porcentajeMBR := float64(sizeMBR) / float64(mbr.Mbr_tamanio) * 100

		// Agregar el MBR a la tabla
		cad += fmt.Sprintf("        <td bgcolor=\"#87CEFA\" width=\"%d\">MBR<br/></td>\n",
			int(porcentajeMBR*10))

		// Posición actual en el disco (después del MBR)
		posActual := sizeMBR

		// Ordenar las particiones por su posición inicial
		type PartInfo struct {
			Indice int
			Inicio int32
			Tamaño int32
		}

		var particiones []PartInfo

		// Preparar el arreglo de particiones a ordenar
		for i := 0; i < 4; i++ {
			if mbr.Mbr_partitions[i].Part_size > 0 {
				particiones = append(particiones, PartInfo{
					Indice: i,
					Inicio: mbr.Mbr_partitions[i].Part_start,
					Tamaño: mbr.Mbr_partitions[i].Part_size,
				})
			}
		}

		// Ordenar por posición de inicio
		sort.Slice(particiones, func(i, j int) bool {
			return particiones[i].Inicio < particiones[j].Inicio
		})

		// Procesar cada partición en orden
		for _, partInfo := range particiones {
			i := partInfo.Indice

			// Si hay espacio libre antes de esta partición
			if mbr.Mbr_partitions[i].Part_start > posActual {
				espacioLibre := mbr.Mbr_partitions[i].Part_start - posActual
				porcentajeLibre := float64(espacioLibre) / float64(mbr.Mbr_tamanio) * 100

				// Agregar espacio libre a la tabla
				cad += fmt.Sprintf("        <td bgcolor=\"#D3D3D3\" width=\"%d\">Libre<br/>%.2f%% del disco</td>\n",
					int(porcentajeLibre*10), porcentajeLibre)
			}

			// Porcentaje que representa la partición actual
			porcentajePart := float64(mbr.Mbr_partitions[i].Part_size) / float64(mbr.Mbr_tamanio) * 100

			// Nombre de la partición
			nombreParticion := Estructuras.GetName(string(mbr.Mbr_partitions[i].Part_name[:]))

			// Tipo de partición
			tipoParticion := string(mbr.Mbr_partitions[i].Part_type[:])

			if tipoParticion == "P" {
				// Partición primaria
				cad += fmt.Sprintf("        <td bgcolor=\"#98FB98\" width=\"%d\">Primaria<br/>%s<br/>%.2f%% del disco</td>\n",
					int(porcentajePart*10), nombreParticion, porcentajePart)
			} else if tipoParticion == "E" {
				// Partición extendida
				cad += fmt.Sprintf("        <td width=\"%d\">\n", int(porcentajePart*10))
				cad += "          <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" width=\"100%\">\n"
				cad += "            <tr>\n"
				cad += fmt.Sprintf("              <td bgcolor=\"#FFFACD\" colspan=\"10\">Extendida<br/>%s<br/>%.2f%% del disco</td>\n",
					nombreParticion, porcentajePart)
				cad += "            </tr>\n"
				cad += "            <tr>\n"

				// Procesar las particiones lógicas
				procesarParticionesLogicas(file, mbr.Mbr_partitions[i], mbr.Mbr_tamanio, &cad)

				cad += "            </tr>\n"
				cad += "          </table>\n"
				cad += "        </td>\n"
			}

			// Actualizar la posición actual
			posActual = mbr.Mbr_partitions[i].Part_start + mbr.Mbr_partitions[i].Part_size
		}

		// Verificar si hay espacio libre al final del disco
		if posActual < mbr.Mbr_tamanio {
			espacioLibre := mbr.Mbr_tamanio - posActual
			porcentajeLibre := float64(espacioLibre) / float64(mbr.Mbr_tamanio) * 100

			// Agregar espacio libre final a la tabla
			cad += fmt.Sprintf("        <td bgcolor=\"#D3D3D3\" width=\"%d\">Libre<br/>%.2f%% del disco</td>\n",
				int(porcentajeLibre*10), porcentajeLibre)
		}

		// Cerrar la tabla y el grafo
		cad += "      </tr>\n"
		cad += "    </table>\n"
		cad += "  >];\n"
		cad += "}\n"

		// Generar el reporte usando la misma estrategia que el reporte MBR
		carpeta := filepath.Dir(path)
		// Procesar la ruta de manera similar a tu función reporte_mbr
		rutaReporte := "." + carpeta + "/" + nombreReporte + ".dot"

		// Usar la misma función que usa tu reporte_mbr, que ya funciona correctamente
		Acciones.RepGraphizMBR(rutaReporte, cad, nombreReporte)
		logger.LogInfo(" Reporte DISK del disco %s creado exitosamente", disco)
	} else {
		logger.LogError("REP Error: Id no existe")
	}
}

// Función para procesar las particiones lógicas dentro de una partición extendida
func procesarParticionesLogicas(disco *os.File, particionExtendida Estructuras.Partition, tamañoTotal int32, cad *string) {
	var ebr Estructuras.EBR

	// Leer el primer EBR
	if err := Acciones.ReadObject(disco, &ebr, int64(particionExtendida.Part_start)); err != nil {
		fmt.Println("REP Error: No se pudo leer el EBR inicial")
		return
	}

	// Posición actual dentro de la partición extendida
	posActual := particionExtendida.Part_start

	// Si el primer EBR tiene una partición lógica
	if ebr.EbrP_size > 0 {
		// Agregar el EBR
		porcentajeEBR := 0.5 // Valor pequeño fijo para el EBR
		*cad += fmt.Sprintf("              <td bgcolor=\"#B0C4DE\" width=\"%d\">EBR</td>\n", int(porcentajeEBR*10))

		// Agregar la partición lógica
		nombreLogica := Estructuras.GetName(string(ebr.EbrP_name[:]))
		porcentajeLogica := float64(ebr.EbrP_size) / float64(tamañoTotal) * 100

		*cad += fmt.Sprintf("              <td bgcolor=\"#ADD8E6\" width=\"%d\">Lógica<br/>%s<br/>%.2f%%</td>\n",
			int(porcentajeLogica*10), nombreLogica, porcentajeLogica)

		posActual = ebr.EbrP_start + ebr.EbrP_size
	}

	// Recorrer la lista enlazada de EBRs
	for ebr.EbrP_next != -1 {
		posAnterior := posActual
		posActual = ebr.EbrP_next

		// Leer el siguiente EBR
		if err := Acciones.ReadObject(disco, &ebr, int64(ebr.EbrP_next)); err != nil {
			fmt.Println("REP Error: Error al leer el siguiente EBR")
			break
		}

		// Si hay espacio libre entre particiones lógicas
		if posActual > posAnterior {
			espacioLibre := posActual - posAnterior
			porcentajeLibre := float64(espacioLibre) / float64(tamañoTotal) * 100

			if porcentajeLibre > 0.1 {
				*cad += fmt.Sprintf("              <td bgcolor=\"#E6E6FA\" width=\"%d\">Libre<br/>%.2f%%</td>\n",
					int(porcentajeLibre*10), porcentajeLibre)
			}
		}

		// Si el EBR actual tiene una partición lógica
		if ebr.EbrP_size > 0 {
			// Agregar el EBR
			porcentajeEBR := 0.5 // Valor pequeño fijo para el EBR
			*cad += fmt.Sprintf("              <td bgcolor=\"#B0C4DE\" width=\"%d\">EBR</td>\n", int(porcentajeEBR*10))

			// Agregar la partición lógica
			nombreLogica := Estructuras.GetName(string(ebr.EbrP_name[:]))
			porcentajeLogica := float64(ebr.EbrP_size) / float64(tamañoTotal) * 100

			*cad += fmt.Sprintf("              <td bgcolor=\"#ADD8E6\" width=\"%d\">Lógica<br/>%s<br/>%.2f%%</td>\n",
				int(porcentajeLogica*10), nombreLogica, porcentajeLogica)

			posActual = ebr.EbrP_start + ebr.EbrP_size
		}
	}

	// Si hay espacio libre al final de la partición extendida
	finExtendida := particionExtendida.Part_start + particionExtendida.Part_size
	if posActual < finExtendida {
		espacioLibre := finExtendida - posActual
		porcentajeLibre := float64(espacioLibre) / float64(tamañoTotal) * 100

		if porcentajeLibre > 0.1 {
			*cad += fmt.Sprintf("              <td bgcolor=\"#E6E6FA\" width=\"%d\">Libre<br/>%.2f%%</td>\n",
				int(porcentajeLibre*10), porcentajeLibre)
		}
	}
}
