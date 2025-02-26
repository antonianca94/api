package routes

import (
	"api/controllers" // ajuste o caminho conforme sua estrutura de projeto
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterVendorRoutes(app *fiber.App, db *sql.DB) {
	vendorGroup := app.Group("/vendors")
	vendorGroup.Get("/", controllers.GetAllVendors(db))
	vendorGroup.Get("/:id", controllers.GetVendorByID(db))   // Nova rota para obter vendor por ID
	vendorGroup.Post("/", controllers.CreateVendor(db))      // Nova rota para criar vendor
	vendorGroup.Delete("/:id", controllers.DeleteVendor(db)) // Nova rota para deletar vendor

}
