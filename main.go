package main

import (
	"api/routes"
	"database/sql"
	"log"

	_ "api/docs" // Certifique-se de importar o pacote docs gerado pelo swag

	_ "github.com/go-sql-driver/mysql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	swagger "github.com/gofiber/swagger" // Módulo do Fiber para Swagger
)

var db *sql.DB

// @title API do AgroFood
// @version 1.0
// @description API para gerenciar o sistema Agrofood.
// @host localhost:3002
// @BasePath /
func main() {
	var err error
	// Conexão com o banco de dados
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/agrofood")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verificar a conexão
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Inicializa o Fiber
	app := fiber.New()

	app.Use(cors.New())

	// Middleware de logger
	app.Use(logger.New())

	// Registrar as rotas de usuários
	routes.RegisterUserRoutes(app, db)
	routes.RegisterRoleRoutes(app, db)
	routes.RegisterProductRoutes(app, db)

	// Adicionar rota para a documentação Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // serve swagger

	// Iniciar o servidor
	log.Fatal(app.Listen(":3002"))
}
