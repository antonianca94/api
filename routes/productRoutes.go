package routes

import (
	"api/controllers" // ajuste o caminho conforme sua estrutura de projeto
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterProductRoutes(app *fiber.App, db *sql.DB) {

	productGroup := app.Group("/products")
	productGroup.Get("/home", controllers.GetAllProductsHome(db))
	productGroup.Get("/", controllers.GetAllProducts(db))
	productGroup.Get("/user/:user_id", controllers.GetAllProductsByUserID(db))
	productGroup.Get("/sku/:sku", controllers.GetProductBySKU(db))
}
