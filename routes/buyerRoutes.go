package routes

import (
	"api/controllers" 
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func RegisterBuyerRoutes(app *fiber.App, db *sql.DB) {
	buyerGroup := app.Group("/buyers")
	buyerGroup.Get("/", controllers.GetAllBuyers(db))
	buyerGroup.Get("/:id", controllers.GetBuyerByID(db))   
	buyerGroup.Post("/", controllers.CreateBuyer(db))      
	buyerGroup.Delete("/:id", controllers.DeleteBuyer(db)) 
	buyerGroup.Patch("/:id", controllers.UpdateBuyer(db))     
	buyerGroup.Get("/user/:users_id", controllers.GetBuyerByUserID(db))

}
