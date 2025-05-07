package structures

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// EBR representa una partición lógica
type EBR struct {
	PartMount byte     // Estado de montaje: '0' (libre) o '1' (montada)
	PartFit   byte     // Tipo de ajuste: 'B', 'F', 'W'
	PartStart int32    // Byte donde inicia la partición lógica
	PartSize  int32    // Tamaño total de la partición en bytes
	PartNext  int32    // Byte donde está el siguiente EBR (-1 si no hay)
	PartName  [16]byte // Nombre de la partición lógica
}

// WriteEBR escribe un EBR en una posición específica del disco
func WriteEBR(path string, ebr *EBR, position int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return err
	}

	return binary.Write(file, binary.LittleEndian, ebr)
}

// ReadEBR lee un EBR desde una posición específica del disco
func ReadEBR(path string, position int64) (EBR, error) {
	var ebr EBR

	file, err := os.Open(path)
	if err != nil {
		return ebr, err
	}
	defer file.Close()

	_, err = file.Seek(position, 0)
	if err != nil {
		return ebr, err
	}

	err = binary.Read(file, binary.LittleEndian, &ebr)
	return ebr, err
}

// PrintEBR imprime los detalles de un EBR (para depuración)
func (ebr *EBR) PrintEBR() {
	fmt.Println("========== EBR ==========")
	fmt.Printf("-> Mount:  %c\n", ebr.PartMount)
	fmt.Printf("-> Fit:    %c\n", ebr.PartFit)
	fmt.Printf("-> Start:  %d\n", ebr.PartStart)
	fmt.Printf("-> Size:   %d\n", ebr.PartSize)
	fmt.Printf("-> Next:   %d\n", ebr.PartNext)
	fmt.Printf("-> Name:   %s\n", strings.Trim(string(ebr.PartName[:]), "\x00"))
	fmt.Println("==========================")
}

// IsFree verifica si el EBR está vacío (sin partición)
func (ebr *EBR) IsFree() bool {
	return ebr.PartSize == -1
}

// MatchesName compara el nombre del EBR con uno dado
func (ebr *EBR) MatchesName(name string) bool {
	return strings.Trim(string(ebr.PartName[:]), "\x00") == name
}


// Deserialize carga un EBR desde una posición específica en disco
func (ebr *EBR) Deserialize(path string, offset int64) error {
	read, err := ReadEBR(path, offset)
	if err != nil {
		return err
	}
	*ebr = read
	return nil
}
