package controllers

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para categoria
type Category struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	IDCategoriesProducts *int   `json:"id_categories_products,omitempty"`
}

// CreateCategory cria uma nova Categoria
// @Summary Cria uma nova Categoria
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param category body Category true "Dados da nova Categoria"
// @Success 201 {object} Category "Categoria criada com sucesso"
// @Failure 400 {object} map[string]string "Dados de entrada inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar Categoria"
// @Router /categories [post]
func CreateCategory(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var newCategory Category

		// Lê os dados da nova Categoria do corpo da requisição
		if err := c.BodyParser(&newCategory); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Insere a nova Categoria no banco de dados
		query := "INSERT INTO categories_products (name, description, id_categories_products) VALUES (?, ?, ?)"
		result, err := db.Exec(query, newCategory.Name, newCategory.Description, newCategory.IDCategoriesProducts)
		if err != nil {
			log.Println("Erro ao criar categoria:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar categoria"})
		}

		// Obtém o ID da nova Categoria inserida
		categoryID, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID da nova categoria:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID da nova categoria"})
		}

		// Atribui o ID à nova Categoria
		newCategory.ID = int(categoryID)

		return c.Status(201).JSON(newCategory)
	}
}

// UpdateCategory atualiza parcialmente uma Categoria com base no ID
// @Summary Atualiza parcialmente uma Categoria pelo ID
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path int true "ID da Categoria"
// @Param category body Category true "Dados da Categoria para atualização parcial"
// @Success 200 {object} Category "Categoria atualizada com sucesso"
// @Failure 400 {object} map[string]string "ID inválido ou dados de entrada inválidos"
// @Failure 404 {object} map[string]string "Categoria não encontrada"
// @Failure 500 {object} map[string]string "Erro ao atualizar Categoria"
// @Router /categories/{id} [patch]
func UpdateCategory(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		categoryID, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
		}

		// Struct temporária para pegar os dados de entrada
		var categoryUpdates Category
		if err := c.BodyParser(&categoryUpdates); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Verifica se a categoria existe
		var existingCategory Category
		err = db.QueryRow("SELECT id, name, description, id_categories_products FROM categories_products WHERE id = ?", categoryID).Scan(
			&existingCategory.ID,
			&existingCategory.Name,
			&existingCategory.Description,
			&existingCategory.IDCategoriesProducts)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Categoria não encontrada"})
		} else if err != nil {
			log.Println("Erro ao buscar categoria:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar categoria"})
		}

		// Atualiza somente os campos que foram enviados no body
		if categoryUpdates.Name != "" {
			existingCategory.Name = categoryUpdates.Name
		}
		if categoryUpdates.Description != "" {
			existingCategory.Description = categoryUpdates.Description
		}
		// Para campos nullable, verificamos se foi enviado no request
		if categoryUpdates.IDCategoriesProducts != nil {
			existingCategory.IDCategoriesProducts = categoryUpdates.IDCategoriesProducts
		}

		// Atualiza os dados no banco de dados
		_, err = db.Exec("UPDATE categories_products SET name = ?, description = ?, id_categories_products = ? WHERE id = ?",
			existingCategory.Name, existingCategory.Description, existingCategory.IDCategoriesProducts, categoryID)
		if err != nil {
			log.Println("Erro ao atualizar categoria:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar categoria"})
		}

		return c.Status(200).JSON(existingCategory)
	}
}

// GetCategoryByID retorna uma categoria baseado no ID
// @Summary Busca uma categoria pelo ID
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path int true "ID da categoria"
// @Success 200 {object} Category
// @Failure 404 {object} map[string]string "Categoria não encontrada"
// @Failure 500 {object} map[string]string "Falha ao buscar categoria"
// @Router /categories/{id} [get]
func GetCategoryByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID da categoria inválido"})
		}

		var category Category
		query := "SELECT id, name, description, id_categories_products FROM categories_products WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&category.ID, &category.Name, &category.Description, &category.IDCategoriesProducts)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Categoria não encontrada"})
			}
			log.Println("Erro ao buscar categoria pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar categoria"})
		}

		// Retorna a categoria encontrada
		return c.Status(200).JSON(category)
	}
}

// GetCategories retorna a lista de Categorias
// @Summary Lista todas as Categorias
// @Tags Categories
// @Accept  json
// @Produce  json
// @Success 200 {array} Category
// @Router /categories [get]
func GetCategories(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, description, id_categories_products FROM categories_products")
		if err != nil {
			log.Println("Erro ao buscar categorias:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar categorias"})
		}
		defer rows.Close()

		var categories []Category
		for rows.Next() {
			var category Category
			if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.IDCategoriesProducts); err != nil {
				log.Println("Erro ao ler dados da categoria:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Falha ao ler dados da categoria"})
			}
			categories = append(categories, category)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro ao processar categorias:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar categorias"})
		}

		return c.Status(200).JSON(categories)
	}
}

// DeleteCategoryByID deleta uma categoria baseado no ID
// @Summary Deleta uma categoria pelo ID
// @Tags Categories
// @Accept  json
// @Produce  json
// @Param id path int true "ID da categoria"
// @Success 200 {object} map[string]string "Category deleted successfully"
// @Failure 404 {object} map[string]string "Category not found"
// @Failure 500 {object} map[string]string "Failed to delete category"
// @Router /categories/{id} [delete]
func DeleteCategoryByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é um número válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID da categoria inválido"})
		}

		query := "DELETE FROM categories_products WHERE id = ?"
		result, err := db.Exec(query, id)
		if err != nil {
			log.Println("Erro ao deletar categoria:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao deletar categoria"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Erro ao verificar linhas afetadas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao verificar linhas afetadas"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Categoria não encontrada"})
		}

		// Resposta com status 200 (com mensagem de sucesso)
		return c.Status(200).JSON(fiber.Map{"message": "Categoria deletada com sucesso"})
	}
}
