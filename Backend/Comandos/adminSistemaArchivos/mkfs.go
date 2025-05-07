package AdminSistemaArchivos

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	ext2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

/*
Formateo completo de la particion a ext2, creara un archivo en la raiz `user.txt`

	este tendra los usuarios y contrasenias

	mkfs:

		-id    (Obligatorio)   - Indicará el id que se generó con el comando mount.
		-type  (Opcional)      - Indicará que tipo de formateo se realizará.
								Full: en este caso se realizará un formateo completo.


	Recibe un ID de partición montada
	Valida que la partición exista
	Calcula cuántos inodos y bloques caben en la partición
	Crea todas las estructuras necesarias del sistema EXT2:

	Un superbloque
	Bitmaps para inodos y bloques
	Una tabla de inodos
	Una tabla de bloques

	Configura la carpeta raíz (/) y crea un archivo inicial (users.txt)
	Escribe todas estas estructuras al disco
*/
func Mkfs(parametros []string) string {
	// 1) estructura para devolver las respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("mkfs")
	// Encabezado
	logger.LogInfo("[ MKFS ]")

	// 2) validacion de paramtros
	var id string //obligatorio
	paramCorrectos := true
	var pathDico string // necesitamos el path ya que en esa direccion y en la partición (ID) se realizara el comando mkfs
	idInit := false

	for _, parametro := range parametros[1:] {

		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ F DISK ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		switch strings.ToLower(tknParam[0]) {
		case "id":

			id = strings.ToUpper(tknParam[1]) // asignar la entrada al ID

			if id != "" {
				//BUsca en struck de particiones montadas el id ingresado
				for _, montado := range Estructuras.Montadas {
					if montado.Id == id { // buscar el ID en la lista de particiones montadas
						pathDico = montado.PathM // asignar path
					}
				}
				if pathDico == "" {
					logger.LogError("ERROR [ MKFS ]: La particion que se solicita '%s' no existe o no ha sido montada.", id)
					paramCorrectos = false
				}
				idInit = true
			} else {
				logger.LogError("ERROR [ MKFS ]: el parametro ID esta vacio.")
				paramCorrectos = false
			}

		case "type":
			if strings.ToLower(tknParam[1]) != "full" {
				logger.LogError("ERROR [ MKFS ]: Valor de -type desconocido: '%s'", string(tknParam[1]))
				paramCorrectos = false
				break
			}

		default:
			logger.LogError("ERROR [ MKFS ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break

		}
	}

	// 3) logica para mkfs
	if paramCorrectos && idInit {
		// Abrir el Disco de la particion que se quiere formatear
		file, err := Acciones.OpenFile(pathDico)
		if err != nil {
			return logger.GetErrors()
		}

		//Cargar el mbr --> ahí esta la info del disco
		var mbr Estructuras.MBR

		// Read object from bin file
		if err := Acciones.ReadObject(file, &mbr, 0); err != nil {
			return logger.GetErrors()
		}

		// Close bin file
		defer file.Close()

		//Buscar particion con el id solicitado
		formatear := true
		for i := 0; i < 4; i++ {

			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))

			if identificador == id { // el id debe coincidir con el de la particion que se busca

				formatear = false //Si encontro la particion

				//Crear el super bloque que contiene los datos del sistema de archivos. Es similar al mbr en los discos
				var newSuperBloque ext2.Superblock
				Acciones.ReadObject(file, &newSuperBloque, int64(mbr.Mbr_partitions[i].Part_start))

				//Calcular el numero de inodos que caben en la particion. El numero de bloques es el triple de inodos
				//(formula a partir del tamaño de la particion, esta en el enunciado pag. 10)
				//tamaños fisicos: SuperBloque = 92; Inodo = 124; Bloque = 64
				/*

					Calcula cuántos inodos (n) caben en la partición usando la fórmula

					numerador = tamaño_partición - tamaño_superbloque
					denominador = 4 + tamaño_inodo + 3*tamaño_bloque
					n = numerador / denominador

				*/
				numerador := int(mbr.Mbr_partitions[i].Part_size) - binary.Size(ext2.Superblock{})
				denominador := 4 + binary.Size(ext2.Inode{}) + 3*binary.Size(ext2.Fileblock{})

				n := int32(numerador / denominador) //numero de inodos

				//inicializar atributos generales del superbloque
				newSuperBloque.S_blocks_count = int32(3 * n)      //Total de bloques creados (pueden usarse)
				newSuperBloque.S_free_blocks_count = int32(3 * n) //Numero de bloques libre (Todos estan libres por ahora)

				newSuperBloque.S_inodes_count = n      //Total de inodos creados (pueden usarse)
				newSuperBloque.S_free_inodes_count = n //numero de inodos libres (todos estan libres por ahora)

				newSuperBloque.S_inode_size = int32(binary.Size(ext2.Inode{}))
				newSuperBloque.S_block_size = int32(binary.Size(ext2.Fileblock{}))

				//obtener hora de montaje del sistema de archivos
				ahora := time.Now()
				copy(newSuperBloque.S_mtime[:], ahora.Format("02/01/2006 15:04:05"))
				//Si fecha de desmontaje coincide con montaje es porque aun no se monta
				copy(newSuperBloque.S_umtime[:], ahora.Format("02/01/2006 15:04:05"))
				newSuperBloque.S_mnt_count += 1 //Se esta montando por primera vez
				newSuperBloque.S_magic = 0xEF53

				exito := ext2.CrearEXT2(n, mbr.Mbr_partitions[i], newSuperBloque, ahora.Format("02/01/2006 15:04:05"), file, logger)

				if exito {
					//Fin del formateo
					logger.LogInfo("Particion con id %s formateada correctamente, en la fecha: %s", id, string(ahora.Format("02/01/2006 15:04:05")))

					//Si hubiera una sesion iniciada eliminarla
					break //para que ya no siga recorriendo las demas particiones
				}
				break

			}
		}

		if formatear {
			logger.LogError("ERROR [ MKFS ] No se pudo formatear la particion con id %s", id)
			logger.LogError("ERROR [ MKFS ] No existe el id")
		}
	}

	// 4) devolver respuestas
	// Al final de la función:
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()
}
