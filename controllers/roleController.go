package controllers

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para o usuário
type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateRole cria uma nova Role
// @Summary Cria uma nova Role
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param role body Role true "Dados da nova Role"
// @Success 201 {object} Role "Role criada com sucesso"
// @Failure 400 {object} map[string]string "Dados de entrada inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar Role"
// @Router /roles [post]
func CreateRole(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var newRole Role

		// Lê os dados da nova Role do corpo da requisição
		if err := c.BodyParser(&newRole); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Insere a nova Role no banco de dados
		query := "INSERT INTO roles (name, description) VALUES (?, ?)"
		result, err := db.Exec(query, newRole.Name, newRole.Description)
		if err != nil {
			log.Println("Erro ao criar role:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar role"})
		}

		// Obtém o ID da nova Role inserida
		roleID, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID da nova role:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID da nova role"})
		}

		// Atribui o ID à nova Role
		newRole.ID = int(roleID)

		return c.Status(201).JSON(newRole)
	}
}

// UpdateRole atualiza parcialmente uma Role com base no ID
// @Summary Atualiza parcialmente uma Role pelo ID
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path int true "ID da Role"
// @Param role body Role true "Dados da Role para atualização parcial"
// @Success 200 {object} Role "Role atualizada com sucesso"
// @Failure 400 {object} map[string]string "ID inválido ou dados de entrada inválidos"
// @Failure 404 {object} map[string]string "Role não encontrada"
// @Failure 500 {object} map[string]string "Erro ao atualizar Role"
// @Router /roles/{id} [patch]
func UpdateRole(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		roleID, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
		}

		// Struct temporária para pegar os dados de entrada
		var roleUpdates Role
		if err := c.BodyParser(&roleUpdates); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Verifica se a role existe
		var existingRole Role
		err = db.QueryRow("SELECT id, name, description FROM roles WHERE id = ?", roleID).Scan(&existingRole.ID, &existingRole.Name, &existingRole.Description)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Role não encontrada"})
		} else if err != nil {
			log.Println("Erro ao buscar role:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar role"})
		}

		// Atualiza somente os campos que foram enviados no body
		if roleUpdates.Name != "" {
			existingRole.Name = roleUpdates.Name
		}
		if roleUpdates.Description != "" {
			existingRole.Description = roleUpdates.Description
		}

		// Atualiza os dados no banco de dados
		_, err = db.Exec("UPDATE roles SET name = ?, description = ? WHERE id = ?", existingRole.Name, existingRole.Description, roleID)
		if err != nil {
			log.Println("Erro ao atualizar role:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar role"})
		}

		return c.Status(200).JSON(existingRole)
	}
}

// GetRoleByID retorna uma role baseado no ID
// @Summary Busca uma role pelo ID
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path int true "ID da role"
// @Success 200 {object} Role
// @Failure 404 {object} map[string]string "Role não encontrada"
// @Failure 500 {object} map[string]string "Falha ao buscar role"
// @Router /roles/{id} [get]
func GetRoleByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID da role inválido"})
		}

		var role Role
		query := "SELECT id, name, description FROM roles WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Role não encontrada"})
			}
			log.Println("Erro ao buscar role pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar role"})
		}

		// Retorna a role encontrada
		return c.Status(200).JSON(role)
	}
}

// GetRoles retorna a lista as Roles
// @Summary Lista todas as Roles
// @Tags Roles
// @Accept  json
// @Produce  json
// @Success 200 {array} Role
// @Router /roles [get]
func GetRoles(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, description FROM roles")
		if err != nil {
			log.Println("Error querying roles:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch roles"})
		}
		defer rows.Close()

		var roles []Role
		for rows.Next() {
			var role Role
			if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
				log.Println("Error scanning role:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Failed to read role data"})
			}
			roles = append(roles, role)
		}

		if err := rows.Err(); err != nil {
			log.Println("Error with rows:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Error processing roles"})
		}

		return c.Status(200).JSON(roles)
	}
}

// DeleteRoleByID deleta uma role baseado no ID
// @Summary Deleta uma role pelo ID
// @Tags Roles
// @Accept  json
// @Produce  json
// @Param id path int true "ID da role"
// @Success 204 {object} map[string]string "Role deleted successfully"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Failed to delete role"
// @Router /roles/{id} [delete]
func DeleteRoleByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é um número válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
		}

		query := "DELETE FROM roles WHERE id = ?"
		result, err := db.Exec(query, id)
		if err != nil {
			log.Println("Error deleting role:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete role"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Error checking rows affected:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check rows affected"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
		}

		// Resposta com status 204 (sem conteúdo)
		return c.SendStatus(200)
	}
}
