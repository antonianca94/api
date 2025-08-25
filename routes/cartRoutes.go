package routes

import (
	"api/controllers"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterCartRoutes(app *fiber.App, db *sql.DB) {
	cartGroup := app.Group("/cart")

	// Rotas para carrinho
	cartGroup.Post("/", controllers.CreateCart(db))
	cartGroup.Get("/", controllers.GetCarts(db))
	cartGroup.Get("/:id", controllers.GetCartByID(db))
	cartGroup.Get("/user/:user_id", controllers.GetCartByUserID(db))
	cartGroup.Get("/:id/items", controllers.GetCartWithItems(db))
	cartGroup.Patch("/:id", controllers.UpdateCart(db))
	cartGroup.Delete("/:id", controllers.DeleteCartByID(db))

	// Rotas para itens do carrinho
	cartGroup.Post("/cart-items", controllers.CreateCartItem(db))
	cartGroup.Get("/cart-items", controllers.GetCartItems(db))
	cartGroup.Get("/cart-items/:id", controllers.GetCartItemByID(db))
	cartGroup.Patch("/cart-items/:id", controllers.UpdateCartItem(db))
	cartGroup.Delete("/cart-items/:id", controllers.DeleteCartItemByID(db))

}
