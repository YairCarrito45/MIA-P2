package controllers

import (
	"MIA-P2/Backend/models"
	"MIA-P2/Backend/services"
	"MIA-P2/Backend/Estructuras"
	"fmt"
	"strings"
	"path/filepath"

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

// HandleLogin procesa una solicitud de inicio de sesión (login)
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

// HandleDiskInfo devuelve información del disco montado
func HandleDiskInfo(c *fiber.Ctx) error {
	id := c.Params("id")

	for _, montada := range Estructuras.Montadas {
		if montada.Id == id {
			return c.JSON(fiber.Map{
				"name":               filepath.Base(montada.PathM),
				"path":               montada.PathM,
				"id":                 montada.Id,
				"fit":                "WF", // ajustar si es necesario
				"size":               0,    // puedes obtenerlo leyendo el MBR si quieres
				"mounted_partitions": getMountedPartitionIDs(montada.PathM),
			})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "No se encontró información para la partición " + id,
	})
}

// Función auxiliar: devuelve todos los IDs montados del mismo disco
func getMountedPartitionIDs(path string) []string {
	var ids []string
	for _, m := range Estructuras.Montadas {
		if m.PathM == path {
			ids = append(ids, m.Id)
		}
	}
	return ids
}
