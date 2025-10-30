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
	vendorGroup.Get("/user/:users_id", controllers.GetVendorByUserID(db))

	app.Post("/checkout-multi-vendor/:user_id", controllers.FinalizeCheckoutMultiVendor(db))

    app.Get("/orders/user/:user_id/by-vendor", controllers.GetOrdersByVendor(db))

	app.Get("/orders/user/:user_id", controllers.GetUserOrders(db))

	app.Get("/orders/:id", controllers.GetOrderByID(db))

    app.Get("/orders/:id/details", controllers.GetOrderWithVendorInfo(db))
}
