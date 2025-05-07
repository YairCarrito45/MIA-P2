package controllers

import (
	"Gestor/models"
	"Gestor/services"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AnalizarComandos procesa los comandos recibidos desde el frontend
func AnalizarComandos(c *fiber.Ctx) error {
	var entrada models.EntradaComando

	// Parsear el cuerpo JSON
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

	// Devuelve solo el campo "output"
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
