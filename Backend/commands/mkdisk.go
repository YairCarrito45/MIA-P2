package commands

import (
	structures "backend/structures"
	utils "backend/utils"
	"errors"        // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"           // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"math/rand"     // Paquete para generar números aleatorios
	"os"            // Paquete para interactuar con el sistema operativo
	"path/filepath" // Paquete para trabajar con rutas de archivos y directorios
	//"regexp"        // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv"       // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings"       // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
	"time"
	
)

// MKDISK estructura que representa el comando mkdisk con sus parámetros
type MKDISK struct {
	size int    // Tamaño del disco
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
}

/*
   mkdisk -size=3000 -unit=K -path=/home/user/Disco1.mia
   mkdisk -size=3000 -path=/home/user/Disco1.mia
   mkdisk -size=5 -unit=M -fit=WF -path="/home/keviin/University/PRACTICAS/MIA_LAB_S2_2024/CLASE03/disks/Disco1.mia"
   mkdisk -size=10 -path="/home/mis discos/Disco4.mia"
*/

func ParseMkdisk(tokens []string) (string, error) {
	cmd := &MKDISK{}

	// Iterar sobre todos los tokens
	for _, token := range tokens {
		lower := strings.ToLower(token)

		if strings.HasPrefix(lower, "-size=") {
			value := strings.SplitN(token, "=", 2)[1]
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size

		} else if strings.HasPrefix(lower, "-unit=") {
			value := strings.SplitN(token, "=", 2)[1]
			value = strings.ToUpper(strings.Trim(value, "\""))
			if value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser K o M")
			}
			cmd.unit = value

		} else if strings.HasPrefix(lower, "-fit=") {
			value := strings.SplitN(token, "=", 2)[1]
			value = strings.ToUpper(strings.Trim(value, "\""))
			if value != "FF" && value != "BF" && value != "WF" {
				return "", errors.New("el ajuste debe ser FF, BF o WF")
			}
			cmd.fit = value

		} else if strings.HasPrefix(lower, "-path=") {
			value := strings.SplitN(token, "=", 2)[1]
			value = strings.Trim(value, "\"")
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value

		} else {
			// ❌ Parámetro no reconocido
			return "", fmt.Errorf("parámetro no reconocido: %s", token)
		}
	}

	// Validaciones obligatorias
	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Defaults
	if cmd.unit == "" {
		cmd.unit = "M"
	}
	if cmd.fit == "" {
		cmd.fit = "FF"
	}

	// Ejecutar creación del disco
	err := commandMkdisk(cmd)
	if err != nil {
		return "", err
	}

	diskName := filepath.Base(cmd.path)

	// Éxito
	return fmt.Sprintf(
		"========================== MKDISK ===============================\n"+
			"MKDISK: Disco creado exitosamente\n"+
			"-> Path: %s\n"+
			"-> Nombre: %s\n"+
			"-> Tamaño: %d%s\n"+
			"-> Fit: %s\n"+
			"=================================================================\n",
		cmd.path, diskName, cmd.size, cmd.unit, cmd.fit), nil
}


func commandMkdisk(mkdisk *MKDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(mkdisk.size, mkdisk.unit)
	if err != nil {
		return fmt.Errorf("Error creando disco: %v", err)
	}

	// Crear el disco con el tamaño proporcionado
	err = createDisk(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating disk:", err)
		return err
	}

	// Crear el MBR con el tamaño proporcionado
	err = createMBR(mkdisk, sizeBytes)
	if err != nil {
		fmt.Println("Error creating MBR:", err)
		return err
	}

	return nil
}

func createDisk(mkdisk *MKDISK, sizeBytes int) error {
	// Crear las carpetas necesarias
	err := os.MkdirAll(filepath.Dir(mkdisk.path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creando directorios:", err)
		return err
	}

	// Crear el archivo binario
	file, err := os.Create(mkdisk.path)
	if err != nil {
		fmt.Println("Error creando el archivo:", err)
		return err
	}
	defer file.Close()

	// Escribir en el archivo usando un buffer de 1 MB
	buffer := make([]byte, 1024*1024) // Crea un buffer de 1 MB
	for sizeBytes > 0 {
		writeSize := len(buffer)
		if sizeBytes < writeSize {
			writeSize = sizeBytes // Ajusta el tamaño de escritura si es menor que el buffer
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err // Devuelve un error si la escritura falla
		}
		sizeBytes -= writeSize // Resta el tamaño escrito del tamaño total
	}
	return nil
}

func createMBR(mkdisk *MKDISK, sizeBytes int) error {
	// Seleccionar el tipo de ajuste
	var fitByte byte
	switch mkdisk.fit {
	case "FF":
		fitByte = 'F'
	case "BF":
		fitByte = 'B'
	case "WF":
		fitByte = 'W'
	default:
		return errors.New("tipo de ajuste inválido en createMBR")
	}

	// Crear el MBR con los valores proporcionados
	mbr := &structures.MBR{
		Mbr_size:           int32(sizeBytes),
		Mbr_creation_date:  float32(time.Now().Unix()),
		Mbr_disk_signature: rand.Int31(),
		Mbr_disk_fit:       [1]byte{fitByte},
		Mbr_partitions: [4]structures.Partition{
			// Inicializó todos los char en N y los enteros en -1 para que se puedan apreciar en el archivo binario.

			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
		},
	}

	/* SOLO PARA VERIFICACIÓN */
	// Imprimir MBR
	fmt.Println("\nMBR creado:")
	mbr.PrintMBR()

	// Serializar el MBR en el archivo
	err := mbr.Serialize(mkdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}