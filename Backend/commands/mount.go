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

type MOUNT struct {
	path string
	name string
}

func ParseMount(tokens []string) (string, error) {
	cmd := &MOUNT{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		value = strings.Trim(value, "\"")

		switch key {
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.path == "" || cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -path y/o -name")
	}

	idPartition, err := commandMount(cmd)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"========================== MOUNT ===============================\n"+
		 "MOUNT: Partición Montada\n"+
			"-> Path    : %s\n" +
			"-> Nombre  : %s\n" +
			"-> ID      : %s\n" +
		"=================================================================\n",
		cmd.path, cmd.name, idPartition), nil
	
}

func commandMount(mount *MOUNT) (string, error) {
	var mbr structures.MBR
	if err := mbr.Deserialize(mount.path); err != nil {
		return "", fmt.Errorf("error deserializando el MBR: %w", err)
	}

	// Buscar partición primaria
	partition, _ := mbr.GetPartitionByName(mount.name)


	if partition != nil && partition.Part_type[0] == 'E' {
		return "", errors.New("no se puede montar una partición extendida directamente")
	}

	isLogical := false
	var ebr structures.EBR

	if partition == nil {
		// Buscar lógica
		ebrTemp, err := mbr.GetLogicalPartitionByName(mount.name, mount.path)
		if err != nil {
			return "", errors.New("la partición no existe (ni primaria ni lógica)")
		}
		isLogical = true
		ebr = ebrTemp
	}

	
	

	//valida si ya hay una particion montada
	for _, info := range stores.MountedPartitions {
		if info.Path == mount.path && strings.EqualFold(info.Name, mount.name) {
			return "", fmt.Errorf("la partición '%s' ya está montada", mount.name)
		}
	}

	// Generar ID
	id, correlative, letter, err := generatePartitionID(mount)
	if err != nil {
		return "", err
	}

	// Actualizar datos en estructura RAM
	if isLogical {
		ebr.PartMount = '1'
		ebr.PrintEBR() 
		// NO escribimos en disco, se queda solo en memoria
	} else {
		partition.MountPartition(correlative, id)
	}

	// Guardar en RAM
	stores.MountedPartitions[id] = stores.MountInfo{
		Path:        mount.path,
		Name:        mount.name,
		Letter:      letter,
		Correlative: correlative,
	}

	fmt.Println("Partición montada correctamente. ID:", id)
	return id, nil
}

func generatePartitionID(mount *MOUNT) (string, int, string, error) {
	letter, correlative, err := utils.GetLetterAndPartitionCorrelative(mount.path)
	if err != nil {
		return "", 0, "", err
	}
	lastTwo := stores.Carnet[len(stores.Carnet)-2:]
	id := fmt.Sprintf("%s%d%s", lastTwo, correlative, letter)
	return id, correlative, letter, nil
}
