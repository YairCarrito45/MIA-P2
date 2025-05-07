package Estructuras

import (
	"Gestor/Acciones"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
----------> MBR <----------

	Cuando se crea un nuevo disco este debe contener un MBR,
	este deberá estar en el primer sector del disco.

atributos:

	mbr_tamano         - int          - Tamaño total del disco en bytes
	mbr_fecha_creacion - time         - Fecha y hora de creación del disco
	mbr_dsk_signature  - int          - Número random, que identifica de forma única a cada disco
	dsk_fit            - char         - Tipo de ajuste de la partición. B (Best), F (First) o W (worst)
	mbr_partitions     - partition[4] - Estructura con información de las 4 particiones
*/
type MBR struct {
	Mbr_tamanio        int32        // Tamaño del DISCO en bytes
	Mbr_creation_date  [19]byte     // Fecha y hora de creación del MBR
	Mbr_disk_signature int32        // Firma del disco (ID)
	Mbr_disk_fit       [1]byte      // Tipo de ajuste
	Mbr_partitions     [4]Partition // Particiones del MBR slice del struct Partition (4)
}

/*
Funcion para escribir el MBR en el Disco:

file *os.File   Apuntador a un archivo abierto
tam int         Tamanio del disco
fit string      Tipo de ajuste de la particion
*/
func EscribirMBR(file *os.File, tam int, fit string) (*os.File, error) {
	//obtener hora para el id
	ahora := time.Now()
	//obtener los segundos y minutos
	segundos := ahora.Second()
	minutos := ahora.Minute()
	//hora := ahora.Hour()

	//fmt.Println(hora, minutos, segundos)

	//concatenar los segundos y minutos como una cadena (de 4 digitos)
	cad := fmt.Sprintf("%02d%02d", segundos, minutos)

	//convertir la cadena a numero en un id temporal
	idTmp, err := strconv.Atoi(cad)
	if err != nil {
		fmt.Println("\t ---> ERROR [ MK DISK - mbr ]: La conversion de fecha a entero para id fue incorrecta")
	}

	fmt.Println("\t[ MK DISK - mbr ] ID:", idTmp)

	// Create a new instance of MBR
	var newMBR MBR
	newMBR.Mbr_tamanio = int32(tam)
	newMBR.Mbr_disk_signature = int32(idTmp)
	copy(newMBR.Mbr_disk_fit[:], fit)
	copy(newMBR.Mbr_creation_date[:], ahora.Format("02/01/2006 15:04:05"))

	// Write object in bin file
	if err := Acciones.WriteObject(file, newMBR, 0); err != nil {
		return nil, err
	}

	PrintMBR(newMBR)
	return file, err

}

// PrintMBRToString convierte la información del MBR a una cadena de texto
func PrintMBRToString(data MBR) string {

	var salidaNormal strings.Builder

	salidaNormal.WriteString("==================================\n")
	salidaNormal.WriteString("            Disco                \n")
	salidaNormal.WriteString("==================================\n")
	salidaNormal.WriteString(fmt.Sprintf("Fecha de creación: %s\n", string(data.Mbr_creation_date[:])))
	salidaNormal.WriteString(fmt.Sprintf("Tipo de ajuste (Fit): %s\n", string(data.Mbr_disk_fit[:])))
	salidaNormal.WriteString(fmt.Sprintf("Tamaño del disco: %d Bytes\n", data.Mbr_tamanio))
	salidaNormal.WriteString(fmt.Sprintf("ID del disco: %d\n", data.Mbr_disk_signature))
	salidaNormal.WriteString("----------------------------------\n")
	salidaNormal.WriteString("          Particiones             \n")
	salidaNormal.WriteString("----------------------------------\n")

	for i := 0; i < 4; i++ {
		salidaNormal.WriteString(fmt.Sprintf("Partición %d:\n", i+1))
		salidaNormal.WriteString(fmt.Sprintf("\tNombre: %s\n", string(data.Mbr_partitions[i].Part_name[:])))
		salidaNormal.WriteString(fmt.Sprintf("\tEstado: %s\n", string(data.Mbr_partitions[i].Part_status[:])))
		salidaNormal.WriteString(fmt.Sprintf("\tTipo:   %s\n", string(data.Mbr_partitions[i].Part_type[:])))
		salidaNormal.WriteString(fmt.Sprintf("\tInicio: %d\n", data.Mbr_partitions[i].Part_start))
		salidaNormal.WriteString(fmt.Sprintf("\tTamaño: %d Bytes\n", data.Mbr_partitions[i].Part_size))
		salidaNormal.WriteString(fmt.Sprintf("\tAjuste: %s\n", string(data.Mbr_partitions[i].Part_fit[:])))
		salidaNormal.WriteString(fmt.Sprintf("\tCorrelativo: %d\n", data.Mbr_partitions[i].Part_correlative))
	}
	salidaNormal.WriteString("----------------------------------\n")

	return salidaNormal.String()
}

// PrintMBR imprime la información del MBR en la consola y también devuelve la misma como string
func PrintMBR(data MBR) string {
	output := PrintMBRToString(data)

	// También imprimimos en la consola para propósitos de debugging
	fmt.Print(output)

	return output
}

func RepGraphviz(data MBR, disco *os.File, logger *utils.Logger) string {
	disponible := int32(0)
	cad := ""
	inicioLibre := int32(binary.Size(data))

	for i := 0; i < 4; i++ {
		if data.Mbr_partitions[i].Part_size > 0 {
			disponible = data.Mbr_partitions[i].Part_start - inicioLibre
			inicioLibre = data.Mbr_partitions[i].Part_start + data.Mbr_partitions[i].Part_size

			if disponible > 0 {
				cad += fmt.Sprintf(" <tr>\n <td bgcolor='#808080' COLSPAN=\"2\"> ESPACIO LIBRE <br/> %d bytes </td> \n </tr> \n", disponible)
			}

			cad += " <tr>\n <td bgcolor='#028f76' COLSPAN=\"2\"> PARTICION </td> \n </tr> \n"
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_status </td> \n <td bgcolor='#aad4cf'> %s </td> \n </tr> \n", string(data.Mbr_partitions[i].Part_status[:]))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_type </td> \n <td bgcolor='#aad4cf'> %s </td> \n </tr> \n", string(data.Mbr_partitions[i].Part_type[:]))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_fit </td> \n <td bgcolor='#aad4cf'> %s </td> \n </tr> \n", string(data.Mbr_partitions[i].Part_fit[:]))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_start </td> \n <td bgcolor='#aad4cf'> %d </td> \n </tr> \n", data.Mbr_partitions[i].Part_start)
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_size </td> \n <td bgcolor='#aad4cf'> %d </td> \n </tr> \n", data.Mbr_partitions[i].Part_size)
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_name </td> \n <td bgcolor='#aad4cf'> %s </td> \n </tr> \n", GetName(string(data.Mbr_partitions[i].Part_name[:])))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#aad4cf'> part_id </td> \n <td bgcolor='#aad4cf'> %s </td> \n </tr> \n", GetId(string(data.Mbr_partitions[i].Part_id[:])))

			if string(data.Mbr_partitions[i].Part_type[:]) == "E" {
				cad += repLogicas(data.Mbr_partitions[i], disco, logger)
			}
		}
	}

	disponible = data.Mbr_tamanio - inicioLibre
	if disponible > 0 {
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#808080' COLSPAN=\"2\"> ESPACIO LIBRE <br/> %d bytes </td> \n </tr> \n", disponible)
	}
	return cad
}

func repLogicas(particion Partition, disco *os.File, logger *utils.Logger) string {
	cad := ""
	var actual EBR

	if err := Acciones.ReadObject(disco, &actual, int64(particion.Part_start)); err != nil {
		logger.LogError("REP ERROR: No se encontro un ebr para reportar logicas")
		return ""
	}

	if actual.EbrP_size != 0 {
		cad += " <tr>\n <td bgcolor='#1e646e' COLSPAN=\"2\"> PARTICION LOGICA </td> \n </tr> \n"
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_status </td> \n <td bgcolor='#66a5ad'> %s </td> \n </tr> \n", string(actual.EbrP_mount[:]))
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#4f98a8'> part_fit </td> \n <td bgcolor='#4f98a8'> %s </td> \n </tr> \n", string(actual.EbrP_fit[:]))
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_start </td> \n <td bgcolor='#66a5ad'> %d </td> \n </tr> \n", actual.EbrP_start)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#4f98a8'> part_size </td> \n <td bgcolor='#4f98a8'> %d </td> \n </tr> \n", actual.EbrP_size)
		cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_name </td> \n <td bgcolor='#66a5ad'> %s </td> \n </tr> \n", GetName(string(actual.EbrP_name[:])))
	}

	for actual.EbrP_next != -1 {
		if err := Acciones.ReadObject(disco, &actual, int64(actual.EbrP_next)); err != nil {
			logger.LogError("REP ERROR: fallo al leer particiones logicas")
			return cad
		}

		if actual.EbrP_size > 0 {
			cad += " <tr>\n <td bgcolor='#1e646e' COLSPAN=\"2\"> PARTICION LOGICA </td> \n </tr> \n"
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_status </td> \n <td bgcolor='#66a5ad'> %s </td> \n </tr> \n", string(actual.EbrP_mount[:]))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#4f98a8'> part_fit </td> \n <td bgcolor='#4f98a8'> %s </td> \n </tr> \n", string(actual.EbrP_fit[:]))
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_start </td> \n <td bgcolor='#66a5ad'> %d </td> \n </tr> \n", actual.EbrP_start)
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#4f98a8'> part_size </td> \n <td bgcolor='#4f98a8'> %d </td> \n </tr> \n", actual.EbrP_size)
			cad += fmt.Sprintf(" <tr>\n <td bgcolor='#66a5ad'> part_name </td> \n <td bgcolor='#66a5ad'> %s </td> \n </tr> \n", GetName(string(actual.EbrP_name[:])))
		}
	}
	return cad
}
