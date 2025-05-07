package main

/*
	Aplicación web para gestionar un sistema de archivos EXT2,
	usando React para el frontend y Go para el backend.
*/

import (
	"Gestor/controllers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	// Habilitamos CORS para permitir peticiones desde el frontend
	app.Use(cors.New())

	// Grupo /api
	api := app.Group("/api")

	// Ruta para analizar comandos
	api.Post("/analizar", controllers.AnalizarComandos)

	// Ruta de estado
	api.Get("/status", controllers.GetStatus)

	// Iniciar servidor
	log.Println("Servidor escuchando en http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
