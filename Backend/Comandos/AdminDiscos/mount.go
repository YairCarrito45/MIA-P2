package AdminDiscos

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
Montar una particion del disco.

	ID de cada particion:

	type Partition struct {
		part_status      - char      - Indica si la partición está MONTADA o no (actualizar)

		...
		part_correlative  - int      - Indica el correlativo de la partición este valor será inicialmente -1 hasta que sea montado
		part_id           - char[4]  - Indica el ID de la partición generada al MONTAR esta partición
	}

	ID:
	(+) numero '29'
	(+) numeros de la particion (part_correlative)
	(+) leta

mount

	-name (obligatoria)  nombre de la particion
	-path (obligatoria)  ruta en donde se creará el archivo
*/
func Mount(parametros []string) string {
	// 1) validar parrametros

	// Crear un logger para este comando
	logger := utils.NewLogger("mount")

	// Encabezado
	logger.LogInfo("[ MOUNT ]")

	var name string
	var path string // para la ruta

	paramCorrectos := true // validar que todos los parametros ingresen de forma correcta
	nameInit := false      // para saber si entro el parametro size, false cuando no esta inicializado
	pathInit := false      // para verificar la existencia del path

	// Recorriendo los paramtros
	for _, parametro := range parametros[1:] { // a partir del primero, ya que el primero es la ruta
		fmt.Println(" -> Parametro: ", parametro)
		//logger.LogInfo()

		// token Parametro (parametro, valor) --> dos valores: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ MOUNT ]: Valor desconocido del parametro, mas de 2 valores para: %s", tknParam[0])
			paramCorrectos = false
			break // sale de analizar el parametro y no lo ejecuta
		}

		// id(parametro) - valor
		switch strings.ToLower(tknParam[0]) {
		case "path":
			path = tknParam[1]

			if path != "" {
				// ruta correcta
				path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				path = Acciones.RutaCorrecta(path)

				// nombre del disco
				//path = strings.Trim(path, `"`) // Elimina comillas si están presentes
				ruta := strings.Split(path, "/")
				nombreDisco := ruta[len(ruta)-1] // el ultimo valor de la ruta

				pathInit = true

				_, err := os.Stat(path)
				if os.IsNotExist(err) {
					logger.LogError("ERROR [ MOUNT ]: El disco %s no existe", nombreDisco)
					paramCorrectos = false
					break // Terminar el bucle porque encontramos un nombre único
				}
			} else {
				logger.LogError("ERROR [ F DISK ]: error en ruta")
				paramCorrectos = false
				break
			}
		case "name":
			// Eliminar comillas
			name = strings.ReplaceAll(tknParam[1], "\"", "")

			// Eliminar espacios en blanco al final
			nameValido := strings.TrimSpace(name)
			if nameValido != "" {
				nameInit = true
			} else {
				logger.LogError("ERROR [ MOUNT ]: name parametro obligatorio, no se permite vacio")
				paramCorrectos = false
				break
			}

		default:
			logger.LogError("ERROR [ MOUNT ]: parametro desconocido: %s", tknParam[0])
			paramCorrectos = false
			break
		}
	}

	// 2) realizar logica de montar
	if paramCorrectos {
		if pathInit && nameInit {
			// --------- LOGICA PARA MOUNT ------------------
			// Abrir y cargar el disco
			disco, err := Acciones.OpenFile(path)
			if err != nil {
				logger.LogError("ERROR [ MOUNT ]: No se pudo abrir el disco")
				return logger.GetErrors()
			}

			//Se crea un mbr para cargar el mbr del disco
			var mbr Estructuras.MBR
			//Guardo el mbr leido
			if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
				defer disco.Close()
				logger.LogError("ERROR [ MOUNT ]: No se pudo leer el disco")
				return logger.GetErrors()
			}

			// cerrar el archivo del disco
			defer disco.Close()

			montar := false // para guardar error si no se puede montar

			// menor a 4 por que solo puden haber 4 particiones primarias
			for i := 0; i < 4; i++ {
				// estamos buscando la particion por su nombre
				nombreParticion := Estructuras.GetName(string(mbr.Mbr_partitions[i].Part_name[:]))

				if nombreParticion == name { // el nombre de la particion coincide con el ingresado en el comando
					montar = true
					if string(mbr.Mbr_partitions[i].Part_type[:]) != "E" { // puede ser P o L

						// Part_start = "0" // NO montada
						// part_status = "1" // Montada
						if string(mbr.Mbr_partitions[i].Part_status[:]) != "1" {

							var id string             // part_id
							var nuevaLetra byte = 'A' // A ira aumentando B, C, D , ... , Z
							contador := 1             //
							modificada := false       //para saber si ya hay una particion montada en el disco

							// Se busca si el disco ya tiene particiones montadas
							//Verifica si el path existe dentro de las particiones montadas para calcular la nueva letra
							for k := 0; k < len(Estructuras.Pmontaje); k++ { //Recorre la lista de discos montados
								if Estructuras.Pmontaje[k].MPath == path { // Si encuentra que el path del disco ya está en la lista, significa que ya hay al menos una partición montada en ese disco.
									//Modifica el struct
									Estructuras.Pmontaje[k].Cont = Estructuras.Pmontaje[k].Cont + 1
									contador = int(Estructuras.Pmontaje[k].Cont)
									nuevaLetra = Estructuras.Pmontaje[k].Letter
									modificada = true
									break
								}
							}

							// Si el disco no tiene particiones montadas, asigna una nueva letra
							if !modificada {
								if len(Estructuras.Pmontaje) > 0 {
									// Si hay discos montados en Estructuras.Pmontaje, toma la letra del último disco y la incrementa (A → B → C ...).
									nuevaLetra = Estructuras.Pmontaje[len(Estructuras.Pmontaje)-1].Letter + 1
								}
								Estructuras.AddPathM(path, nuevaLetra, 1)
							}

							id = "78" + strconv.Itoa(contador) + string(nuevaLetra) //Id de particion
							fmt.Println("ID:  Letra ", string(nuevaLetra), " cont ", contador)
							//Agregar al struct de montadas
							Estructuras.AddMontadas(id, path)

							//TODO modificar la particion que se va a montar
							copy(mbr.Mbr_partitions[i].Part_status[:], "1")
							copy(mbr.Mbr_partitions[i].Part_id[:], id)
							mbr.Mbr_partitions[i].Part_correlative = int32(contador)

							//sobreescribir el mbr para guardar los cambios
							if err := Acciones.WriteObject(disco, mbr, 0); err != nil { //Sobre escribir el mbr
								return logger.GetErrors()
							}
							fmt.Println("...............................")
							Estructuras.PrintMBR(mbr)
							fmt.Println("...............................")

							logger.LogInfo("Particion con nombre %s montada correctamente. ID: %s", name, id)
						} else {
							logger.LogError("ERROR [ MOUNT ]: Esta particion YA fue montada previamente")
							return logger.GetErrors()
						}
					} else {
						logger.LogError("ERROR [ MOUNT ]: No se puede montar una particion extendida")
						return logger.GetErrors()
					}
				}
			}

			if !montar {
				logger.LogError("ERROR MOUNT. No se pudo montar la particion %s ", name)
				logger.LogError("ERROR MOUNT. No se encontro la particion")
				return logger.GetErrors()
			}

		} else {
			logger.LogError("ERROR [ MOUNT ]: parametros minimos obligatirios incompletos")
		}

	} else {
		logger.LogError("ERROR [ MOUNT ]: parametros ingresados incorrectamente ")
	}

	// 3) retornar informacion
	// Devolvemos solo la salida normal si no hay errores
	if logger.HasErrors() {
		// Si hay errores, los concatenamos a la salida
		return logger.GetErrors()
	}
	return logger.GetOutput()

}
