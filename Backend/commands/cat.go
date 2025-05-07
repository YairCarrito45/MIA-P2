package commands

import (
	"fmt"
	"os"
	"strings"

	"backend/stores"
	"backend/structures"
)

func ParseCat(tokens []string) (string, error) {
	if !stores.Auth.IsAuthenticated() {
		return "", fmt.Errorf("Error: debe iniciar sesión primero.")
	}

	// Convertir tokens a mapa de args
	args := make(map[string]string)
	for _, token := range tokens {
		kv := strings.SplitN(token, "=", 2)
		if len(kv) == 2 {
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			value := strings.Trim(strings.TrimSpace(kv[1]), "\"")
			args[key] = value
		}
	}

	// Obtener superbloque de la sesión activa
	sb, _, path, err := stores.GetMountedPartitionSuperblock(stores.Auth.PartitionID)
	if err != nil {
		return "", err
	}

	// Ejecutar lectura de archivos y devolver resultado
	return ExecuteCatCommand(args, path, *sb)
}

func ExecuteCatCommand(args map[string]string, path string, sb structures.SuperBlock) (string, error) {
	var output strings.Builder

	for i := 1; ; i++ {
		param := fmt.Sprintf("file%d", i)
		filePath, exists := args[param]
		if !exists {
			break
		}

		// Buscar el inodo del archivo
		inodeIndex, err := structures.FindInodeByPath(path, filePath, sb)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error: El archivo '%s' no existe.\n", filePath))
			continue
		}

		// Leer inodo
		var inode structures.Inode
		err = inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
		if err != nil {
			output.WriteString(fmt.Sprintf("Error: No se pudo leer el inodo del archivo '%s'.\n", filePath))
			continue
		}

		// Verificar permisos de lectura
		if !HasReadPermission(inode) {
			output.WriteString(fmt.Sprintf("Error: No tiene permiso de lectura en '%s'.\n", filePath))
			continue
		}

		// Leer contenido del archivo
		content, err := ReadFileContent(path, inode, sb)
		if err != nil {
			output.WriteString(fmt.Sprintf("Error al leer el archivo '%s'.\n", filePath))
			continue
		}

		// Agregar contenido al resultado
		output.WriteString(string(content) + "\n")
	}

	// Si no hubo nada que imprimir
	if output.Len() == 0 {
		return "CAT: No se pudo mostrar ningún archivo.\n", nil
	}

	return "========================== CAT ===============================\n" +
		"Contenido de archivos:\n\n" +
		output.String() +
		"==============================================================\n", nil
}

func HasReadPermission(inode structures.Inode) bool {
	perm := string(inode.I_perm[:])
	username, _, _ := stores.Auth.GetCurrentUser()

	// root siempre tiene permiso
	if username == "root" {
		return true
	}

	var accessChar byte
	if int(inode.I_uid) == stores.Auth.UserID {
		accessChar = perm[0]
	} else if int(inode.I_gid) == stores.Auth.GroupID {
		accessChar = perm[1]
	} else {
		accessChar = perm[2]
	}

	return accessChar == '4' || accessChar == '6' || accessChar == '7'
}

func ReadFileContent(path string, inode structures.Inode, sb structures.SuperBlock) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var content []byte
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		offset := sb.S_block_start + int32(blockIndex)*sb.S_block_size
		var block structures.FileBlock
		err := block.Deserialize(path, int64(offset))
		if err != nil {
			return nil, err
		}
		content = append(content, block.B_content[:]...)
	}

	return content, nil
}
