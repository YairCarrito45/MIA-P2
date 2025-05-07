package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	utils "backend/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
	
)

type MKDIR struct {
	path string
	p    bool
}

func ParseMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-p(\s|$)`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := strings.Trim(kv[1], "\"")
			cmd.path = value
		case "-p":
			if len(kv) > 1 {
				return "", errors.New("el parámetro -p no debe llevar valor")
			}
			cmd.p = true
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	err := commandMkdir(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MKDIR: Directorio %s creado correctamente.", cmd.path), nil
}

func commandMkdir(mkdir *MKDIR) error {
	if !stores.Auth.IsAuthenticated() {
		return errors.New("no se ha iniciado sesión en ninguna partición")
	}

	partitionID := stores.Auth.GetPartitionID()
	sb, mountedPartition, partitionPath, err := stores.GetMountedPartitionSuperblock(partitionID)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	err = createDirectory(mkdir.path, sb, partitionPath, mountedPartition, mkdir.p)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	return nil
}

func createDirectory(dirPath string, sb *structures.SuperBlock, partitionPath string, mountedPartition *structures.Partition, allowParents bool) error {
	parentDirs, destDir := utils.GetParentDirectories(dirPath)

	if !allowParents {
		exists := sb.DirectoriesExist(partitionPath, parentDirs)
		if !exists {
			return fmt.Errorf("las carpetas padres no existen y no se especificó -p")
		}
	}

	if len(parentDirs) > 0 {
		parentInode := sb.FindInodeByPath(partitionPath, "/"+strings.Join(parentDirs, "/"))
		if parentInode != nil && !utils.HasWritePermission(*parentInode) {
			return fmt.Errorf("no tiene permiso de escritura en la carpeta padre")
		}
	}

	err := sb.CreateFolder(partitionPath, parentDirs, destDir, allowParents)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	err = sb.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}


