package SystemFileExt2

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"os"
)

// SUPERBLOQUES
type Superblock struct {
	S_filesystem_type   int32    //numero que identifica el sistema de archivos usado //0->no formateada; 2->ext2; 3->ext3
	S_inodes_count      int32    //numero total de inods creados
	S_blocks_count      int32    //numero total de bloques creados
	S_free_blocks_count int32    //numero de bloques libres
	S_free_inodes_count int32    //numero de inodos libres
	S_mtime             [19]byte //ultima fecha en que el sistema fue montado "02/01/2006 15:04:05"
	S_umtime            [19]byte //ultima fecha en que el sistema fue desmontado "02/01/2006 15:04:05"
	S_mnt_count         int32    //numero de veces que se ha montado el sistema
	S_magic             int32    //valor que identifica el sistema de archivos (Sera 0xEF53)
	S_inode_size        int32    //tamaño de la etructura inodo
	S_block_size        int32    //tamaño de la estructura bloque
	S_first_ino         int32    //primer inodo libre
	S_first_blo         int32    //primer bloque libre
	S_bm_inode_start    int32    //inicio del bitmap de inodos
	S_bm_block_start    int32    //inicio del bitmap de bloques
	S_inode_start       int32    //inicio de la tabla de inodos
	S_block_start       int32    //inicio de la tabla de bloques
}

func CrearEXT2(n int32, particion Estructuras.Partition, newSuperBloque Superblock, date string, file *os.File, logger *utils.Logger) bool {
	fmt.Println("Superbloque: ", newSuperBloque)
	fmt.Println("Fecha: ", date)

	//completar los atributos del super bloque. La estructura de la particion formateada es:
	// | Superbloque | Bitmap Inodos | Bitmap Bloques | Inodos | Bloques |

	//tipo del sistema de archivos
	newSuperBloque.S_filesystem_type = 2 //2 -> EXT2; 3 -> EXT3

	//Bitmap Inodos inicia donde termina el superbloque fisicamente (y el superbloque esta al inicio de la particion)
	newSuperBloque.S_bm_inode_start = particion.Part_start + int32(binary.Size(Superblock{}))

	//Bitmap bloques inicia donde termina el de inodos. Se suma n que es el numero de inodos maximo
	newSuperBloque.S_bm_block_start = newSuperBloque.S_bm_inode_start + n

	//Se crea el primer Inodo. Esta al final de los bloques que son 3 veces el numero de inodos
	newSuperBloque.S_inode_start = newSuperBloque.S_bm_block_start + 3*n

	//Se crea el primer bloque, este esta al final de los inodos fisicos
	newSuperBloque.S_block_start = newSuperBloque.S_inode_start + n*int32(binary.Size(Inode{}))

	//Se restan 2 bloques y dos inodos. uno para la carpeta raiz y otro para el archivo users.txt
	//lo que se crea al formatear es /users.txt (la carpeta usa un inodo y el archivo otro)
	newSuperBloque.S_free_inodes_count -= 2
	newSuperBloque.S_free_blocks_count -= 2

	//primer inodo libre
	//newSuperBloque.S_first_ino = newSuperBloque.S_inode_start + 2*int32(binary.Size(Inode{})) //multiplico por 2 porque hay 2 inodos creados
	newSuperBloque.S_first_ino = int32(2)

	//primer bloque libre
	//newSuperBloque.S_first_blo = newSuperBloque.S_block_start + 2*int32(binary.Size(Fileblock{})) //multiplicar por 2 porque hay 2 bloques creados
	newSuperBloque.S_first_blo = int32(2)

	//limpio (formateo) el espacio del bitmap de inodos para evitar inconsistencias
	bmInodeData := make([]byte, n)
	bmInodeErr := Acciones.WriteObject(file, bmInodeData, int64(newSuperBloque.S_bm_inode_start))
	if bmInodeErr != nil {
		logger.LogError("ERROR [ MKFS ]: %s", bmInodeErr)
		return false
	}

	//limpiar (formatear) el espacio del bitmap de bloques para evitar inconsistencias
	bmBlockData := make([]byte, 3*n)
	bmBlockErr := Acciones.WriteObject(file, bmBlockData, int64(newSuperBloque.S_bm_block_start))
	if bmBlockErr != nil {
		fmt.Println("ERROR [ MKFS ]: ", bmInodeErr)
		return false
	}

	/*
		Inicializa todos los inodos:

			Crea un inodo "plantilla" con todos los apuntadores a bloques en -1
			Escribe esta plantilla para todos los inodos posibles
	*/
	var newInode Inode
	for i := 0; i < 15; i++ {
		newInode.I_block[i] = -1
	}

	//creo todos los inodos del sistema de archivos
	for i := int32(0); i < n; i++ {
		err := Acciones.WriteObject(file, newInode, int64(newSuperBloque.S_inode_start+i*int32(binary.Size(Inode{}))))
		if err != nil {
			fmt.Println("ERROR [ MKFS ]: ", err)
			return false
		}
	}

	//Crear todos los bloques de carpeta que se pueden crear
	fileBlocks := make([]Fileblock, 3*n) //lo puedo trabajar asi porque son instancias de la estructura, el inode llevaban valores
	fileBlocksErr := Acciones.WriteObject(file, fileBlocks, int64(newSuperBloque.S_bm_block_start))
	if fileBlocksErr != nil {
		fmt.Println("ERROR [ MKFS ]: ", fileBlocksErr)
		return false
	}

	//Crear el Inode 0
	var Inode0 Inode
	Inode0.I_uid = 1
	Inode0.I_gid = 1
	Inode0.I_size = 0 //por ser carpeta no tiene tamaño como tal. para saber si existe basarse en I_ui/I_gid
	//unica vez que las 3 fechas son iguales
	copy(Inode0.I_atime[:], date)
	copy(Inode0.I_ctime[:], date)
	copy(Inode0.I_mtime[:], date)
	copy(Inode0.I_type[:], "0") //como es raiz es de tipo carpeta
	copy(Inode0.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		Inode0.I_block[i] = -1
	}

	Inode0.I_block[0] = 0 //apunta al bloque 0

	//Crear el folder con la estructura
	// 	. 		| 0   -> actual (a si mismo)
	// 	..      | 0   -> el padre
	//users.txt | 1
	//			|-1

	var folderBlock0 Folderblock //Bloque0 -> carpetas
	folderBlock0.B_content[0].B_inodo = 0
	copy(folderBlock0.B_content[0].B_name[:], ".")
	folderBlock0.B_content[1].B_inodo = 0
	copy(folderBlock0.B_content[1].B_name[:], "..")
	folderBlock0.B_content[2].B_inodo = 1
	copy(folderBlock0.B_content[2].B_name[:], "users.txt")
	folderBlock0.B_content[3].B_inodo = -1

	//Inode1 que es el que contiene el archivo (Bloque 0 apunta a este nuevo inodo)
	var Inode1 Inode
	Inode1.I_uid = 1
	Inode1.I_gid = 1
	Inode1.I_size = int32(binary.Size(Folderblock{}))
	copy(Inode1.I_atime[:], date)
	copy(Inode1.I_ctime[:], date)
	copy(Inode1.I_mtime[:], date)
	copy(Inode1.I_type[:], "1") //es del archivo
	copy(Inode0.I_perm[:], "664")
	for i := int32(0); i < 15; i++ {
		Inode1.I_block[i] = -1
	}

	/*

		No se está creando un archivo "users.txt" físicamente como lo harías con un editor de texto.
		Lo que se está haciendo es escribir datos binarios directamente en el disco que representan
		la estructura del sistema de archivos.

	*/
	//Inode1 apunta al bloque1 (en este caso el bloque1 contiene el archivo)
	Inode1.I_block[0] = 1
	data := "1,G,root\n1,U,root,root,123\n"
	var fileBlock1 Fileblock //Bloque1 -> archivo
	copy(fileBlock1.B_content[:], []byte(data))
	logger.LogInfo("Creado users.txt con los datos : \n%s", data)

	//resumen
	//Inodo 0 -> Bloque 0 -> Inodo1 -> bloque1 (archivo)

	//Crear la carpeta raiz /
	//crear el archivo users.txt

	//fmt.Println("Superbloque: ", newSuperBloque)

	// Escribir el superbloque
	Acciones.WriteObject(file, newSuperBloque, int64(particion.Part_start))

	//escribir el bitmap de inodos
	Acciones.WriteObject(file, byte(1), int64(newSuperBloque.S_bm_inode_start))
	Acciones.WriteObject(file, byte(1), int64(newSuperBloque.S_bm_inode_start+1)) //Se escribieron dos inode

	//escribir el bitmap de bloques (se usaron dos bloques)
	Acciones.WriteObject(file, byte(1), int64(newSuperBloque.S_bm_block_start))
	Acciones.WriteObject(file, byte(1), int64(newSuperBloque.S_bm_block_start+1))

	//escribir inodes
	//Inode0
	Acciones.WriteObject(file, Inode0, int64(newSuperBloque.S_inode_start))
	//Inode1
	Acciones.WriteObject(file, Inode1, int64(newSuperBloque.S_inode_start+int32(binary.Size(Inode{}))))

	//Escribir bloques
	//bloque0
	Acciones.WriteObject(file, folderBlock0, int64(newSuperBloque.S_block_start))
	//bloque1
	Acciones.WriteObject(file, fileBlock1, int64(newSuperBloque.S_block_start+int32(binary.Size(Fileblock{}))))
	// Fin crear EXT2
	return true
}
