package controllers

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para o produto
type Image struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	ProductID int    `json:"products_id,omitempty"`
}

// @Summary Obter imagem do produto
// @Description Obtém as imagens associadas a um produto específico
// @Tags Images
// @Param product_id path int true "ID do Produto"
// @Success 200 {array} Image
// @Failure 404 {object} map[string]string "Imagem não encontrada"
// @Failure 500 {object} map[string]string "Erro ao buscar imagem"
// @Router /images/{product_id} [get]
func GetImageOfProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productID := c.Params("product_id")

		// Consulta para buscar as imagens do produto
		query := `SELECT id, name, path, type FROM images WHERE products_id = ?`

		// Executa a consulta
		rows, err := db.Query(query, productID)
		if err != nil {
			log.Println("Erro ao buscar imagens:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar imagem"})
		}
		defer rows.Close()

		var images []Image
		for rows.Next() {
			var image Image
			if err := rows.Scan(&image.ID, &image.Name, &image.Path, &image.Type); err != nil {
				log.Println("Erro ao escanear imagem:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler imagem"})
			}
			images = append(images, image)
		}

		if len(images) == 0 {
			return c.Status(404).JSON(fiber.Map{"message": "Imagem não encontrada"})
		}

		return c.Status(200).JSON(images)
	}
}


// @Summary Criar imagem do produto
// @Description Cria uma nova imagem associada a um produto
// @Tags Images
// @Accept json
// @Produce json
// @Param image body Image true "Dados da imagem"
// @Success 201 {object} Image "Imagem criada com sucesso"
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar imagem"
// @Router /images [post]
func CreateImage(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var image Image

		// Parse do body da requisição
		if err := c.BodyParser(&image); err != nil {
			log.Println("Erro ao fazer parse do body:", err)
			return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
		}

		// Validação básica
		if image.Name == "" || image.Path == "" || image.Type == "" || image.ProductID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Todos os campos são obrigatórios"})
		}

		// Query para inserir a imagem
		query := `INSERT INTO images (name, path, type, products_id) VALUES (?, ?, ?, ?)`

		// Executa a inserção
		result, err := db.Exec(query, image.Name, image.Path, image.Type, image.ProductID)
		if err != nil {
			log.Println("Erro ao criar imagem:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar imagem"})
		}

		// Obtém o ID da imagem criada
		id, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID da imagem:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID da imagem"})
		}

		image.ID = int(id)

		return c.Status(201).JSON(image)
	}
}