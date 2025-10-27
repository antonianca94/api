package controllers

import (
	"database/sql"
	"log"
	"strings"
	"github.com/gofiber/fiber/v2"
)

// Define um struct para o buyer
type Buyer struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Address      string `json:"address"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	UsersId      int    `json:"users_id"`
	Cep          string `json:"cep"`
	Cnpj         string `json:"cnpj"`
}

type BuyerValidationResult struct {
	CnpjExists  bool `json:"cnpj_exists"`
	EmailExists bool `json:"email_exists"`
}

// validateBuyerData - Função otimizada que usa uma única query para validar CNPJ e email
func validateBuyerData(db *sql.DB, cnpj, email string, excludeID int) (*BuyerValidationResult, error) {
	var query string
	var params []interface{}
	
	if excludeID > 0 {
		// Para updates - exclui o próprio buyer
		query = `
			SELECT 
				COUNT(CASE WHEN cnpj = ? THEN 1 END) as cnpj_count,
				COUNT(CASE WHEN email = ? THEN 1 END) as email_count
			FROM agrofood.buyers 
			WHERE (cnpj = ? OR email = ?) AND id != ?`
		params = []interface{}{cnpj, email, cnpj, email, excludeID}
	} else {
		// Para inserts
		query = `
			SELECT 
				COUNT(CASE WHEN cnpj = ? THEN 1 END) as cnpj_count,
				COUNT(CASE WHEN email = ? THEN 1 END) as email_count
			FROM agrofood.buyers 
			WHERE cnpj = ? OR email = ?`
		params = []interface{}{cnpj, email, cnpj, email}
	}

	var cnpjCount, emailCount int
	err := db.QueryRow(query, params...).Scan(&cnpjCount, &emailCount)
	if err != nil {
		return nil, err
	}

	return &BuyerValidationResult{
		CnpjExists:  cnpjCount > 0,
		EmailExists: emailCount > 0,
	}, nil
}

// @Summary Obter todos os compradores
// @Description Obtém todos os compradores
// @Tags Buyers
// @Success 200 {array} Buyer
// @Failure 500 {object} map[string]string "Erro ao buscar compradores"
// @Router /buyers [get]
func GetAllBuyers(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		buyersQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.buyers
        `

		rows, err := db.Query(buyersQuery)
		if err != nil {
			log.Println("Erro ao buscar buyers:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar compradores"})
		}
		defer rows.Close()

		var buyers []Buyer
		for rows.Next() {
			var buyer Buyer
			if err := rows.Scan(&buyer.ID, &buyer.Name, &buyer.Description, &buyer.Address, &buyer.Neighborhood, &buyer.City, &buyer.State, &buyer.Country, &buyer.Phone, &buyer.Email, &buyer.UsersId, &buyer.Cep, &buyer.Cnpj); err != nil {
				log.Println("Erro ao escanear buyer:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler comprador"})
			}
			buyers = append(buyers, buyer)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar compradores"})
		}

		return c.Status(200).JSON(buyers)
	}
}

// @Summary Obter comprador por ID
// @Description Obtém um comprador específico pelo ID
// @Tags Buyers
// @Param id path int true "ID do Comprador"
// @Success 200 {object} Buyer
// @Failure 404 {object} map[string]string "Comprador não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar comprador"
// @Router /buyers/{id} [get]
func GetBuyerByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		buyerID := c.Params("id")

		buyerQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.buyers
            WHERE id = ?
        `

		var buyer Buyer
		err := db.QueryRow(buyerQuery, buyerID).Scan(&buyer.ID, &buyer.Name, &buyer.Description, &buyer.Address, &buyer.Neighborhood, &buyer.City, &buyer.State, &buyer.Country, &buyer.Phone, &buyer.Email, &buyer.UsersId, &buyer.Cep, &buyer.Cnpj)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Comprador não encontrado"})
			}
			log.Println("Erro ao buscar buyer:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar comprador"})
		}

		return c.Status(200).JSON(buyer)
	}
}

// @Summary Criar um novo comprador
// @Description Cria um novo comprador
// @Tags Buyers
// @Accept json
// @Produce json
// @Param buyer body Buyer true "Comprador para criar"
// @Success 201 {object} Buyer
// @Failure 400 {object} map[string]string "Erro ao criar comprador"
// @Failure 500 {object} map[string]string "Erro ao criar comprador"
// @Router /buyers [post]
func CreateBuyer(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var buyer Buyer

		if err := c.BodyParser(&buyer); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Erro ao analisar o corpo da requisição"})
		}

		// Validação otimizada em uma única consulta
		validation, err := validateBuyerData(db, buyer.Cnpj, buyer.Email, 0)
		if err != nil {
			log.Printf("Erro ao validar dados do comprador: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha na validação"})
		}

		// Verificar conflitos
		if validation.CnpjExists && validation.EmailExists {
			return c.Status(400).JSON(fiber.Map{"error": "CNPJ e email já estão cadastrados"})
		}
		if validation.CnpjExists {
			return c.Status(400).JSON(fiber.Map{"error": "CNPJ já está cadastrado"})
		}
		if validation.EmailExists {
			return c.Status(400).JSON(fiber.Map{"error": "Email já está cadastrado"})
		}

		insertQuery := `
            INSERT INTO agrofood.buyers (name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `

		result, err := db.Exec(insertQuery, buyer.Name, buyer.Description, buyer.Address, buyer.Neighborhood, buyer.City, buyer.State, buyer.Country, buyer.Phone, buyer.Email, buyer.UsersId, buyer.Cep, buyer.Cnpj)
		if err != nil {
			log.Println("Erro ao criar comprador:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar comprador"})
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter o ID do novo comprador:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar comprador"})
		}

		buyer.ID = int(id)
		return c.Status(200).JSON(fiber.Map{"message": "Comprador cadastrado com sucesso!"})
	}
}

// @Summary Deletar comprador por ID
// @Description Deleta um comprador específico pelo ID
// @Tags Buyers
// @Param id path int true "ID do Comprador"
// @Success 200 {object} map[string]string "Comprador deletado com sucesso"
// @Failure 404 {object} map[string]string "Comprador não encontrado"
// @Failure 500 {object} map[string]string "Erro ao deletar comprador"
// @Router /buyers/{id} [delete]
func DeleteBuyer(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		buyerID := c.Params("id")

		deleteQuery := `
            DELETE FROM agrofood.buyers
            WHERE id = ?
        `

		result, err := db.Exec(deleteQuery, buyerID)
		if err != nil {
			log.Println("Erro ao deletar buyer:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar comprador"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Erro ao obter o número de linhas afetadas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar comprador"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Comprador não encontrado"})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Comprador deletado com sucesso"})
	}
}

// @Summary Atualizar comprador por ID
// @Description Atualiza um comprador específico pelo ID (permite atualizações parciais)
// @Tags Buyers
// @Accept json
// @Produce json
// @Param id path int true "ID do Comprador"
// @Param buyer body Buyer true "Dados do comprador para atualizar"
// @Success 200 {object} Buyer
// @Failure 400 {object} map[string]string "Erro ao analisar requisição"
// @Failure 404 {object} map[string]string "Comprador não encontrado"
// @Failure 500 {object} map[string]string "Erro ao atualizar comprador"
// @Router /buyers/{id} [patch]
func UpdateBuyer(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		buyerID := c.Params("id")

		// Primeiro, verificar se o buyer existe
		checkQuery := `SELECT id FROM agrofood.buyers WHERE id = ?`
		var existingID int
		err := db.QueryRow(checkQuery, buyerID).Scan(&existingID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Comprador não encontrado"})
			}
			log.Println("Erro ao verificar buyer:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar comprador"})
		}

		// Parse do body da requisição
		var updateData map[string]interface{}
		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Erro ao analisar o corpo da requisição"})
		}

		// Construir query dinâmica baseada nos campos fornecidos
		var setParts []string
		var args []interface{}

		// Campos permitidos para atualização
		allowedFields := map[string]string{
			"name":         "name",
			"description":  "description",
			"address":      "address",
			"neighborhood": "neighborhood",
			"city":         "city",
			"state":        "state",
			"country":      "country",
			"phone":        "phone",
			"email":        "email",
			"users_id":     "users_id",
			"cep":          "cep",
			"cnpj":         "cnpj",
		}

		for field, value := range updateData {
			if dbField, exists := allowedFields[field]; exists {
				setParts = append(setParts, dbField+" = ?")
				args = append(args, value)
			}
		}

		// Verificar se há campos para atualizar
		if len(setParts) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Nenhum campo válido fornecido para atualização"})
		}

		// Adicionar o ID no final dos argumentos
		args = append(args, buyerID)

		// Construir e executar a query de update
		updateQuery := `UPDATE agrofood.buyers SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`

		_, err = db.Exec(updateQuery, args...)
		if err != nil {
			log.Println("Erro ao atualizar comprador:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar comprador"})
		}

		// Buscar e retornar o buyer atualizado
		buyerQuery := `
			SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
			FROM agrofood.buyers
			WHERE id = ?
		`

		var buyer Buyer
		err = db.QueryRow(buyerQuery, buyerID).Scan(&buyer.ID, &buyer.Name, &buyer.Description, &buyer.Address, &buyer.Neighborhood, &buyer.City, &buyer.State, &buyer.Country, &buyer.Phone, &buyer.Email, &buyer.UsersId, &buyer.Cep, &buyer.Cnpj)
		if err != nil {
			log.Println("Erro ao buscar comprador atualizado:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar comprador atualizado"})
		}

		return c.Status(200).JSON(buyer)
	}
}

// @Summary Obter comprador por User ID
// @Description Obtém um comprador específico pelo users_id
// @Tags Buyers
// @Param users_id path int true "ID do Usuário"
// @Success 200 {object} Buyer
// @Failure 404 {object} map[string]string "Comprador não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar comprador"
// @Router /buyers/user/{users_id} [get]
func GetBuyerByUserID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("users_id")

		buyerQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.buyers
            WHERE users_id = ?
        `

		var buyer Buyer
		err := db.QueryRow(buyerQuery, userID).Scan(&buyer.ID, &buyer.Name, &buyer.Description, &buyer.Address, &buyer.Neighborhood, &buyer.City, &buyer.State, &buyer.Country, &buyer.Phone, &buyer.Email, &buyer.UsersId, &buyer.Cep, &buyer.Cnpj)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Comprador não encontrado para este usuário"})
			}
			log.Println("Erro ao buscar comprador por users_id:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar comprador"})
		}

		return c.Status(200).JSON(buyer)
	}
}