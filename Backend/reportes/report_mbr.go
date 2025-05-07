package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ReportMBR(mbr *structures.MBR, reportPath string, diskPath string) error {
	if err := utils.CreateParentDirs(reportPath); err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(reportPath)

	// Colores mejorados en esquema moderno y profesional
	dotContent := fmt.Sprintf(`digraph G {
	bgcolor="#f8f9fa"
	node [shape=plaintext fontname="Arial"]
	tabla [label=<
	<table border="0" cellborder="1" cellspacing="0" cellpadding="6" bgcolor="white">
	<tr><td colspan="2" bgcolor="#2c3e50" align="center"><font color="white"><b>REPORTE DE MBR</b></font></td></tr>
	<tr><td bgcolor="#ecf0f1">mbr_tamano</td><td>%d</td></tr>
	<tr><td bgcolor="#ecf0f1">mbr_fecha_creacion</td><td>%s</td></tr>
	<tr><td bgcolor="#ecf0f1">mbr_disk_signature</td><td>%d</td></tr>
	`, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	partIndex := 1 // contador de particiones principales

	for _, part := range mbr.Mbr_partitions {
		if part.Part_size <= 0 {
			continue
		}

		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])

		dotContent += fmt.Sprintf(`
		<tr><td colspan="2" bgcolor="#3498db" align="center"><font color="white"><b>Partición %d</b></font></td></tr>
		<tr><td bgcolor="#e8f4fc">part_status</td><td>%c</td></tr>
		<tr><td bgcolor="#e8f4fc">part_type</td><td>%c</td></tr>
		<tr><td bgcolor="#e8f4fc">part_fit</td><td>%c</td></tr>
		<tr><td bgcolor="#e8f4fc">part_start</td><td>%d</td></tr>
		<tr><td bgcolor="#e8f4fc">part_size</td><td>%d</td></tr>
		<tr><td bgcolor="#e8f4fc">part_name</td><td>%s</td></tr>
		`, partIndex, partStatus, partType, partFit, part.Part_start, part.Part_size, partName)

		partIndex++

		// Si es extendida, recorrer EBRs
		if (partType == 'E' || partType == 'e') {
			current := int64(part.Part_start)
			logicalIndex := 1
			found := false

			for current != -1 {
				ebr, err := structures.ReadEBR(diskPath, current)
				if err != nil {
					break
				}

				if ebr.PartSize > 0 {
					found = true

					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					ebrFit := rune(ebr.PartFit)

					dotContent += fmt.Sprintf(`
					<tr><td colspan="2" bgcolor="#9b59b6" align="center"><font color="white"><b>Partición Lógica %d</b></font></td></tr>
					<tr><td bgcolor="#f4ecf7">part_fit</td><td>%c</td></tr>
					<tr><td bgcolor="#f4ecf7">part_start</td><td>%d</td></tr>
					<tr><td bgcolor="#f4ecf7">part_size</td><td>%d</td></tr>
					<tr><td bgcolor="#f4ecf7">part_next</td><td>%d</td></tr>
					<tr><td bgcolor="#f4ecf7">part_name</td><td>%s</td></tr>
					`, logicalIndex, ebrFit, ebr.PartStart, ebr.PartSize, ebr.PartNext, ebrName)

					logicalIndex++
				}

				if ebr.PartNext == -1 {
					break
				}
				current = int64(ebr.PartNext)
			}

			if !found {
				dotContent += `<tr><td colspan="2" align="center" bgcolor="#f8f9fa"><i>No hay particiones lógicas</i></td></tr>`
			}
		}
	}

	dotContent += "</table>>]; }"

	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error creando archivo DOT: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(dotContent); err != nil {
		return fmt.Errorf("error escribiendo contenido DOT: %v", err)
	}

	// Mejorar calidad de imagen
	cmd := exec.Command("dot", "-Tpng", "-Gdpi=200", dotFileName, "-o", outputImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error ejecutando Graphviz: %v", err)
	}

	fmt.Println("Reporte MBR generado:", outputImage)
	return nil
}