package routes

import (
	"api/controllers"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterCategoryRoutes(app *fiber.App, db *sql.DB) {
	categoryGroup := app.Group("/categories")

	categoryGroup.Post("/", controllers.CreateCategory(db))
	categoryGroup.Get("/", controllers.GetCategories(db))
	categoryGroup.Get("/:id", controllers.GetCategoryByID(db))
	categoryGroup.Patch("/:id", controllers.UpdateCategory(db))
	categoryGroup.Delete("/:id", controllers.DeleteCategoryByID(db))

}
