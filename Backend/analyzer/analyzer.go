package analyzer

import (
	"backend/commands"
	"backend/stores"
	"errors"
	"fmt"
	"strings"
)

// Analyzer analiza el comando de entrada y ejecuta la acción correspondiente
func Analyzer(input string) (string, error) {
	// Eliminar espacios en blanco al inicio/final
	input = strings.TrimSpace(input)

	// Ignorar comentarios o líneas vacías
	if input == "" || strings.HasPrefix(input, "#") {
		return "", nil
	}

	// Tokenizar por espacios
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return "", errors.New("no se proporcionó ningún comando válido")
	}

	// Identificar y ejecutar comando
	switch strings.ToLower(tokens[0]) {
	case "mkdisk":
		return commands.ParseMkdisk(tokens[1:])
	case "rmdisk":
		return commands.ParseRmdisk(tokens[1:])
	case "fdisk":
		return commands.ParseFdisk(tokens[1:])
	case "mount":
		return commands.ParseMount(tokens[1:])
	case "mkfs":
		return commands.ParseMkfs(tokens[1:])
	case "rep":
		return commands.ParseRep(tokens[1:])
	case "mkdir":
		return commands.ParseMkdir(tokens[1:])
	case "login":
		return commands.ParseLogin(tokens[1:])
	case "logout":
		return commands.ParseLogout(tokens[1:])	
	case "rmgrp":
		return commands.ParseRmgrp(tokens[1:])
	case "mkgrp":
		return commands.ParseMkgrp(tokens[1:])	
	case "cat":
		return commands.ParseCat(tokens[1:])
		
	
	case "mounted":
		return stores.ShowMountedPartitions(), nil
	default:
		return "", fmt.Errorf("comando desconocido: %s", tokens[0])
	}
}
