package Rep

import (
	"Gestor/Acciones"
	"Gestor/Estructuras"
	"Gestor/utils"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	strExt2 "Gestor/Estructuras/SystemFileExt2"

	"strings"
)

func reporte_ls(path string, id string, ruta_dir string, logger *utils.Logger) {
	var pathDisco string
	existe := false

	// Busca en struct de particiones montadas el id ingresado
	for _, montado := range Estructuras.Montadas {
		if montado.Id == id {
			pathDisco = montado.PathM
			existe = true
			break
		}
	}

	if existe {
		// Obtener nombre del reporte y del disco
		tmp := strings.Split(path, "/")
		nombreReporte := strings.Split(tmp[len(tmp)-1], ".")[0]

		logger.LogInfo("%s", nombreReporte)

		// Disco a reportar
		tmp = strings.Split(pathDisco, "/")
		disco := strings.Split(tmp[len(tmp)-1], ".")[0]

		// Abrir el archivo del disco
		file, err := Acciones.OpenFile(pathDisco)
		if err != nil {
			logger.LogError("REP ERROR: No se pudo abrir el disco %s", pathDisco)
			return
		}
		defer file.Close()

		// Leer el MBR para encontrar la partición con el ID especificado
		var mbr Estructuras.MBR
		if err := Acciones.ReadObject(file, &mbr, 0); err != nil {
			logger.LogError("REP ERROR: No se pudo leer el MBR del disco %s", pathDisco)
			return
		}

		// Buscar la partición con el ID especificado
		var particion Estructuras.Partition
		encontrada := false

		for i := 0; i < 4; i++ {
			identificador := Estructuras.GetId(string(mbr.Mbr_partitions[i].Part_id[:]))
			if identificador == id {
				particion = mbr.Mbr_partitions[i]
				encontrada = true
				break
			}
		}

		if !encontrada {
			logger.LogError("REP ERROR: No se encontró la partición con ID %s", id)
			return
		}

		// Leer el superbloque
		var superBloque strExt2.Superblock
		if err := Acciones.ReadObject(file, &superBloque, int64(particion.Part_start)); err != nil {
			logger.LogError("REP ERROR: No se pudo leer el superbloque de la partición con ID %s", id)
			return
		}

		// Asegurar que la ruta comienza con /
		if !strings.HasPrefix(ruta_dir, "/") {
			ruta_dir = "/" + ruta_dir
		}

		// Buscar el inodo de la carpeta que queremos listar
		idInodoCarpeta := int32(0) // Por defecto, la raíz

		// Si no es la raíz, buscar el inodo correspondiente
		if ruta_dir != "/" {
			idInodoCarpeta = strExt2.BuscarInodo(0, ruta_dir, superBloque, file)

			if idInodoCarpeta == -1 {
				logger.LogError("REP ERROR: No se encontró la carpeta %s", ruta_dir)
				return
			}
		}

		// Leer el inodo de la carpeta
		var inodoCarpeta strExt2.Inode
		Acciones.ReadObject(file, &inodoCarpeta, int64(superBloque.S_inode_start+(idInodoCarpeta*int32(binary.Size(strExt2.Inode{})))))

		// Verificar que sea una carpeta
		if string(inodoCarpeta.I_type[:]) != "0" {
			logger.LogError("REP ERROR: La ruta %s no corresponde a una carpeta", ruta_dir)
			return
		}

		// Crear contenido del reporte
		var contenidoReporte strings.Builder

		// Encabezado
		contenidoReporte.WriteString("REPORTE LS\n")
		contenidoReporte.WriteString("=========\n\n")
		contenidoReporte.WriteString("Disco: " + disco + "\n")
		contenidoReporte.WriteString("Partición: " + id + "\n")
		contenidoReporte.WriteString("Directorio: " + ruta_dir + "\n\n")

		// Tabla de información
		contenidoReporte.WriteString(fmt.Sprintf("%-12s %-10s %-15s %-10s %-12s %-8s %-10s %-s\n",
			"Permisos", "Owner", "Grupo", "Size (Bytes)", "Fecha", "Hora", "Tipo", "Nombre"))
		contenidoReporte.WriteString(fmt.Sprintf("%-12s %-10s %-15s %-10s %-12s %-8s %-10s %-s\n",
			"----------", "----------", "---------------", "----------", "------------", "--------", "----------", "-------"))

		// Lista de archivos y carpetas
		// Recorrer los bloques de la carpeta y mostrar su contenido
		var archivosListados int = 0

		for i := 0; i < 12; i++ {
			idBloque := inodoCarpeta.I_block[i]
			if idBloque != -1 {
				// Leer el bloque de carpeta
				var folderBlock strExt2.Folderblock
				Acciones.ReadObject(file, &folderBlock, int64(superBloque.S_block_start+(idBloque*int32(binary.Size(strExt2.Folderblock{})))))

				// Procesar cada entrada del bloque
				for j := 0; j < 4; j++ {
					idInodoEntrada := folderBlock.B_content[j].B_inodo

					// Si es una entrada válida y no es "." o ".."
					if idInodoEntrada != -1 {
						nombre := strExt2.GetB_name(string(folderBlock.B_content[j].B_name[:]))

						// Saltarse las entradas . y .. excepto en la raíz donde las mostramos
						if j < 2 && idInodoCarpeta != 0 {
							continue
						}

						// Leer el inodo de esta entrada
						var inodoEntrada strExt2.Inode
						Acciones.ReadObject(file, &inodoEntrada, int64(superBloque.S_inode_start+(idInodoEntrada*int32(binary.Size(strExt2.Inode{})))))

						// Determinar el tipo
						tipoEntrada := "Archivo"
						if string(inodoEntrada.I_type[:]) == "0" {
							tipoEntrada = "Carpeta"
						}

						// Obtener información del usuario y grupo propietario
						// En este caso usamos los IDs directamente, pero podrías tener una función
						// para mapear IDs a nombres de usuario/grupo
						owner := fmt.Sprintf("User%d", inodoEntrada.I_uid)
						grupo := fmt.Sprintf("Grupo%d", inodoEntrada.I_gid)

						// Formatear permisos (podrías convertirlos a formato rwx si lo prefieres)
						permisos := string(inodoEntrada.I_perm[:])

						// Extraer fecha y hora de modificación
						fechaModificacion := strings.Trim(string(inodoEntrada.I_mtime[:]), "\x00")
						partesFecha := strings.Split(fechaModificacion, " ")
						fecha := ""
						hora := ""
						if len(partesFecha) >= 2 {
							fecha = partesFecha[0]
							hora = partesFecha[1]
						} else {
							fecha = fechaModificacion
						}

						// Añadir esta entrada al reporte
						contenidoReporte.WriteString(fmt.Sprintf("%-12s %-10s %-15s %-10d %-12s %-8s %-10s %-s\n",
							permisos, owner, grupo, inodoEntrada.I_size, fecha, hora, tipoEntrada, nombre))

						archivosListados++
					}
				}
			}
		}

		// Información de resumen
		contenidoReporte.WriteString(fmt.Sprintf("\nTotal de archivos/carpetas: %d\n", archivosListados))

		// Crear el directorio si no existe
		carpeta := filepath.Dir(path)
		if err := os.MkdirAll("."+carpeta, os.ModePerm); err != nil {
			logger.LogError("REP ERROR: No se pudo crear el directorio para el reporte: %v", err)
			return
		}

		// Escribir el archivo de reporte
		rutaReporte := "." + path
		err = os.WriteFile(rutaReporte, []byte(contenidoReporte.String()), 0644)
		if err != nil {
			logger.LogError("REP ERROR: No se pudo crear el archivo de reporte: %v", err)
			return
		}

		logger.LogInfo("Reporte LS del directorio %s en el disco %s creado exitosamente en %s",
			ruta_dir, disco, rutaReporte)
	} else {
		logger.LogError("REP ERROR: La partición con ID %s no está montada", id)
	}
}
