package main

import (
	"MIA-P2/Backend/controllers"
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
	api.Post("/analizar", controllers.AnalizarComandos)
	api.Get("/status", controllers.GetStatus)

	// ✅ Nuevas rutas directas para login y disco
	app.Post("/login", controllers.HandleLogin)
	app.Get("/diskinfo/:id", controllers.HandleDiskInfo)

	api.Get("/disks", controllers.GetDisks)

	api.Get("/partitions", controllers.HandlePartitions)

	// Iniciar servidor
	log.Println("Servidor escuchando en http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
