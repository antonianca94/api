package routes

import (
	"api/controllers"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// Função para registrar as rotas de usuários
func RegisterRoleRoutes(app *fiber.App, db *sql.DB) {
	// Grupo de rotas para /users
	roleGroup := app.Group("/roles")
	roleGroup.Get("/", controllers.GetRoles(db))             // Usando o controlador para a rota de listagem de usuários
	roleGroup.Delete("/:id", controllers.DeleteRoleByID(db)) // Rota para deletar uma role por ID
	roleGroup.Get("/:id", controllers.GetRoleByID(db))
	roleGroup.Patch("/:id", controllers.UpdateRole(db))
	roleGroup.Post("/", controllers.CreateRole(db))

}
