package Acciones

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
Crear el archivo .mia que simulara el DISCO

	path -> es la ruta en donde se hara el archivo
	nombreDisco -> es el nombre del archivo

	returna un error, en dado caso falla la creacion
*/
func CrearDisco(path string, nombreDisco string) error { // retorna un error

	// Variable para guardar la ruta procesada
	var processedPath string

	/*
		// Detectar si estamos en macOS
		isMacOS := runtime.GOOS == "darwin"
		if isMacOS && strings.HasPrefix(path, "/home") {
			// En macOS, redirigir las rutas /home a una carpeta en el directorio del usuario
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf(" \n --> MKDISK (CrearDisco), ERROR: al obtener el directorio del usuario: %w", err)
			}

			// Reemplazar /home con el directorio simulado en la carpeta del usuario
			// Por ejemplo, /home/user/archivo.mia se convierte en /Users/gio/home/user/archivo.mia
			relPath := strings.TrimPrefix(path, "/home")
			processedPath = filepath.Join(homeDir, "home", relPath)
		} else {
			// En Linux o si la ruta no comienza con /home, usar la ruta original
			processedPath = path
		}
	*/

	processedPath = RutaCorrecta(path)

	fmt.Println("La path que entra al crear un disco: `", path, "`") // path original
	fmt.Println("La path procesada: `", processedPath, "`")          // path funcional en macOS

	// Verificar si el archivo ya existe ANTES de crear directorios
	if _, err := os.Stat(processedPath); err == nil {
		// El archivo ya existe, retornar un error
		errorMsg := fmt.Sprintf("\t ---> ERROR [ MK DISK ] (CrearDisco): el disco '%s' ya existe en la ruta '%s'", nombreDisco, path)
		fmt.Println(errorMsg)
		return err
	} else if !os.IsNotExist(err) {
		// Otro error al verificar el archivo
		return fmt.Errorf("\t ---> ERROR [ MK DISK ]: al verificar si el disco existe: %w", err)
	}

	// asegurar que exista la ruta (el directorio) creando la ruta
	// Obtener la carpeta donde se guardará el archivo para asegurarnos de que existe antes de crearlo.
	dir := filepath.Dir(processedPath) // extrae la ruta del directorio que contiene el archivo.

	// Crear el directorio si no existe y todas sus carpetas padre si no existen.
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("\t ---> ERROR [ MK DISK ](CrearDisco): al crear el disco: ", err)
		return err
	}

	// Verificar si el archivo ya existe
	newFile, err := os.Create(processedPath) // Crear el archivo si no existe
	if err != nil {
		fmt.Println("\t ---> ERROR [ MK DISK ](CrearDisco): al crear el disco: ", err)
		return err
	}
	defer newFile.Close() // Cierra el archivo

	//fmt.Println(" \n --> MKDISK, disco:", nombreDisco, "creado EXITOSAMENTE.")
	return nil

}

/*
abre un archivo en modo lectura/escritura (os.O_RDWR)
y retorna un puntero a os.File junto con un posible error.
*/
func OpenFile(name string) (*os.File, error) {
	name = RutaCorrecta(name)
	//fmt.Println("ruta desde open file: ", name)
	file, err := os.OpenFile(name, os.O_RDWR, 0664) // abrir el archivo en modo lectura y escritura
	if err != nil {
		fmt.Println("\t ---> ERROR [ MK DISK ](openFile): ", err)
		return nil, err
	}

	fmt.Println("[ DISK ] se abrio con exito el archivo")
	return file, nil // Si no hay error, retorna el archivo abierto
}

/*
Escribir datos binarios en un archivo en una posición específica.

	para esciribir 0 en el archivo binario.

	parametros:
		file *os.File: Un puntero a un archivo ya abierto.
		data interface{}: Datos genéricos que se escribirán en el archivo. disco de Ceros (0)
		position int64: La posición en el archivo donde se escribirá data. desde el inicio

	seek()
		El archivo está vacío al inicio.
		Seek(100, 0) mueve el puntero a la posición 100.
		Go detecta que el archivo es más corto que 100 bytes, por lo que lo expande.
		Los bytes desde el inicio (0) hasta la posición 99 se rellenan automáticamente con 0s
*/
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)                               //(posicion , desde donde(inicio) ) -> (5,0) significa a la posicion 5 desde el inicio del archivo
	err := binary.Write(file, binary.LittleEndian, data) //Escribir los datos en formato binario
	if err != nil {
		fmt.Println("Err WriteObject == ", err)
		return err
	}
	return nil
}

// Function to Read an object from a bin file
/*
lee datos de un archivo en una posición específica y los guarda en una variable (data).

	Mueve el "cursor" del archivo a una posición específica indicada por position.
	Esto se hace desde el inicio del archivo gracias al uso del 0 en file.Seek(position, 0).

	Luego, lee los datos desde esa posición hacia adelante hasta llenar la variable data.
	La cantidad de datos que se lee depende del tamaño de data.

*/
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0) // lee desde la posicion (posicion, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}

	return nil
}

func RutaCorrecta(path string) string {
	path = strings.Trim(path, `"`)

	if strings.HasPrefix(path, "/home") {
		// Cambia el prefijo /home por ./home (ruta relativa al proyecto)
		return "." + path
	}

	return path
}

/*

mkdisk -size=5 -unit=M -path="/home/mis discos/Disco3.mia"
fdisk -size=1 -type=L -unit=M -fit=BF-path="home/mis discos/Disco3.mia"-name="Particion3"
rmdisk                 -path="/home/mis discos/Disco3.mia"
*/

func RepGraphizMBR(path string, contenido string, nombre string) error {
	//asegurar la ruta
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Error al crear el reporte, path: ", err)
		return err
	}
	// Abrir o crear un archivo para escritura
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error al crear el archivo:", err)
		return err
	}
	defer file.Close()

	// Escribir en el archivo
	_, err = file.WriteString(contenido)
	if err != nil {
		fmt.Println("Error al escribir en el archivo:", err)
		return err
	}

	rep2 := dir + "/" + nombre + ".png"
	cmd := exec.Command("dot", "-Tpng", path, "-o", rep2)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error al generar el reporte PNG: %v", err)
	}

	return err
}
