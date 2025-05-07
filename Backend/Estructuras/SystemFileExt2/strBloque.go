package SystemFileExt2

import "strings"

// BLOQUE DE CARPETAS
// tamaño en bytes 4(estructuras content)*12(B_name)*4(B_inodo) = 64
type Folderblock struct {
	B_content [4]Content //contenido de la carpeta
}

type Content struct {
	B_name  [12]byte //nombre de carpeta/archivo
	B_inodo int32    //apuntador a un inodo asociado al archivo/carpeta
}

// Metodo que anula bytes nulos para B_name
func GetB_name(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)

	if posicionNulo != -1 {
		if posicionNulo != 0 {
			//tiene bytes nulos
			nombre = nombre[:posicionNulo]
		} else {
			//el  nombre esta vacio
			nombre = "-"
		}

	}
	return nombre //-1 el nombre no tiene bytes nulos
}

// BLOQUE DE ARCHIVOS
type Fileblock struct {
	B_content [64]byte //contenido del archivo
}

// Metodo que anula bytes nulos para B_content PARA EL REPORTE
func GetB_content(nombre string) string {
	// Reemplazar todos los saltos de línea con un guion (-)
	nombre = strings.ReplaceAll(nombre, "\n", "<br/>")
	posicionNulo := strings.IndexByte(nombre, 0)

	if posicionNulo != -1 {
		if posicionNulo != 0 {
			//tiene bytes nulos
			nombre = nombre[:posicionNulo]
		} else {
			//el  nombre esta vacio
			nombre = "-"
		}

	}
	//regreso los saltos de linea ya sin bytes nulos
	//nombre = strings.ReplaceAll(nombre, "-", "\n")
	return nombre //-1 el nombre no tiene bytes nulos
}

// BLOQUE DE APUNTADORES INDIRECTOS
type Pointerblock struct {
	B_pointers [16]int32 //apuntadores a bloques (archivo/carpeta)
}

// Journaling EXT3
type Journaling struct {
	Size      int32
	Ultimo    int32
	Contenido [50]Content_J
}

type Content_J struct {
	Operation [10]byte
	Path      [100]byte
	Content   [100]byte
	Date      [16]byte
}

// funciones de journaling
func GetOperation(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	nombre = nombre[:posicionNulo] //guarda la cadena hasta donde encontro un byte nulo
	return nombre
}

func GetPath(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	nombre = nombre[:posicionNulo] //guarda la cadena hasta donde encontro un byte nulo
	return nombre
}

func GetContent(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	nombre = nombre[:posicionNulo] //guarda la cadena hasta donde encontro un byte nulo
	return nombre
}

// para leer byte por byte los bitmaps (reportes)
type Bite struct {
	Val [1]byte
}
