package AdminUserGroup

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

/*
se utiliza para iniciar sesión en el sistema en una particion especifica.

login

	-user	(Obligatorio)	Especifica el nombre del usuario que iniciará sesión.
	-pass	(Obligatorio)	Indicará la contraseña del usuario que inicia sesión.
	-id		(Obligatorio)	Indicará el id de la partición montada de la cual van a iniciar sesión.
							De lograr iniciar sesión todas las acciones se realizarán sobre este id.
*/
func Login(parametros []string) string {
	// 1) estructura para devolver respuestas
	// Crear un logger para este comando
	logger := utils.NewLogger("login")
	// Encabezado
	logger.LogInfo("[ LOGIN ]")

	// 2) validar parametros
	var user string //obligatorio
	var pass string //obligatorio
	var id string   //obligatorio. Id de la particion en la que quiero iniciar sesion
	var pathDisco string

	paramCorrectos := true
	userInit := false
	passInit := false
	idInit := false

	if Estructuras.UsuarioActual.Status {
		logger.LogError("ERROR [ LOGIN ]: Ya existe una sesion iniciada, cierre sesion para iniciar otra")
		return logger.GetErrors()
	}

	for _, parametro := range parametros[1:] {

		fmt.Println(" -> Parametro: ", parametro)
		// token Parametro (parametro, valor) --> dos tknParam: ["clave", "valor"]
		tknParam := strings.Split(parametro, "=")

		// si el token parametro no tiene su identificador y valor es un error
		if len(tknParam) != 2 {
			logger.LogError("ERROR [ LOGIN ]: Valor desconocido del parametro, más de 2 tknParam para: %s", tknParam[0])
			return logger.GetErrors()
		}

		//Capturar valores de los parametros
		//ID

		switch strings.ToLower(tknParam[0]) {
		case "user": // nombre del usuario
			user = tknParam[1]
			user = strings.Trim(user, `"`) // Elimina comillas si están presentes

			if user == "" {
				logger.LogError("ERROR [ LOGIN ]: el parametro USER no puede ser nulo")
				paramCorrectos = false
			}
			userInit = true

		case "pass":

			pass = tknParam[1]
			pass = strings.Trim(pass, `"`) // Elimina comillas si están presentes

			if pass == "" {
				logger.LogError("ERROR [ LOGIN ]: el parametro PASS no puede ser nulo")
				paramCorrectos = false
			}
			passInit = true

		case "id": // id de la particion montada
			id = strings.ToUpper(tknParam[1])

			id = strings.Trim(id, `"`) // Elimina comillas si están presentes

			if id != "" {
				//BUsca en struck de particiones montadas el id ingresado
				for _, montado := range Estructuras.Montadas {
					if montado.Id == id {
						pathDisco = montado.PathM
						fmt.Println("El id es correcto y para una particion montada")
						break
					}
				}
				if pathDisco == "" {
					logger.LogError("ERROR [ LOGIN ]: Verificar id, no se encuentra registrado")
					paramCorrectos = false
				}

				idInit = true
			} else {
				logger.LogError("ERROR [ LOGIN ]: el parametro ID no puede ser nulo")
				paramCorrectos = false
			}

		default:
			logger.LogError("ERROR [ LOGIN ]: Parametro desconocido: '%s", string(tknParam[0]))
			paramCorrectos = false
			break
		}
	}

	// 3) validar logica para comando
	if paramCorrectos && idInit && userInit && passInit {

		//abrir el disco que podría contener el id
		disco, err := Acciones.OpenFile(pathDisco)
		if err != nil {
			return logger.GetErrors()
		}

		//cargar el mbr del disco
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(disco, &mbr, 0); err != nil {
			logger.LogError("ERROR [ LOGIN ]: Al intentar leer el mbr del disco de ID: %s", id)
			return logger.GetErrors()
		}

		//cerrar el archivo del disco
		defer disco.Close()

		//Asegurar que el id exista
		index := -1
		for i := 0; i < 4; i++ { // buscar en las posibles 4 particiones
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == id { // id generado luego de montar la particion
				index = i
				break //para que ya no siga recorriendo si ya encontro la particion
			}
		}

		var superBloque SystemFileExt2.Superblock
		// (5,0) significa a la posicion 5 desde el inicio del archivo
		// index es el numero de la particion que esta montad
		// ReadObject(archivo, data interface{}, position int64)
		errSB := Acciones.ReadObject(disco, &superBloque, int64(mbr.Mbr_partitions[index].Part_start))
		if errSB != nil {
			logger.LogError("ERROR [ LOGIN ]:  Particion sin formato, intente formatear la particion con MKFS")
			return logger.GetErrors()
		}

		//Se que el users.txt esta en el inodo 1
		var inodo SystemFileExt2.Inode
		//le agrego una estructura inodo porque busco el inodo 1 (sabemos que aqui esta users.txt)
		// Sumar ambos valores (S_inode_start + tamaño del Inode):
		// Este cálculo te lleva directamente al inicio del inodo 1.
		Acciones.ReadObject(disco, &inodo, int64(superBloque.S_inode_start+int32(binary.Size(SystemFileExt2.Inode{}))))

		//leer datos del users.txt (todos los fileblocks que esten en este inodo (archivo))
		var contenido string

		var fileBlock SystemFileExt2.Fileblock
		for _, item := range inodo.I_block {
			if item != -1 { // -1 es porque no esta puntando a ningun lado ------------ item = cada apuntador de bloque a nodo
				Acciones.ReadObject(disco, &fileBlock, int64(superBloque.S_block_start+(item*int32(binary.Size(SystemFileExt2.Fileblock{})))))
				contenido += string(fileBlock.B_content[:])
			}
		}

		linea := strings.Split(contenido, "\n") // ---> 1, G, root \n 1, U, root, root, 123 \n
		//UID, Tipo, Grupo, Usuario, contraseña

		loginFail := true //para saber si encontro el usuaio
		for _, reg := range linea {
			usuario := strings.Split(reg, ",")

			// deberia ser 5 por:
			// 			1, U, root, root, 123
			if len(usuario) == 5 {
				//que no este borrado logicamente
				if usuario[0] != "0" { // --> 1 <--- , U, root, root, 123
					if usuario[3] == user { // 1, U, root, --> root <--- , 123
						if usuario[4] == pass { // 1, U, root, root, ---> 123 <---
							loginFail = false                 // si se puedo logear
							Estructuras.UsuarioActual.Id = id //id de la particion
							// lineaAnalizar , grupo
							buscarIdGrp(linea, usuario[2], logger) //id del grupo al que pertenece el usuario
							idUsr(usuario[0], logger)              //id del usuario
							Estructuras.UsuarioActual.Nombre = user
							Estructuras.UsuarioActual.Status = true // estado loggeado
							Estructuras.UsuarioActual.PathD = pathDisco
							logger.LogInfo("Inicio de sesion exitoso. \nBienvenido: %s", user)
						} else {
							loginFail = false
							logger.LogError("[ LOGIN ] ERROR: Contraseña incorrecta")
						}
						break
					}
				}
			}
		}

		if loginFail {
			logger.LogError("[ LOGIN ] ERROR: No se encontro el usuario: %s", user)
		}
	} else {
		logger.LogError("ERROR [ LOGIN ] Falta algun parametro obligatorio para la ejecucion del comando ")
	}

	// 4) validar salidas
	// Al final de la función:
	if logger.HasErrors() {
		return logger.GetErrors()
	}
	return logger.GetOutput()

}

func buscarIdGrp(lineaID []string, grupo string, logger *utils.Logger) {
	for _, registro := range lineaID[:len(lineaID)-1] {
		datos := strings.Split(registro, ",")
		if len(datos) == 3 { // GID, TIPO, Grupo
			if datos[2] == grupo { // 1, G, --> root <--
				//convertir a numero
				id, errId := strconv.Atoi(datos[0]) // ---> 1 <---, U, root
				if errId != nil {
					logger.LogError("[ LOGIN ] ERROR: Error desconcocido con el idGrp: %s", datos[0])
					return
				}
				Estructuras.UsuarioActual.IdGrp = int32(id)
				return
			}
		}
	}
}

func idUsr(id string, logger *utils.Logger) {
	idU, errId := strconv.Atoi(id)
	if errId != nil {
		logger.LogError("[ LOGIN ] ERROR: Error desconcocido con el idUsr")
		return
	}
	Estructuras.UsuarioActual.IdUsr = int32(idU)
}
