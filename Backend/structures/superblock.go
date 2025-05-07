package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type SuperBlock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_inodes_count int32
	S_free_blocks_count int32
	S_mtime             float32
	S_umtime            float32
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
	// Total: 68 bytes
}

// Serialize escribe la estructura SuperBlock en un archivo binario en la posición especificada
func (sb *SuperBlock) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura SuperBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura SuperBlock desde un archivo binario en la posición especificada
func (sb *SuperBlock) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura SuperBlock
	sbSize := binary.Size(sb)
	if sbSize <= 0 {
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura SuperBlock
	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura SuperBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, sb)
	if err != nil {
		return err
	}

	return nil
}

// PrintSuperBlock imprime los valores de la estructura SuperBlock
func (sb *SuperBlock) Print() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.S_umtime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)
}

// Imprimir inodos
func (sb *SuperBlock) PrintInodes(path string) error {
	// Imprimir inodos
	fmt.Println("\nInodos\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Imprimir el inodo
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

// Impriir bloques
func (sb *SuperBlock) PrintBlocks(path string) error {
	// Imprimir bloques
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &Inode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return err
		}
		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.I_block {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			// Si el inodo es de tipo carpeta
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

				// Si el inodo es de tipo archivo
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *SuperBlock) CreateFolder(path string, parentsDir []string, destDir string, allowParents bool) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFolderInInode(path, 0, parentsDir, destDir)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFolderInInode(path, i, parentsDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sb *SuperBlock) GetUsersBlock(path string) (*FileBlock, error) {
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start + 1*sb.S_inode_size)) // Inodo 1 → users.txt
	if err != nil {
		return nil, err
	}

	if inode.I_type[0] != '1' {
		return nil, fmt.Errorf("el inodo no corresponde a un archivo")
	}

	blockIndex := inode.I_block[0]
	if blockIndex == -1 {
		return nil, fmt.Errorf("el archivo users.txt no tiene bloques asignados")
	}

	block := &FileBlock{}
	err = block.Deserialize(path, int64(sb.S_block_start + blockIndex*sb.S_block_size))
	if err != nil {
		return nil, err
	}

	return block, nil
}



// Verifica si todas las carpetas en dirNames existen en el sistema de archivos
func (sb *SuperBlock) DirectoriesExist(diskPath string, dirNames []string) bool {
	currentInode := int32(0) // Inodo raíz

	for _, dirName := range dirNames {
		found := false

		// Leer el inodo actual
		inode := &Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(currentInode*sb.S_inode_size)))
		if err != nil {
			return false
		}

		// Recorrer bloques del inodo (esperamos FolderBlocks)
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}

			block := &FolderBlock{}
			err := block.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				continue
			}

			for _, content := range block.B_content {
				name := strings.TrimRight(string(bytes.Trim(content.B_name[:], "\x00")), " ")
				if name == dirName {
					currentInode = content.B_inodo
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return false // Una carpeta no existe
		}
	}

	return true
}



// Encuentra y devuelve el inodo correspondiente a una ruta absoluta.
// Si no existe, retorna nil.
func (sb *SuperBlock) FindInodeByPath(diskPath string, fullPath string) *Inode {
	if fullPath == "/" {
		inode := &Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start))
		if err != nil {
			return nil
		}
		return inode
	}

	pathParts := strings.Split(strings.Trim(fullPath, "/"), "/")
	currentInode := int32(0) // Inicia en raíz

	for _, part := range pathParts {
		found := false

		inode := &Inode{}
		err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(currentInode*sb.S_inode_size)))
		if err != nil {
			return nil
		}

		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				continue
			}

			folderBlock := &FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				continue
			}

			for _, content := range folderBlock.B_content {
				name := strings.TrimRight(string(bytes.Trim(content.B_name[:], "\x00")), " ")
				if name == part {
					currentInode = content.B_inodo
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return nil // No se encontró una parte del path
		}
	}

	// Cargar el inodo final
	finalInode := &Inode{}
	err := finalInode.Deserialize(diskPath, int64(sb.S_inode_start+(currentInode*sb.S_inode_size)))
	if err != nil {
		return nil
	}

	return finalInode
}


// ReadDirectoryTree genera una estructura en forma de árbol del sistema de archivos
func (sb *SuperBlock) ReadDirectoryTree(path string) (map[string]interface{}, error) {
	rootInode := &Inode{}
	err := rootInode.Deserialize(path, int64(sb.S_inode_start))
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el inodo raíz: %v", err)
	}

	return sb.recursiveReadInode(path, rootInode, "/")
}

func (sb *SuperBlock) recursiveReadInode(path string, inode *Inode, name string) (map[string]interface{}, error) {
	node := map[string]interface{}{
		"name":     name,
		"type":     "folder",
		"children": []interface{}{},
	}

	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		block := &FolderBlock{}
		err := block.Deserialize(path, int64(sb.S_block_start+blockIndex*sb.S_block_size))
		if err != nil {
			continue
		}

		for _, content := range block.B_content {
			childName := strings.TrimRight(string(bytes.Trim(content.B_name[:], "\x00")), " ")
			if childName == "" || childName == "." || childName == ".." {
				continue
			}

			childInode := &Inode{}
			offset := int64(sb.S_inode_start) + int64(content.B_inodo)*int64(sb.S_inode_size)
			err := childInode.Deserialize(path, offset)
			if err != nil {
				continue
			}

			if childInode.I_type[0] == '1' {
				child := map[string]interface{}{
					"name": childName,
					"type": "file",
				}
				node["children"] = append(node["children"].([]interface{}), child)
			} else if childInode.I_type[0] == '0' {
				childNode, err := sb.recursiveReadInode(path, childInode, childName)
				if err == nil {
					node["children"] = append(node["children"].([]interface{}), childNode)
				}
			}
		}
	}

	return node, nil
}
