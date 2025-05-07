package commands

import (
	structures "backend/structures"
	utils "backend/utils"
	"errors"  // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"     // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv" // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
)

// FDISK estructura que representa el comando fdisk con sus parámetros
type FDISK struct {
	size int    // Tamaño de la partición
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
	typ  string // Tipo de partición (P, E, L)
	name string // Nombre de la partición
}

/*
	fdisk -size=1 -type=L -unit=M -fit=BF -name="Particion3" -path="/home/keviin/University/PRACTICAS/MIA_LAB_S2_2024/CLASEEXTRA/disks/Disco1.mia"
	fdisk -size=300 -path=/home/Disco1.mia -name=Particion1
	fdisk -type=E -path=/home/Disco2.mia -Unit=K -name=Particion2 -size=300
*/

// CommandFdisk parsea el comando fdisk y devuelve una instancia de FDISK
func ParseFdisk(tokens []string) (string, error) {
	cmd := &FDISK{} // Crea una nueva instancia de FDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando fdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			if value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser K o M")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			// Verifica que el tipo sea "P", "E" o "L"
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			// Verifica que el nombre no esté vacío
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = "M"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	// Si no se proporcionó el tipo, se establece por defecto a "P"
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Crear la partición con los parámetros proporcionados
	err := commandFdisk(cmd)
	if err != nil {
		return "", err
	}

	// Devuelve un mensaje de éxito con los detalles de la partición creada
	return fmt.Sprintf(
		"========================== FDISK ===============================\n"+
		"FDISK: Partición creada exitosamente\n"+
		"-> Path: %s\n"+
		"-> Nombre: %s\n"+
		"-> Tamaño: %d%s\n"+
		"-> Tipo: %s\n"+
		"-> Fit: %s\n"+
		"=================================================================\n",
		cmd.path, cmd.name, cmd.size, cmd.unit, cmd.typ, cmd.fit), nil
}

func commandFdisk(fdisk *FDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	if fdisk.typ == "P" {
		// Crear partición primaria
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err)
			return err
		}
	} else if fdisk.typ == "E" {
		err = createExtendedPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err)
			return err
		}	
	} else if fdisk.typ == "L" {
		err = createLogicalPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición lógica:", err)
			return err
		}
	}
	return nil
}



func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Validar máximo 4 particiones primarias + extendidas
	if countUsedPartitions(mbr.Mbr_partitions) >= 4 {
		return errors.New("no se pueden crear más de 4 particiones (primarias + extendidas) en el MBR")
	}

	// Validar que el nombre no esté repetido
	for _, part := range mbr.Mbr_partitions {
		if strings.Trim(string(part.Part_name[:]), "\x00") == fdisk.name {
			return fmt.Errorf("ya existe una partición con el nombre '%s'", fdisk.name)
		}
	}

	// Obtener la primera partición disponible
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay particiones disponibles")
	}

	// --- Añadir esta validación ---
	// Verificar si hay espacio suficiente en el disco
	if int64(startPartition)+int64(sizeBytes) > int64(mbr.Mbr_size) {
		return errors.New("no hay suficiente espacio en el disco para la partición primaria")
	}

	// Crear partición primaria
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// Guardar cambios en el MBR
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		return fmt.Errorf("error serializando el MBR: %v", err)
	}

	fmt.Println("Partición primaria creada exitosamente.")
	return nil
}


/// particion extendida

func createExtendedPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	// Deserializar el MBR del disco
	err := mbr.Deserialize(fdisk.path)
	if err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Validar máximo 4 particiones (primarias + extendida)
	if countUsedPartitions(mbr.Mbr_partitions) >= 4 {
		return errors.New("no se pueden crear más de 4 particiones (primarias + extendida) en el disco")
	}

	// Validar si ya existe una partición extendida
	for _, part := range mbr.Mbr_partitions {
		if part.Part_type[0] == 'E' {
			return errors.New("ya existe una partición extendida en este disco")
		}
	}

	// Validar nombre único
	for _, part := range mbr.Mbr_partitions {
		if strings.Trim(string(part.Part_name[:]), "\x00") == fdisk.name {
			return fmt.Errorf("ya existe una partición con el nombre '%s'", fdisk.name)
		}
	}

	// Obtener la primera partición libre
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay espacio disponible para crear una nueva partición extendida")
	}

	// Validar que la partición realmente quepa en el disco
	if int64(startPartition)+int64(sizeBytes) > int64(mbr.Mbr_size) {
		return errors.New("no hay suficiente espacio en el disco para esta partición extendida")
	}

	// Crear la partición extendida
	availablePartition.CreatePartition(startPartition, sizeBytes, "E", fdisk.fit, fdisk.name)
	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// Crear EBR inicial vacío
	ebr := structures.EBR{
		PartMount: '0',
		PartFit:   availablePartition.Part_fit[0],
		PartStart: int32(startPartition),
		PartSize:  -1,
		PartNext:  -1,
	}
	copy(ebr.PartName[:], "empty")

	err = structures.WriteEBR(fdisk.path, &ebr, int64(startPartition))
	if err != nil {
		return fmt.Errorf("error al escribir el EBR inicial: %v", err)
	}

	// Guardar cambios en el MBR
	err = mbr.Serialize(fdisk.path)
	if err != nil {
		return fmt.Errorf("error serializando el MBR: %v", err)
	}

	fmt.Println("Partición extendida creada exitosamente.")
	return nil
}



func createLogicalPartition(fdisk *FDISK, sizeBytes int) error {
	// 1. Leer el MBR del disco
	var mbr structures.MBR
	if err := mbr.Deserialize(fdisk.path); err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// 2. Encontrar la partición extendida
	var extended structures.Partition
	found := false
	for _, part := range mbr.Mbr_partitions {
		if part.Part_type[0] == 'E' {
			extended = part
			found = true
			break
		}
	}
	
	if !found {
		return errors.New("no existe una partición extendida para alojar la lógica")
	}

	for _, part := range mbr.Mbr_partitions {
		if strings.Trim(string(part.Part_name[:]), "\x00") == fdisk.name {
			return fmt.Errorf("ya existe una partición con el nombre '%s'", fdisk.name)
		}
	}

	// 3. Iniciar recorrido en el primer EBR
	pos := int64(extended.Part_start)
	for {
		ebr, err := structures.ReadEBR(fdisk.path, pos)
		if err != nil {
			return fmt.Errorf("error leyendo EBR: %v", err)
		}

		// Si este EBR tiene PartSize -1, es uno vacío (espacio libre inicial)
		if ebr.PartSize == -1 {
			// Crear nueva partición lógica aquí
			newEBR := structures.EBR{
				PartMount: '0',
				PartFit:   fdisk.fit[0],
				PartStart: int32(pos),
				PartSize:  int32(sizeBytes),
				PartNext:  -1,
			}
			copy(newEBR.PartName[:], fdisk.name)

			err := structures.WriteEBR(fdisk.path, &newEBR, pos)
			if err != nil {
				return fmt.Errorf("error escribiendo nuevo EBR: %v", err)
			}

			fmt.Println("Partición lógica creada exitosamente.")
			return nil
		}

		// Si hay otro EBR después, seguimos recorriendo
		if ebr.PartNext != -1 {
			pos = int64(ebr.PartNext)
			continue
		}

		// Si no hay otro, vamos a crear uno nuevo al final del actual
		newStart := pos + int64(ebr.PartSize)
		newEBR := structures.EBR{
			PartMount: '0',
			PartFit:   fdisk.fit[0],
			PartStart: int32(newStart),
			PartSize:  int32(sizeBytes),
			PartNext:  -1,
		}
		copy(newEBR.PartName[:], fdisk.name)

		// Actualizar el EBR actual con la dirección del siguiente
		ebr.PartNext = int32(newStart)
		if err := structures.WriteEBR(fdisk.path, &ebr, pos); err != nil {
			return fmt.Errorf("error actualizando EBR anterior: %v", err)
		}

		// Escribir el nuevo EBR
		if err := structures.WriteEBR(fdisk.path, &newEBR, newStart); err != nil {
			return fmt.Errorf("error escribiendo nuevo EBR: %v", err)
		}

		fmt.Println("Partición lógica agregada exitosamente.")
		return nil
	}
}


func countUsedPartitions(partitions [4]structures.Partition) int {
	count := 0
	for _, part := range partitions {
		if part.Part_status[0] != 'N' {
			count++
		}
	}
	return count
}