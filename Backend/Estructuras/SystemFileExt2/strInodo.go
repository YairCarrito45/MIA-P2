package SystemFileExt2

import (
	"Gestor/Acciones"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// INODO
type Inode struct {
	I_uid   int32     //ID del usuario propietario del archivo o carpeta
	I_gid   int32     //ID del grupo al que pertenece el archivo o carpeta
	I_size  int32     //tamaño del archivo en bytes
	I_atime [16]byte  //ultima fecha que se leyó el inodo sin modificarlo "02/01/2006 15:04"
	I_ctime [16]byte  //fecha en que se creo el inodo "02/01/2006 15:04"
	I_mtime [16]byte  //ultima fecha en la que se modifica el inodo "02/01/2006 15:04"
	I_block [15]int32 //-1 si no estan usados. los valores del arreglo son: primeros 12 -> bloques directo;: 13 -> bloque simple indirecto; 14->bloque doble indirecto; 15 -> bloque triple indirecto
	I_type  [1]byte   //1 -> archivo; 0 -> carpeta
	I_perm  [3]byte   //permisos del usuario o carpeta
}

func BuscarInodo(idInodo int32, path string, superBloque Superblock, file *os.File) int32 {
	fmt.Println("Debug - BuscarInodo: idInodo =", idInodo, "path =", path)

	//Dividir la ruta por cada /
	stepsPath := strings.Split(path, "/")
	//el arreglo vendra [ ,val1, val2] por lo que me corro una posicion
	tmpPath := stepsPath[1:]
	//fmt.Println("Ruta actual ", tmpPath)
	fmt.Println("Debug - BuscarInodo: tmpPath =", tmpPath)

	//cargo el inodo a partir del cual voy a buscar
	var Inode0 Inode
	Acciones.ReadObject(file, &Inode0, int64(superBloque.S_inode_start+(idInodo*int32(binary.Size(Inode{})))))
	//Recorrer los bloques directos (carpetas/archivos) en la raiz
	var folderBlock Folderblock
	for i := 0; i < 12; i++ {
		idBloque := Inode0.I_block[i]
		if idBloque != -1 {
			Acciones.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(Folderblock{})))))
			//Recorrer el bloque actual buscando la carpeta/archivo en la raiz
			for j := 2; j < 4; j++ {
				//apuntador es el apuntador del bloque al inodo (carpeta/archivo), si existe es distinto a -1
				apuntador := folderBlock.B_content[j].B_inodo
				if apuntador != -1 {
					pathActual := GetB_name(string(folderBlock.B_content[j].B_name[:]))
					if tmpPath[0] == pathActual {
						//buscarInodo(apuntador, ruta[1:], path, superBloque, iSuperBloque, file, r)
						if len(tmpPath) > 1 {
							return buscarIrecursivo(apuntador, tmpPath[1:], superBloque.S_inode_start, superBloque.S_block_start, file)
						} else {
							return apuntador
						}
					}
				}
			}
		}
	}
	//agregar busqueda en los apuntadores indirectos
	//i=12 -> simple; i=13 -> doble; i=14 -> triple
	//Si no encontro nada retornar 0 (la raiz)
	return idInodo
}

// Buscar inodo de forma recursiva
func buscarIrecursivo(idInodo int32, path []string, iStart int32, bStart int32, file *os.File) int32 {
	//cargo el inodo actual
	var inodo Inode
	Acciones.ReadObject(file, &inodo, int64(iStart+(idInodo*int32(binary.Size(Inode{})))))

	//Nota: el inodo tiene tipo. No es necesario pero se podria validar que sea carpeta
	//recorro el inodo buscando la siguiente carpeta
	var folderBlock Folderblock
	for i := 0; i < 12; i++ {
		idBloque := inodo.I_block[i]
		if idBloque != -1 {
			Acciones.ReadObject(file, &folderBlock, int64(bStart+(idBloque*int32(binary.Size(Folderblock{})))))
			//Recorrer el bloque buscando la carpeta actua
			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				if apuntador != -1 {
					pathActual := GetB_name(string(folderBlock.B_content[j].B_name[:]))
					if path[0] == pathActual {
						if len(path) > 1 {
							//sin este if path[1:] termina en un arreglo de tamaño 0 y retornaria -1
							return buscarIrecursivo(apuntador, path[1:], iStart, bStart, file)
						} else {
							//cuando el arreglo path tiene tamaño 1 esta en la carpeta que busca
							return apuntador
						}
					}
				}
			}
		}
	}
	//agregar busqueda en los apuntadores indirectos
	//i=12 -> simple; i=13 -> doble; i=14 -> triple
	return -1
}
