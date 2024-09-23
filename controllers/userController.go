package controllers

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para o usuário
type User struct {
	ID       int    `json:"id"`
	Status   int    `json:"status"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Cpf      string `json:"cpf"`
	RolesId  int    `json:"roles_id"`
	RoleName string `json:"role_name"`
}

// Define um struct para os detalhes do usuário a partir da view
type UserDetails struct {
	UserID   int    `json:"user_id"`
	Status   int    `json:"status"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
	Cpf      string `json:"cpf"`
	RolesID  int    `json:"roles_id"`
	RoleName string `json:"role_name"`
}

// GetUsers retorna a lista de usuários
// @Summary Lista todos os usuários
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {array} User
// @Router /users [get]
func GetUsers(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT * FROM users_all")
		if err != nil {
			log.Println("Erro ao consultar usuários:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar usuários"})
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Status, &user.Username, &user.Name, &user.RolesId, &user.RoleName); err != nil {
				log.Println("Erro ao escanear usuário:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Falha ao ler dados do usuário"})
			}
			users = append(users, user)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar usuários"})
		}

		return c.Status(200).JSON(users)
	}
}

// GetUserByID retorna um único usuário baseado no ID
// @Summary Retorna um usuário pelo ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id path int true "ID do usuário"
// @Success 200 {object} User
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Failed to fetch user"
// @Router /users/{id} [get]
func GetUserByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var user User
		query := "SELECT id, status, username, password, name, surname, cpf, roles_id FROM users WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&user.ID, &user.Status, &user.Username, &user.Password, &user.Name, &user.Surname, &user.Cpf, &user.RolesId)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Usuário não encontrado"})
			}
			log.Println("Erro ao buscar usuário pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar usuário"})
		}

		return c.Status(200).JSON(user)
	}
}

// GetUserDetailsByID retorna os detalhes de um usuário baseado no ID da view
// @Summary Retorna os detalhes de um usuário pelo ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id path int true "ID do usuário"
// @Success 200 {object} UserDetails
// @Failure 404 {object} map[string]string "Detalhes do usuário não encontrados"
// @Failure 500 {object} map[string]string "Falha ao buscar detalhes do usuário"
// @Router /users/details/{id} [get]
func GetUserDetailsByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var userDetails UserDetails
		query := "SELECT * FROM user_details WHERE user_id = ?"
		err := db.QueryRow(query, id).Scan(&userDetails.UserID, &userDetails.Status, &userDetails.Username, &userDetails.Name, &userDetails.Surname, &userDetails.Password, &userDetails.Cpf, &userDetails.RolesID, &userDetails.RoleName)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Detalhes do usuário não encontrados"})
			}
			log.Println("Erro ao buscar detalhes do usuário por ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar detalhes do usuário"})
		}

		return c.Status(200).JSON(userDetails)
	}
}

// DeleteUserByID exclui um usuário baseado no ID
// @Summary Exclui um usuário pelo ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id path int true "ID do usuário"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Failed to delete user"
// @Router /users/{id} [delete]
func DeleteUserByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := "DELETE FROM users WHERE id = ?"

		// Execute o DELETE e ignore o resultado se for eficiente o suficiente
		result, err := db.Exec(query, id)
		if err != nil {
			// Use um log mais leve ou remova o log para alta performance
			log.Printf("Erro ao deletar usuário com ID %s: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao deletar usuário"})
		}

		// Verifique se a linha foi realmente deletada (opcional, dependendo do caso)
		if rowsAffected, err := result.RowsAffected(); err != nil {
			log.Printf("Erro ao verificar linhas afetadas para o usuário ID %s: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao verificar linhas afetadas"})
		} else if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Usuário não encontrado"})
		}

		// Resposta de sucesso sem payload
		return c.SendStatus(200)
	}
}

// PatchUserByID atualiza parcialmente os dados de um usuário pelo ID
// @Summary Atualiza parcialmente os dados de um usuário pelo ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id path int true "ID do usuário"
// @Param user body User true "Dados do usuário a serem atualizados"
// @Success 200 {object} map[string]string "User updated successfully"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Failed to update user"
// @Router /users/{id} [patch]
func PatchUserByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Corpo da solicitação inválido"})
		}

		// Construa dinamicamente a query para atualizar apenas os campos enviados
		query := "UPDATE users SET "
		params := []interface{}{}

		if user.Status != 0 {
			query += "status = ?, "
			params = append(params, user.Status)
		}
		if user.Username != "" {
			query += "username = ?, "
			params = append(params, user.Username)
		}
		if user.Password != "" {
			query += "password = ?, "
			params = append(params, user.Password)
		}
		if user.Name != "" {
			query += "name = ?, "
			params = append(params, user.Name)
		}
		if user.Surname != "" {
			query += "surname = ?, "
			params = append(params, user.Surname)
		}
		if user.Cpf != "" {
			query += "cpf = ?, "
			params = append(params, user.Cpf)
		}
		if user.RolesId != 0 {
			query += "roles_id = ?, "
			params = append(params, user.RolesId)
		}

		// Verifique se há campos para atualizar
		if len(params) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Nenhum campo para atualizar"})
		}

		// Remova a última vírgula e adicione a cláusula WHERE
		query = query[:len(query)-2] + " WHERE id = ?"
		params = append(params, id)

		// Execute a query com os parâmetros
		_, err := db.Exec(query, params...)
		if err != nil {
			log.Printf("Erro ao atualizar usuário com ID %s: %v", id, err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao atualizar usuário", "details": err.Error()})
		}

		// Sucesso
		return c.Status(200).JSON(fiber.Map{"message": "Usuário atualizado com sucesso"})
	}
}

// CreateUser cria um novo usuário
// @Summary Cria um novo usuário
// @Description Este endpoint cria um novo usuário
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body User true "Dados do usuário"
// @Success 200 {object} map[string]interface{} "message: User created successfully"
// @Failure 400 {object} map[string]string "error: Invalid request body"
// @Failure 500 {object} map[string]string "error: Failed to create user"
// @Router /users [post]
func CreateUser(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user User

		// Parse o corpo da requisição
		if err := c.BodyParser(&user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Corpo da solicitação inválido"})
		}

		// Verifique se os campos obrigatórios estão presentes
		if user.Username == "" || user.Password == "" || user.Name == "" || user.Surname == "" || user.Cpf == "" || user.RolesId == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Campos obrigatórios ausentes"})
		}

		// Crie a query de inserção
		query := `INSERT INTO users (status, username, password, name, surname, cpf, roles_id) 
		          VALUES (?, ?, ?, ?, ?, ?, ?)`

		// Execute a query com os parâmetros
		result, err := db.Exec(query, user.Status, user.Username, user.Password, user.Name, user.Surname, user.Cpf, user.RolesId)
		if err != nil {
			log.Printf("Erro ao inserir usuário: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar usuário", "details": err.Error()})
		}

		// Obtenha o ID do novo usuário inserido
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Erro ao obter o ID do último inserido: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao recuperar o ID do usuário"})
		}

		// Retorne uma resposta de sucesso com o ID do novo usuário
		return c.Status(200).JSON(fiber.Map{"message": "Usuário criado com sucesso", "user_id": lastInsertID})
	}
}
