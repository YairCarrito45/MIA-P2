package commands

import (
	"backend/stores"
	"backend/structures"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// MKGRP estructura del comando
type MKGRP struct {
	name string
}

// ParseMkgrp analiza los parámetros del comando mkgrp
func ParseMkgrp(tokens []string) (string, error) {
	cmd := &MKGRP{}

	// Extraer parámetros con regex
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		value := strings.Trim(kv[1], "\"")

		if key == "-name" {
			cmd.name = value
		} else {
			return "", fmt.Errorf("parámetro no reconocido: %s", key)
		}
	}

	if cmd.name == "" {
		return "", errors.New("el parámetro -name es obligatorio")
	}

	err := commandMkgrp(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Grupo '%s' creado exitosamente.", cmd.name), nil
}

// commandMkgrp ejecuta la lógica del comando
func commandMkgrp(cmd *MKGRP) error {
	// Verificar sesión
	if !stores.Auth.IsAuthenticated() {
		return errors.New("debe iniciar sesión para ejecutar este comando")
	}
	if stores.Auth.Username != "root" {
		return errors.New("solo el usuario root puede crear grupos")
	}

	// Obtener el superbloque y ruta
	sb, _, path, err := stores.GetMountedPartitionSuperblock(stores.Auth.PartitionID)
	if err != nil {
		return err
	}

	// Leer el bloque de users.txt
	usersBlock, err := sb.GetUsersBlock(path)
	if err != nil {
		return err
	}

	// Leer contenido actual
	content := strings.Trim(string(usersBlock.B_content[:]), "\x00")
	lines := strings.Split(content, "\n")

	// Verificar si el grupo ya existe
	for _, line := range lines {
		fields := strings.Split(line, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		if len(fields) >= 3 && fields[1] == "G" && fields[2] == cmd.name {
			return fmt.Errorf("el grupo '%s' ya existe", cmd.name)
		}
	}

	// Calcular nuevo ID
	newID := 1
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 3 && fields[1] == "G" {
			id := strings.TrimSpace(fields[0])
			if id != "0" {
				newID++
			}
		}
	}

	// Agregar nueva línea al contenido
	newLine := fmt.Sprintf("%d,G,%s", newID, cmd.name)
	lines = append(lines, newLine)
	newContent := strings.Join(lines, "\n")

	// Copiar nuevo contenido al bloque
	var newBlock [64]byte
	copy(newBlock[:], []byte(newContent))
	copy(usersBlock.B_content[:], newBlock[:])

	// Obtener inodo de users.txt (siempre es inodo 1)
	inode := &structures.Inode{}
	err = inode.Deserialize(path, int64(sb.S_inode_start+sb.S_inode_size)) // inodo 1
	if err != nil {
		return err
	}

	blockIndex := inode.I_block[0]
	if blockIndex == -1 {
		return fmt.Errorf("el archivo users.txt no tiene bloque asignado")
	}

	// Guardar el nuevo contenido en el disco
	err = usersBlock.Serialize(
		path,
		int64(sb.S_block_start) + int64(blockIndex)*int64(sb.S_block_size),
	)
	
	if err != nil {
		return err
	}

	return nil
}
