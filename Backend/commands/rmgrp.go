package commands

import (
	"backend/stores"
	"backend/structures"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// RMGRP representa el comando rmgrp
type RMGRP struct {
	name string
}

// ParseRmgrp analiza los parámetros
func ParseRmgrp(tokens []string) (string, error) {
	cmd := &RMGRP{}

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

	err := commandRmgrp(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Grupo '%s' eliminado correctamente.", cmd.name), nil
}


func commandRmgrp(cmd *RMGRP) error {
	// Verificar sesión
	if !stores.Auth.IsAuthenticated() {
		return errors.New("debe iniciar sesión para ejecutar este comando")
	}
	if stores.Auth.Username != "root" {
		return errors.New("solo el usuario root puede eliminar grupos")
	}

	// Obtener SuperBlock y path
	sb, _, path, err := stores.GetMountedPartitionSuperblock(stores.Auth.PartitionID)
	if err != nil {
		return err
	}

	// Obtener el bloque de users.txt
	usersBlock, err := sb.GetUsersBlock(path)
	if err != nil {
		return err
	}

	content := strings.Trim(string(usersBlock.B_content[:]), "\x00")
	lines := strings.Split(content, "\n")

	found := false
	for i, line := range lines {
		fields := strings.Split(line, ",")
		for j := range fields {
			fields[j] = strings.Trim(fields[j], "\x00 ")
		}
		if len(fields) >= 3 && fields[1] == "G" && fields[2] == cmd.name {
			if fields[0] == "0" {
				return fmt.Errorf("el grupo '%s' ya fue eliminado", cmd.name)
			}
			fields[0] = "0" // Marcar como eliminado
			lines[i] = strings.Join(fields, ",")
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("el grupo '%s' no existe", cmd.name)
	}

	newContent := strings.Join(lines, "\n")

	var newBlock [64]byte
	copy(newBlock[:], []byte(newContent))
	copy(usersBlock.B_content[:], newBlock[:])

	// Obtener el inodo para obtener la posición real del bloque
	inode := &structures.Inode{}
	err = inode.Deserialize(path, int64(sb.S_inode_start+sb.S_inode_size)) // inodo 1
	if err != nil {
		return err
	}

	blockIndex := inode.I_block[0]
	if blockIndex == -1 {
		return fmt.Errorf("el archivo users.txt no tiene bloque asignado")
	}

	err = usersBlock.Serialize(path, int64(sb.S_block_start)+int64(blockIndex)*int64(sb.S_block_size))
	if err != nil {
		return err
	}

	return nil
}
