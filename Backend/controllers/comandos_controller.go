package controllers

import (
	"MIA-P2/Backend/Estructuras"
	"MIA-P2/Backend/models"
	"MIA-P2/Backend/services"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AnalizarComandos procesa los comandos recibidos desde el frontend
func AnalizarComandos(c *fiber.Ctx) error {
	var entrada models.EntradaComando

	if err := c.BodyParser(&entrada); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"output": "Error al procesar la solicitud: " + err.Error(),
		})
	}

	if strings.TrimSpace(entrada.Texto) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"output": "No se proporcionó ningún comando para analizar.",
		})
	}

	lineas := services.GetLineasComando(entrada.Texto)
	var salidas []string
	var errores []string
	todosExitosos := true

	for _, linea := range lineas {
		if linea == "" {
			continue
		}

		resultado := services.AnalizarComando(linea)

		fmt.Printf("Resultado del comando: %s\n", resultado.Comando)
		fmt.Printf("  Éxito: %v\n", resultado.Exito)

		if resultado.Salida != "" {
			salidas = append(salidas, resultado.Salida)
		}
		if resultado.Errores != "" {
			errores = append(errores, resultado.Errores)
			todosExitosos = false
		}
	}

	salidaFinal := strings.Join(salidas, "\n")
	if salidaFinal == "" {
		salidaFinal = "Comandos procesados pero no generaron salida."
	}
	if !todosExitosos && len(errores) > 0 {
		salidaFinal += "\n--- Errores ---\n" + strings.Join(errores, "\n")
	}

	return c.JSON(fiber.Map{
		"output": salidaFinal,
	})
}

// GetStatus devuelve el estado del servidor
func GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"output": "Servidor funcionando correctamente",
	})
}

// HandleLogin procesa una solicitud de inicio de sesión
func HandleLogin(c *fiber.Ctx) error {
	var loginData struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		PartitionID string `json:"partition_id"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos inválidos: " + err.Error(),
		})
	}

	comando := fmt.Sprintf("login -user=%s -pass=%s -id=%s", loginData.Username, loginData.Password, loginData.PartitionID)
	resultado := services.AnalizarComando(comando)

	if !resultado.Exito {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": resultado.Errores,
		})
	}

	return c.JSON(fiber.Map{
		"output": resultado.Salida,
	})
}

// HandleDiskInfo devuelve información de un disco montado
func HandleDiskInfo(c *fiber.Ctx) error {
	id := c.Params("id")

	for _, montada := range Estructuras.Montadas {
		if montada.Id == id {
			// Obtener tamaño del archivo del disco
			info, err := os.Stat(montada.PathM)
			size := int64(0)
			if err == nil {
				size = info.Size()
			}
			return c.JSON(fiber.Map{
				"name":               filepath.Base(montada.PathM),
				"path":               montada.PathM,
				"id":                 montada.Id,
				"fit":                "WF", // modificar si se tiene info real del fit
				"size":               size,
				"mounted_partitions": getMountedPartitionIDs(montada.PathM),
			})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "No se encontró información para la partición " + id,
	})
}

// GetDisks devuelve una lista de discos con o sin particiones montadas
func GetDisks(c *fiber.Ctx) error {
	var discosMap = make(map[string]map[string]interface{})

	for _, montada := range Estructuras.Montadas {
		info, err := os.Stat(montada.PathM)
		size := int64(0)
		if err == nil {
			size = info.Size()
		}

		if _, ok := discosMap[montada.PathM]; !ok {
			discosMap[montada.PathM] = map[string]interface{}{
				"name":               filepath.Base(montada.PathM),
				"path":               montada.PathM,
				"size":               size, // ✅ aquí lo asignas correctamente
				"fit":                "WF", // ajusta si usas otro
				"mounted_partitions": []string{},
			}
		}
		discosMap[montada.PathM]["mounted_partitions"] = append(
			discosMap[montada.PathM]["mounted_partitions"].([]string),
			montada.Id,
		)
	}

	var discos []map[string]interface{}
	for _, info := range discosMap {
		discos = append(discos, info)
	}

	return c.JSON(discos)
}

// Función auxiliar para IDs de particiones montadas por disco
func getMountedPartitionIDs(path string) []string {
	var ids []string
	for _, m := range Estructuras.Montadas {
		if m.PathM == path {
			ids = append(ids, m.Id)
		}
	}
	return ids
}


// HandlePartitions devuelve las particiones activas para un disco
func HandlePartitions(c *fiber.Ctx) error {
	path := c.Query("path")
	if strings.TrimSpace(path) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "El parámetro 'path' es obligatorio",
		})
	}

	// Abrir el archivo del disco
	archivo, err := os.Open(path)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo abrir el archivo del disco: " + err.Error(),
		})
	}
	defer archivo.Close()

	// Leer el MBR desde el inicio
	var mbr Estructuras.MBR
	if err := binary.Read(archivo, binary.LittleEndian, &mbr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo leer el MBR: " + err.Error(),
		})
	}

	var particiones []fiber.Map
	for _, montada := range Estructuras.Montadas {
		if montada.PathM == path {
			// Buscar la partición montada en el MBR para obtener su tamaño
			for _, part := range mbr.Mbr_partitions {
				if Estructuras.GetId(string(part.Part_id[:])) == montada.Id {
					particiones = append(particiones, fiber.Map{
						"id":     montada.Id,
						"name":   Estructuras.GetName(string(part.Part_name[:])),
						"fit":    string(part.Part_fit[:]),
						"size":   part.Part_size,
						"status": "montada",
					})
					break
				}
			}
		}
	}

	return c.JSON(particiones)
}