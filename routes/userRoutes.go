package routes

import (
	"api/controllers"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App, db *sql.DB) {
	userGroup := app.Group("/users")
	userGroup.Post("/", controllers.CreateUser(db))
	userGroup.Get("/", controllers.GetUsers(db))
	userGroup.Get("/:id", controllers.GetUserByID(db))
	userGroup.Get("/details/:id", controllers.GetUserDetailsByID(db))
	userGroup.Delete("/:id", controllers.DeleteUserByID(db))
	userGroup.Patch("/:id", controllers.PatchUserByID(db))
}
