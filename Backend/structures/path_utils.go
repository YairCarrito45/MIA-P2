package structures

import (
	"errors"
	"strings"
)

// FindInodeByPath busca un archivo o carpeta dado su path absoluto (ej: /home/a.txt)
// y retorna el índice del inodo correspondiente.
func FindInodeByPath(diskPath string, absolutePath string, sb SuperBlock) (int32, error) {
	// Limpiar y dividir el path en carpetas
	trimmed := strings.Trim(absolutePath, "/")
	if trimmed == "" {
		return 0, nil // raíz
	}
	parts := strings.Split(trimmed, "/")

	// Comenzamos desde el inodo raíz (posición 0)
	currentInodeIndex := int32(0)

	for _, part := range parts {
		// Leer el inodo actual
		inode := Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(currentInodeIndex*sb.S_inode_size)))
		if err != nil {
			return -1, errors.New("Error al deserializar inodo en ruta")
		}

		// Verificar si es carpeta
		if inode.I_type[0] != '0' {
			return -1, errors.New("Ruta intermedia no es una carpeta")
		}

		found := false

		// Buscar el nombre dentro de los bloques de carpeta
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}

			block := FolderBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+blockIndex*sb.S_block_size))
			if err != nil {
				return -1, errors.New("Error al leer bloque de carpeta")
			}

			for _, entry := range block.B_content {
				entryName := strings.Trim(string(entry.B_name[:]), "\x00")
				if entryName == part && entry.B_inodo != -1 {
					currentInodeIndex = entry.B_inodo
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return -1, errors.New("Archivo o carpeta no encontrado: " + part)
		}
	}

	return currentInodeIndex, nil
}
