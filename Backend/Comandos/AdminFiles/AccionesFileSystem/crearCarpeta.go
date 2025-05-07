package AccionesFileSystem

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	strExt2 "Gestor/Estructuras/SystemFileExt2"
	"Gestor/utils"
	"encoding/binary"
	"os"
	"time"
)

func CrearCarpeta(idInode int32, carpeta string, initSuperBloque int64, disco *os.File, logger *utils.Logger) int32 {
	var superBloque strExt2.Superblock
	Acciones.ReadObject(disco, &superBloque, initSuperBloque)

	var inodo strExt2.Inode
	Acciones.ReadObject(disco, &inodo, int64(superBloque.S_inode_start+(idInode*int32(binary.Size(strExt2.Inode{})))))

	//fmt.Println("Creando carpeta ", carpeta)
	logger.LogInfo("Creando carpeta %s", carpeta)
	//Recorrer los bloques directos del inodo para ver si hay espacio libre
	for i := 0; i < 12; i++ {
		idBloque := inodo.I_block[i]
		if idBloque != -1 {
			//Existe un folderblock con idBloque que se debe revisar si tiene espacio para la nueva carpeta
			var folderBlock strExt2.Folderblock
			Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

			//Recorrer el bloque para ver si hay espacio
			for j := 2; j < 4; j++ {
				apuntador := folderBlock.B_content[j].B_inodo
				//Hay espacio en el bloque
				if apuntador == -1 {
					//modifico el bloque actual
					copy(folderBlock.B_content[j].B_name[:], carpeta)
					ino := superBloque.S_first_ino //primer inodo libre
					folderBlock.B_content[j].B_inodo = ino
					//ACTUALIZAR EL FOLDERBLOCK ACTUAL (idBloque) EN EL ARCHIVO
					Acciones.WriteObject(disco, folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

					//creo el nuevo inodo /ruta
					var newInodo strExt2.Inode
					newInodo.I_uid = Estructuras.UsuarioActual.IdUsr
					newInodo.I_gid = Estructuras.UsuarioActual.IdGrp
					newInodo.I_size = 0 //es carpeta
					//Agrego las fechas
					ahora := time.Now()
					date := ahora.Format("02/01/2006 15:04")
					copy(newInodo.I_atime[:], date)
					copy(newInodo.I_ctime[:], date)
					copy(newInodo.I_mtime[:], date)
					copy(newInodo.I_type[:], "0") //es carpeta
					copy(newInodo.I_mtime[:], "664")

					//apuntadores iniciales
					for i := int32(0); i < 15; i++ {
						newInodo.I_block[i] = -1
					}
					//El apuntador a su primer bloque (el primero disponible)
					block := superBloque.S_first_blo
					newInodo.I_block[0] = block
					//escribo el nuevo inodo (ino)
					Acciones.WriteObject(disco, newInodo, int64(superBloque.S_inode_start+(ino*int32(binary.Size(strExt2.Inode{})))))

					//crear el nuevo bloque
					var newFolderBlock strExt2.Folderblock
					newFolderBlock.B_content[0].B_inodo = ino //idInodo actual
					copy(newFolderBlock.B_content[0].B_name[:], ".")
					newFolderBlock.B_content[1].B_inodo = folderBlock.B_content[0].B_inodo //el padre es el bloque anterior
					copy(newFolderBlock.B_content[1].B_name[:], "..")
					newFolderBlock.B_content[2].B_inodo = -1
					newFolderBlock.B_content[3].B_inodo = -1
					//escribo el nuevo bloque (block)
					Acciones.WriteObject(disco, newFolderBlock, int64(superBloque.S_block_start+(block*int32(binary.Size(strExt2.Folderblock{})))))

					//modifico el superbloque
					superBloque.S_free_inodes_count -= 1
					superBloque.S_free_blocks_count -= 1
					superBloque.S_first_blo += 1
					superBloque.S_first_ino += 1
					//Escribir en el archivo los cambios del superBloque
					Acciones.WriteObject(disco, superBloque, initSuperBloque)

					//escribir el bitmap de bloques (se uso un bloque).
					Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+block))

					//escribir el bitmap de inodos (se uso un inodo).
					Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_inode_start+ino))
					//retorna el inodo creado (por si va a crear otra carpeta en ese inodo)
					return ino
				}
			} //fin de for de buscar espacio en el bloque actual (existente)
		} else {
			//No hay bloques con espacio disponible (existe al menos el primer bloque pero esta lleno)
			//modificar el inodo actual (por el nuevo apuntador)
			block := superBloque.S_first_blo //primer bloque libre
			inodo.I_block[i] = block
			//Escribir los cambios del inodo inicial
			Acciones.WriteObject(disco, &inodo, int64(superBloque.S_inode_start+(idInode*int32(binary.Size(strExt2.Inode{})))))

			//cargo el primer bloque del inodo actual para tomar los datos de actual y padre (son los mismos para el nuevo)
			var folderBlock strExt2.Folderblock
			bloque := inodo.I_block[0] //cargo el primer folderblock para obtener los datos del actual y su padre
			Acciones.ReadObject(disco, &folderBlock, int64(superBloque.S_block_start+(bloque*int32(binary.Size(strExt2.Folderblock{})))))

			//creo el bloque que va a apuntar a la nueva carpeta
			var newFolderBlock1 strExt2.Folderblock
			newFolderBlock1.B_content[0].B_inodo = folderBlock.B_content[0].B_inodo //actual
			copy(newFolderBlock1.B_content[0].B_name[:], ".")
			newFolderBlock1.B_content[1].B_inodo = folderBlock.B_content[1].B_inodo //padre
			copy(newFolderBlock1.B_content[1].B_name[:], "..")
			ino := superBloque.S_first_ino                        //primer inodo libre
			newFolderBlock1.B_content[2].B_inodo = ino            //apuntador al inodo nuevo
			copy(newFolderBlock1.B_content[2].B_name[:], carpeta) //nombre del inodo nuevo
			newFolderBlock1.B_content[3].B_inodo = -1
			//escribo el nuevo bloque (block)
			Acciones.WriteObject(disco, newFolderBlock1, int64(superBloque.S_block_start+(block*int32(binary.Size(strExt2.Folderblock{})))))

			//creo el nuevo inodo /ruta
			var newInodo strExt2.Inode
			newInodo.I_uid = Estructuras.UsuarioActual.IdUsr
			newInodo.I_gid = Estructuras.UsuarioActual.IdGrp
			newInodo.I_size = 0 //es carpeta
			//Agrego las fechas
			ahora := time.Now()
			date := ahora.Format("02/01/2006 15:04")
			copy(newInodo.I_atime[:], date)
			copy(newInodo.I_ctime[:], date)
			copy(newInodo.I_mtime[:], date)
			copy(newInodo.I_type[:], "0") //es carpeta
			copy(newInodo.I_mtime[:], "664")

			//apuntadores iniciales
			for i := int32(0); i < 15; i++ {
				newInodo.I_block[i] = -1
			}
			//El apuntador a su primer bloque (el primero disponible)
			block2 := superBloque.S_first_blo + 1
			newInodo.I_block[0] = block2
			//escribo el nuevo inodo (ino) creado en newFolderBlock1
			Acciones.WriteObject(disco, newInodo, int64(superBloque.S_inode_start+(ino*int32(binary.Size(strExt2.Inode{})))))

			//crear nuevo bloque del inodo
			var newFolderBlock2 strExt2.Folderblock
			newFolderBlock2.B_content[0].B_inodo = ino //idInodo actual
			copy(newFolderBlock2.B_content[0].B_name[:], ".")
			newFolderBlock2.B_content[1].B_inodo = newFolderBlock1.B_content[0].B_inodo //el padre es el bloque anterior
			copy(newFolderBlock2.B_content[1].B_name[:], "..")
			newFolderBlock2.B_content[2].B_inodo = -1
			newFolderBlock2.B_content[3].B_inodo = -1
			//escribo el nuevo bloque
			Acciones.WriteObject(disco, newFolderBlock2, int64(superBloque.S_block_start+(block2*int32(binary.Size(strExt2.Folderblock{})))))

			//modifico el superbloque
			superBloque.S_free_inodes_count -= 1
			superBloque.S_free_blocks_count -= 2
			superBloque.S_first_blo += 2
			superBloque.S_first_ino += 1
			Acciones.WriteObject(disco, superBloque, initSuperBloque)

			//escribir el bitmap de bloques (se uso dos bloques: block y block2).
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+block))
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_block_start+block2))

			//escribir el bitmap de inodos (se uso un inodo: ino).
			Acciones.WriteObject(disco, byte(1), int64(superBloque.S_bm_inode_start+ino))
			return ino
		}
	}
	return 0
}
