package Estructuras

import (
	"Gestor/Acciones"
	"Gestor/utils"
	"encoding/binary"

	"fmt"
	"os"
	"strings"
)

/*
----------> PARTITION <----------

	Una PARTICION es una división lógica de un disco que
	los sistemas de archivos tratan como una unidad separada.

atributos:

	part_status      - char      - Indica si la partición está MONTADA o no
	part_type        - char      - Indica el tipo de partición, primaria (P) o extendida (E).
	part_fit         - char      - Tipo de ajuste de la partición. B(Best), F (First) o W (worst)
	part_start       - int       - Indica en qué byte del disco inicia la partición
	part_s           - int       - Contiene el tamaño total de la partición en bytes
	part_name        - char[16]  - Nombre de la partición
	part_correlative - int       - Indica el correlativo de la partición este valor será inicialmente -1 hasta que sea montado
	part_id          - char[4]   - Indica el ID de la partición generada al MONTAR esta partición, esto se explicará más adelante
*/
type Partition struct {
	Part_status      [1]byte  // Estado de la partición
	Part_type        [1]byte  // Tipo de partición P / E /L
	Part_fit         [1]byte  // Ajuste de la partición
	Part_start       int32    // Byte de inicio de la partición
	Part_size        int32    // Tamaño de la partición
	Part_name        [16]byte // Nombre de la partición
	Part_correlative int32    // Correlativo de la partición
	Part_id          [4]byte  // ID de la partición
}

// Setear valores de la particion
func (p *Partition) SetInfo(newType string, fit string, newStart int32, newSize int32, name string, correlativo int32) {
	p.Part_size = newSize
	p.Part_start = newStart
	p.Part_correlative = correlativo
	copy(p.Part_name[:], name)
	copy(p.Part_fit[:], fit)
	copy(p.Part_status[:], "0")
	copy(p.Part_type[:], newType)
}

// Metodos de Partition --> para obtener el nombre de la particion
func GetName(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	//Si posicionNulo retorna -1 no hay bytes nulos
	if posicionNulo != -1 {
		//guarda la cadena hasta el primer byte nulo (elimina los bytes nulos)
		nombre = nombre[:posicionNulo]
	}
	return nombre
}

func GetId(nombre string) string {
	//si existe id, no contiene bytes nulos
	posicionNulo := strings.IndexByte(nombre, 0)
	//si posicionNulo  no es -1, no existe id.
	if posicionNulo != -1 {
		nombre = "-"
	}
	return nombre
}

func (p *Partition) GetEnd() int32 {
	return p.Part_start + p.Part_size
}

/*
disco *os.File       ---> disco abierto
typePartition string ---> P (primaria), E (extendida), L (logica)
name string          ---> nombre de la particion
size int             ---> tamanio de la particion
unit int             ---> B (bytes), K (kilobytes), M (megabytes)
*/
func EscribirParticion(disco *os.File, typePartition string, name string, size int, unit int, fit string, logger *utils.Logger) bool {
	//Se crea un mbr para cargar el mbr del disco --> la info
	var mbr MBR

	//Guardo el mbr leido --> y se lee desde la posicion (0, 0) --> lee el MBR que tiene el archivo disco
	if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
		logger.LogError("ERROR [ F DISK ]: No se pudo leer el Disco para hacer la particion")
		return false
	}

	//Si la particion es tipo extendida validar que no exista alguna extendida
	// solo puede haber una por disco
	isPartExtend := false // Indica si se puede usar la particion extendida

	isName := true // Valida si el nombre no se repite (true no se repite)

	if typePartition == "E" {
		for i := 0; i < 4; i++ {
			// leyendo la informacion de las  particiones en el MBR del disco.
			tipo := string(mbr.Mbr_partitions[i].Part_type[:])
			//fmt.Println("tipo ", tipo)
			if tipo != "E" {
				isPartExtend = true
			} else {
				isPartExtend = false
				isName = false
				logger.LogError("ERROR [ F DISK ]: Ya existe una particion extendida")
				logger.LogError("ERROR [ F DISK ]: No se puede crear la nueva particion con nombre: %s", name)
				break
			}
		}
	}

	//verificar si  el nombre existe en las particiones primarias o extendida
	if isName {
		for i := 0; i < 4; i++ {
			nombre := GetName(string(mbr.Mbr_partitions[i].Part_name[:]))
			if nombre == name {
				isName = false
				logger.LogError("ERROR [ F DISK ]: Ya existe la particion : %s", name)
				logger.LogError("ERROR [ F DISK ]: No se puede crear la nueva particion con nombre: %s", name)
				break
			}
		}
	}

	//INGRESO DE PARTICIONES PRIMARIAS Y/O EXTENDIDA (SIN LOGICAS)
	sizeNewPart := size * unit //Tamaño de la nueva particion (tamaño * unidades)
	guardar := false           //Indica si se debe guardar la particion, es decir, escribir en el disco
	var newPart Partition      // nuevo struct tipo particion --> en donde se va guardar la particion

	if (typePartition == "P" || (isPartExtend && typePartition == "E")) && isName { //para que isPartExtend sea true, el tipo de la particion tendra que ser "E"

		// obtener el espacio Real fisico que ocupa el MBR en el disco (el tamaño de la estructura - estructuras.md)
		sizeMBR := int32(binary.Size(mbr))

		//Para manejar los demas ajustes hacer un if del FIT para llamar a la funcion adecuada

		// TODO: validar los ajustes
		// F = primer ajuste;  (BF)
		// B = mejor ajuste;   (FF)
		// else -> peor ajuste (WF --> por defecto)

		//INSERTAR PARTICION (Primer ajuste)
		switch fit {
		case "FF": // busca el primer espacio libre que encuentre
			fmt.Println("fit: ", fit)

		case "BF": // Busca el espacio más pequeño donde quepa
			fmt.Println("fit: ", fit)

		case "WF": // busca el espacio libre mas grande disponible
			// worstfit
			fmt.Println("fit: ", fit)

		default:
			// ERROR
		}

		mbr, newPart = primerAjuste(mbr, typePartition, sizeMBR, int32(sizeNewPart), name, fit, logger) //int32(sizeNewPart) es para castear el int a int32 que es el tipo que tiene el atributo en el struct Partition
		// si la particion es mayor a 0 se puede guardar
		guardar = (newPart.Part_size > 0)

		//escribimos el MBR en el archivo. Lo que no se llegue a escribir en el archivo (aqui) se pierde, es decir, los cambios no se guardan
		if guardar {

			//sobreescribir el mbr --> con la nueva info de la particion
			if err := Acciones.WriteObject(disco, mbr, 0); err != nil {
				return false
			}

			//SI es extendida ademas se agrega el ebr de la particion extendida en el disco
			if isPartExtend {
				var ebr EBR // EBR por "defecto"

				ebr.EbrP_start = newPart.Part_start // el nuevo EBR inicia a apartir de donde incia la nueva particion
				ebr.EbrP_next = -1                  // enlazada a otro ebr

				// escribiendo el EBR en la posicion int64(ebr.EbrP_start)
				if err := Acciones.WriteObject(disco, ebr, int64(ebr.EbrP_start)); err != nil {
					return false
				}
			}

			// para verificar que lo guardo
			var TempMBR2 MBR
			// se lee de nuevo el MBR del disco
			if err := Acciones.ReadObject(disco, &TempMBR2, 0); err != nil {
				return false
			}
			logger.LogInfo("[ F DISK ]: Particion con nombre %s, de tipo: %s creada exitosamente", name, typePartition)
			PrintMBR(TempMBR2)
		} else {
			//Lo podría eliminar pero tendria que modificar en el metodo del ajuste todos los errores para que aparezca el nombre que se intento ingresar como nueva particion
			logger.LogError("ERROR [ F DISK ]: No se puede crear la nueva particion con nombre: %s", name)
			return false
		}
		// TODO: else if para ingreso de particiones logicas
	} else if (typePartition == "L") && isName {
		// logica de particion logica
		// 1. Verificar que exista una partición extendida
		var existeExtendida bool = false
		var partExtendida Partition

		// Buscar la partición extendida en el MBR
		for i := 0; i < 4; i++ {
			tipo := string(mbr.Mbr_partitions[i].Part_type[:])
			if tipo == "E" {
				existeExtendida = true
				partExtendida = mbr.Mbr_partitions[i]
				break
			}
		}

		// Si no existe una partición extendida, no se puede crear la lógica
		if !existeExtendida {
			logger.LogError("ERROR [ F DISK ]: No existe una partición extendida para crear particiones lógicas")
			return false
		}

		// 2. Calcular el tamaño de la partición lógica
		sizeLogica := int32(size * unit)

		// 3. Verificar si el nombre no está repetido entre las particiones lógicas
		nombreRepetido := false

		// Primer EBR está al inicio de la partición extendida
		var ebrActual EBR
		var posEBR int64 = int64(partExtendida.Part_start)

		// Leemos el primer EBR (ya debe existir)
		if err := Acciones.ReadObject(disco, &ebrActual, posEBR); err != nil {
			logger.LogError("ERROR [ F DISK ]: No se pudo leer el EBR inicial de la partición extendida")
			return false
		}

		// Verificar nombres y buscar el último EBR en la cadena
		var ultimoEBR EBR = ebrActual
		var posUltimoEBR int64 = posEBR

		for {
			// Si el EBR actual tiene un nombre (está ocupado) y coincide con el solicitado
			ebrNombre := GetName(string(ebrActual.EbrP_name[:]))
			if ebrNombre == name {
				nombreRepetido = true
				logger.LogError("ERROR [ F DISK ]: Ya existe una partición lógica con el nombre %s", name)
				break
			}

			// Si llegamos al último EBR (next = -1)
			if ebrActual.EbrP_next == -1 {
				ultimoEBR = ebrActual
				posUltimoEBR = posEBR
				break
			}

			// Avanzar al siguiente EBR
			posEBR = int64(ebrActual.EbrP_next)
			if err := Acciones.ReadObject(disco, &ebrActual, posEBR); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al leer un EBR en la cadena de particiones lógicas")
				return false
			}
		}

		if nombreRepetido {
			return false
		}

		// 4. Crear la partición lógica según el caso
		tamañoEBR := int32(binary.Size(ebrActual))

		// Caso especial: Primer EBR sin usar (partición extendida recién creada)
		if ebrActual.EbrP_size == 0 && posEBR == int64(partExtendida.Part_start) {
			// El primer EBR está vacío, podemos usarlo

			// Configurar el EBR
			copy(ebrActual.EbrP_name[:], name)
			copy(ebrActual.EbrP_fit[:], fit)
			ebrActual.EbrP_size = sizeLogica
			// El EBR ya tiene su EbrP_start configurado (desde la creación de la partición extendida)
			// Mantenemos EbrP_next = -1
			copy(ebrActual.EbrP_mount[:], "0") // No montada
			copy(ebrActual.EbrType[:], "L")    // Tipo lógica

			// Escribir el EBR actualizado
			if err := Acciones.WriteObject(disco, ebrActual, posEBR); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al escribir el primer EBR")
				return false
			}

			// Verificar que se haya escrito correctamente
			var checkEBR EBR
			if err := Acciones.ReadObject(disco, &checkEBR, posEBR); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al verificar el EBR actualizado")
				return false
			}

			logger.LogInfo("DEBUG: EBR verificado - Nombre: %s, Size: %d, Next: %d",
				GetName(string(checkEBR.EbrP_name[:])), checkEBR.EbrP_size, checkEBR.EbrP_next)

			logger.LogInfo("[ F DISK ]: Partición lógica %s creada en el primer EBR", name)
			return true
		} else {
			// Caso normal: Agregar después del último EBR en la cadena

			// Calcular el espacio disponible en la partición extendida
			finUltima := ultimoEBR.EbrP_start + ultimoEBR.EbrP_size
			finExtendida := partExtendida.Part_start + partExtendida.Part_size
			espacioDisponible := finExtendida - finUltima

			// Verificar si hay suficiente espacio para la nueva partición
			if espacioDisponible < (sizeLogica + tamañoEBR) {
				logger.LogError("ERROR [ F DISK ]: No hay suficiente espacio en la partición extendida para la partición lógica %s", name)
				return false
			}

			// Crear el nuevo EBR para la partición lógica
			var nuevoEBR EBR

			// Configurar el nuevo EBR
			nuevoEBRPos := int64(finUltima)
			copy(nuevoEBR.EbrP_name[:], name)
			copy(nuevoEBR.EbrP_fit[:], fit)
			copy(nuevoEBR.EbrP_mount[:], "0")
			copy(nuevoEBR.EbrType[:], "L")
			nuevoEBR.EbrP_start = finUltima + tamañoEBR
			nuevoEBR.EbrP_size = sizeLogica
			nuevoEBR.EbrP_next = -1

			// Actualizar el EBR anterior para que apunte al nuevo
			ultimoEBR.EbrP_next = finUltima

			// Escribir el EBR anterior actualizado
			if err := Acciones.WriteObject(disco, ultimoEBR, posUltimoEBR); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al actualizar el EBR anterior")
				return false
			}

			// Escribir el nuevo EBR
			if err := Acciones.WriteObject(disco, nuevoEBR, nuevoEBRPos); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al escribir el nuevo EBR")
				return false
			}

			// Verificar que ambos EBRs se hayan escrito correctamente
			var checkPrevEBR EBR
			var checkNuevoEBR EBR

			if err := Acciones.ReadObject(disco, &checkPrevEBR, posUltimoEBR); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al verificar el EBR anterior")
				return false
			}

			if err := Acciones.ReadObject(disco, &checkNuevoEBR, nuevoEBRPos); err != nil {
				logger.LogError("ERROR [ F DISK ]: Error al verificar el nuevo EBR")
				return false
			}

			logger.LogInfo("DEBUG: EBR anterior - Nombre: %s, Next: %d",
				GetName(string(checkPrevEBR.EbrP_name[:])), checkPrevEBR.EbrP_next)
			logger.LogInfo("DEBUG: Nuevo EBR - Nombre: %s, Size: %d, Start: %d",
				GetName(string(checkNuevoEBR.EbrP_name[:])), checkNuevoEBR.EbrP_size, checkNuevoEBR.EbrP_start)

			logger.LogInfo("[ F DISK ]: Partición lógica %s creada después de la última partición lógica", name)
			return true
		}

	} else {
		logger.LogError("ERROR [ F DISK ]: Parametro -type= %s no valido para crear la particion de nombre: %s", typePartition, name)
		return false
	}

	return true
}

func primerAjuste(mbr MBR, typee string, sizeMBR int32, sizeNewPart int32, name string, fit string, logger *utils.Logger) (MBR, Partition) {

	var newPart Partition // struct de particion
	var noPart Partition  //para revertir el set info (simula volverla null)

	// TODO: agragar el correlativo -1

	//PARTICION 1 (libre) - (size = 0 no se ha creado) caso1
	if mbr.Mbr_partitions[0].Part_size == 0 {
		newPart.SetInfo(typee, fit, sizeMBR, sizeNewPart, name, 1) // nueva particion
		if mbr.Mbr_partitions[1].Part_size == 0 {
			if mbr.Mbr_partitions[2].Part_size == 0 {
				//caso particion 4 (no existe)
				if mbr.Mbr_partitions[3].Part_size == 0 {
					//859 <= 1024 - 165
					// validar que el tamanio de la particion quepa
					if sizeNewPart <= mbr.Mbr_tamanio-sizeMBR { // tamanio del disco - tamanio de la estructura MBR
						mbr.Mbr_partitions[0] = newPart
					} else {
						newPart = noPart // regresa a un case Particion vacio
						logger.LogError("ERROR [FDISK]: Espacio insuficiente para nueva particion")
					}
				} // else caso 2
			}
		}
		//Fin de 1 no existe

		//PARTICION 2 (no existe)
		/*
			part0 part1 part2 part3
			1	  0		0	  0
		*/
	} else if mbr.Mbr_partitions[1].Part_size == 0 {
		//Si no hay espacio antes de particion 1
		newPart.SetInfo(typee, fit, mbr.Mbr_partitions[0].GetEnd(), sizeNewPart, name, 2) //el nuevo inicio es donde termina 1
		if mbr.Mbr_partitions[2].Part_size == 0 {
			if mbr.Mbr_partitions[3].Part_size == 0 {
				if sizeNewPart <= mbr.Mbr_tamanio-newPart.Part_start {
					mbr.Mbr_partitions[1] = newPart
				} else {
					newPart = noPart
					logger.LogError("ERROR [FDISK]: Espacio insuficiente para nueva particion")
				}
			}
		}
		//Fin particion 2 no existe

		//PARTICION 3
		/*
			part0 part1 part2 part3
			1	  1		0	  0
		*/
	} else if mbr.Mbr_partitions[2].Part_size == 0 {
		//despues de 2
		newPart.SetInfo(typee, fit, mbr.Mbr_partitions[1].GetEnd(), sizeNewPart, name, 3)
		if mbr.Mbr_partitions[3].Part_size == 0 {
			if sizeNewPart <= mbr.Mbr_tamanio-newPart.Part_start {
				mbr.Mbr_partitions[2] = newPart
			} else {
				newPart = noPart
				logger.LogError("ERROR [FDISK]: Espacio insuficiente para nueva particion")
			}
		}
		//Fin particion 3

		//PARTICION 4
		/*
			part0 part1 part2 part3
			1	  1		1	  0
		*/
	} else if mbr.Mbr_partitions[3].Part_size == 0 {
		if sizeNewPart <= mbr.Mbr_tamanio-mbr.Mbr_partitions[2].GetEnd() {
			//despues de 3
			newPart.SetInfo(typee, fit, mbr.Mbr_partitions[2].GetEnd(), sizeNewPart, name, 4)
			mbr.Mbr_partitions[3] = newPart
		} else {
			newPart = noPart
			logger.LogError("ERROR [FDISK]: Espacio insuficiente")
		}
		//Fin particion 4
		/*
			part0 part1 part2 part3
			1	  1		1	  1
		*/
	} else {
		newPart = noPart
		logger.LogError("ERROR [FDISK]: Particiones primarias y/o extendidas ya no disponibles")
	}

	return mbr, newPart
}
