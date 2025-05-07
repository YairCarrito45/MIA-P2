package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ReportDISK(mbr *structures.MBR, reportPath string, diskPath string) error {
	if err := utils.CreateParentDirs(reportPath); err != nil {
		return err
	}

	dotFileName, outputImage := utils.GetFileNames(reportPath)
	totalSize := float64(mbr.Mbr_size)
	lastByte := int64(utils.MBRSize())

	var dotContent strings.Builder
	dotContent.WriteString(`digraph G {
	rankdir=LR;
	node [shape=record, style="filled", fontname="Arial"];
	Disco[label="MBR`)

	var parts []utils.PartPos
	colorLines := new(strings.Builder) // Aquí agregamos las líneas de color

	logicalCount := 0
	freeCount := 0
	primaryCount := 0
	ebrCount := 0

	for _, part := range mbr.Mbr_partitions {
		if part.Part_size > 0 {
			name := strings.TrimRight(string(part.Part_name[:]), "\x00")
			parts = append(parts, utils.PartPos{
				Start: int64(part.Part_start),
				End:   int64(part.Part_start) + int64(part.Part_size),
				Type:  part.Part_type[0],
				Name:  name,
			})
		}
	}
	utils.SortPartsByStart(parts)

	for _, p := range parts {
		if p.Start > lastByte {
			freePerc := ((float64(p.Start - lastByte)) / totalSize) * 100
			dotContent.WriteString(fmt.Sprintf("|<f%d> Libre (%.2f%%)", freeCount, freePerc))
			colorLines.WriteString(fmt.Sprintf("Disco:f%d [fillcolor=\"#DDDDDD\"];\n", freeCount))
			freeCount++
		}

		partPerc := ((float64(p.End - p.Start)) / totalSize) * 100

		if p.Type == 'E' {
			dotContent.WriteString(fmt.Sprintf("|{ <e%d> Extendida", primaryCount))
			colorLines.WriteString(fmt.Sprintf("Disco:e%d [fillcolor=\"#99CCFF\"];\n", primaryCount))

			current := p.Start
			for current < p.End && current != -1 {
				ebr, err := structures.ReadEBR(diskPath, current)
				if err != nil || ebr.PartSize <= 0 {
					break
				}

				ebrEnd := int64(ebr.PartStart) + int64(ebr.PartSize)
				ebrPerc := ((float64(ebrEnd - current)) / totalSize) * 100

				dotContent.WriteString(fmt.Sprintf("|<ebr%d> EBR", ebrCount))
				colorLines.WriteString(fmt.Sprintf("Disco:ebr%d [fillcolor=\"#FFFF99\"];\n", ebrCount))

				dotContent.WriteString(fmt.Sprintf("|<log%d> Lógica (%.2f%%)", logicalCount, ebrPerc))
				colorLines.WriteString(fmt.Sprintf("Disco:log%d [fillcolor=\"#99FF99\"];\n", logicalCount))

				if ebr.PartNext != -1 && ebrEnd < int64(ebr.PartNext) {
					librePerc := ((float64(int64(ebr.PartNext)-ebrEnd)) / totalSize) * 100
					dotContent.WriteString(fmt.Sprintf("|<flog%d> Libre (%.2f%%)", logicalCount, librePerc))
					colorLines.WriteString(fmt.Sprintf("Disco:flog%d [fillcolor=\"#EEEEEE\"];\n", logicalCount))
				}

				if ebr.PartNext == -1 {
					break
				}
				current = int64(ebr.PartNext)
				logicalCount++
				ebrCount++
			}
			dotContent.WriteString("}")
		} else {
			dotContent.WriteString(fmt.Sprintf("|<p%d> Primaria (%.2f%%)", primaryCount, partPerc))
			colorLines.WriteString(fmt.Sprintf("Disco:p%d [fillcolor=\"#FF9999\"];\n", primaryCount))
		}

		lastByte = p.End
		primaryCount++
	}

	if lastByte < int64(totalSize) {
		finalLibrePerc := ((totalSize - float64(lastByte)) / totalSize) * 100
		dotContent.WriteString(fmt.Sprintf("|<f%d> Libre (%.2f%%)", freeCount, finalLibrePerc))
		colorLines.WriteString(fmt.Sprintf("Disco:f%d [fillcolor=\"#DDDDDD\"];\n", freeCount))
	}

	// Cierra el nodo
	dotContent.WriteString(`"];
`)
	dotContent.WriteString(colorLines.String())
	dotContent.WriteString("}")

	// Escribir archivo .dot
	file, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error creando DOT: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(dotContent.String()); err != nil {
		return fmt.Errorf("error escribiendo DOT: %v", err)
	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error Graphviz: %v", err)
	}

	fmt.Println("Reporte DISK generado exitosamente en:", outputImage)
	return nil
}
