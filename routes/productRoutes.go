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
	productGroup.Get("/:sku", controllers.GetProductBySKU(db))
	productGroup.Get("/category/:category_name", controllers.GetProductsByCategoryName(db))
	productGroup.Get("/category/id/:category_id", controllers.GetProductsByCategoryID(db)) // Nova rota

	productGroup.Post("/", controllers.CreateProduct(db))
	productGroup.Get("/id/:id", controllers.GetProductByID(db))
	productGroup.Delete("/id/:id", controllers.DeleteProductByID(db))

	productGroup.Patch("/id/:id", controllers.UpdateProductByID(db))

}
