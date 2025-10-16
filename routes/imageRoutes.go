package routes

import (
	"api/controllers" // ajuste o caminho conforme sua estrutura de projeto
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterImageRoutes(app *fiber.App, db *sql.DB) {
	// Grupo de rotas para imagens
	imageGroup := app.Group("/images")

	// Rota para obter a imagem de um produto específico
	imageGroup.Get("/:product_id", controllers.GetImageOfProduct(db))
	imageGroup.Post("/", controllers.CreateImage(db))
	imageGroup.Get("/:product_id/type", controllers.GetImagesByProductAndType(db))

	imageGroup.Get("/name/:name", controllers.GetImageByName(db))
	imageGroup.Delete("/:id", controllers.DeleteImage(db))

}
