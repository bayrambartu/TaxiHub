package main

import (
	"log"

	"github.com/bayrambartu/taxihub-driver-service/config"
	"github.com/bayrambartu/taxihub-driver-service/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	_ "github.com/bayrambartu/taxihub-driver-service/docs"
)

// @title TaxiHub Driver Service API
// @version 1.0
// @description TaxiHub Driver Service API
// @host localhost:5001
// @BasePath /api/v1
func main() {
	app := fiber.New()

	app.Get("/swagger/*", swagger.HandlerDefault)

	// DB connection
	client := config.ConnectDB()
	driverCollection := config.GetColleciton(client, "drivers")

	// handler
	driverHandle := &handler.DriverHandle{Collection: driverCollection}

	// routes
	api := app.Group("/api/v1")
	api.Post("/drivers", driverHandle.CreateDriver)
	api.Get("/drivers/", driverHandle.ListDriver)
	api.Get("/drivers/nearby", driverHandle.GetNearbyDrivers)
	api.Put("/drivers/:id", driverHandle.UpdateDriver)

	log.Fatal(app.Listen(":5001"))
}
