package reports

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"backend/structures"
	"backend/utils"
)

// ReportInode genera un reporte de los inodos utilizados y lo guarda en la ruta especificada
func ReportInode(superblock *structures.SuperBlock, diskPath string, path string) error {
	// Crear carpetas destino si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return err
	}

	// Obtener nombres de archivo DOT y de imagen
	dotFileName, outputImage := utils.GetFileNames(path)

	// Abrir disco para leer el bitmap
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %v", err)
	}
	defer file.Close()

	// Iniciar contenido del archivo DOT
	dotContent := `digraph G {
		node [shape=plaintext]
	`

	inodoIndex := 0

	for i := int32(0); i < superblock.S_inodes_count; i++ {
		// Leer bit del bitmap de inodos
		bitmapByte := make([]byte, 1)
		_, err := file.Seek(int64(superblock.S_bm_inode_start+i), 0)
		if err != nil {
			return fmt.Errorf("error al acceder al bitmap de inodos: %v", err)
		}
		_, err = file.Read(bitmapByte)
		if err != nil {
			return fmt.Errorf("error al leer el bitmap de inodos: %v", err)
		}

		// Si el inodo no está en uso, saltarlo
		if bitmapByte[0] != '1' {
			continue
		}

		// Leer estructura del inodo
		inode := &structures.Inode{}
		err = inode.Deserialize(diskPath, int64(superblock.S_inode_start+(i*superblock.S_inode_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar inodo %d: %v", i, err)
		}

		// Formatear fechas
		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)

		// Crear nodo DOT del inodo
		dotContent += fmt.Sprintf(`inode%d [label=<
			<table border="0" cellborder="1" cellspacing="0">
				<tr><td colspan="2"><b>INODO %d</b></td></tr>
				<tr><td>UID</td><td>%d</td></tr>
				<tr><td>GID</td><td>%d</td></tr>
				<tr><td>Size</td><td>%d</td></tr>
				<tr><td>Atime</td><td>%s</td></tr>
				<tr><td>Ctime</td><td>%s</td></tr>
				<tr><td>Mtime</td><td>%s</td></tr>
				<tr><td>Tipo</td><td>%c</td></tr>
				<tr><td>Perm</td><td>%s</td></tr>
				<tr><td colspan="2"><b>Bloques Directos</b></td></tr>`,
			inodoIndex, inodoIndex,
			inode.I_uid, inode.I_gid, inode.I_size,
			atime, ctime, mtime,
			rune(inode.I_type[0]),
			string(inode.I_perm[:]))

		// Bloques directos (0-11)
		for j := 0; j <= 11; j++ {
			dotContent += fmt.Sprintf("<tr><td>%d</td><td>%d</td></tr>", j, inode.I_block[j])
		}

		// Bloques indirectos (12, 13, 14)
		dotContent += fmt.Sprintf(`
			<tr><td colspan="2"><b>Bloque Indirecto</b></td></tr>
			<tr><td>12</td><td>%d</td></tr>
			<tr><td colspan="2"><b>Bloque Indirecto Doble</b></td></tr>
			<tr><td>13</td><td>%d</td></tr>
			<tr><td colspan="2"><b>Bloque Indirecto Triple</b></td></tr>
			<tr><td>14</td><td>%d</td></tr>
			</table>>];
		`, inode.I_block[12], inode.I_block[13], inode.I_block[14])

		// Conexión al siguiente inodo
		if inodoIndex > 0 {
			dotContent += fmt.Sprintf("inode%d -> inode%d;\n", inodoIndex-1, inodoIndex)
		}
		inodoIndex++
	}

	dotContent += "}"

	// Crear archivo .dot
	dotFile, err := os.Create(dotFileName)
	if err != nil {
		return fmt.Errorf("error al crear archivo DOT: %v", err)
	}
	defer dotFile.Close()

	_, err = dotFile.WriteString(dotContent)
	if err != nil {
		return fmt.Errorf("error al escribir archivo DOT: %v", err)
	}

	// Ejecutar Graphviz para generar imagen
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al generar imagen con dot: %v", err)
	}

	fmt.Println("Imagen de inodos generada correctamente:", outputImage)
	return nil
}
