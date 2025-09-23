package routes

import (
	"api/controllers" 
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterVendorRoutes(app *fiber.App, db *sql.DB) {
	vendorGroup := app.Group("/vendors")
	vendorGroup.Get("/", controllers.GetAllVendors(db))
	vendorGroup.Get("/:id", controllers.GetVendorByID(db))   
	vendorGroup.Post("/", controllers.CreateVendor(db))      
	vendorGroup.Delete("/:id", controllers.DeleteVendor(db)) 
	vendorGroup.Patch("/:id", controllers.UpdateVendor(db))     
}
