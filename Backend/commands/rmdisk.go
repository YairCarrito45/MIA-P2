package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// RMDISK estructura que representa el comando rmdisk con su parámetro
type RMDISK struct {
	path string // Ruta del archivo del disco a eliminar
}

// ParseRmdisk recibe los tokens del comando y ejecuta el proceso
func ParseRmdisk(tokens []string) (string, error) {
	cmd := &RMDISK{}

	// Unir tokens y usar expresión regular para encontrar -path="..." o -path=...
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Procesar parámetros
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Quitar comillas si tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Validar que el path se haya definido
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar eliminación
	err := commandRmdisk(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"========================== RMDISK ===============================\n"+
			"RMDISK: Disco eliminado correctamente\n"+
			"-> Path: %s\n"+
			"=================================================================",
		cmd.path), nil
}

// commandRmdisk elimina el archivo del disco si existe
func commandRmdisk(rmdisk *RMDISK) error {
	// Verificar si el archivo existe
	if _, err := os.Stat(rmdisk.path); os.IsNotExist(err) {
		return fmt.Errorf("el archivo no existe en la ruta especificada: %s", rmdisk.path)
	}

	// Intentar eliminar el archivo
	err := os.Remove(rmdisk.path)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("no tienes permisos para eliminar el archivo: %s", rmdisk.path)
		}
		return fmt.Errorf("no se pudo eliminar el archivo: %v", err)
	}

	return nil
}
