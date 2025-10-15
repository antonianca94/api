package controllers

import (
	"database/sql"
	"log"
	"strings"
	"github.com/gofiber/fiber/v2"
)

// Define um struct para o vendor
type Vendor struct {
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

type VendorValidationResult struct {
	CnpjExists  bool `json:"cnpj_exists"`
	EmailExists bool `json:"email_exists"`
}

// validateVendorData - Função otimizada que usa uma única query para validar CNPJ e email
func validateVendorData(db *sql.DB, cnpj, email string, excludeID int) (*VendorValidationResult, error) {
	var query string
	var params []interface{}
	
	if excludeID > 0 {
		// Para updates - exclui o próprio vendor
		query = `
			SELECT 
				COUNT(CASE WHEN cnpj = ? THEN 1 END) as cnpj_count,
				COUNT(CASE WHEN email = ? THEN 1 END) as email_count
			FROM agrofood.vendors 
			WHERE (cnpj = ? OR email = ?) AND id != ?`
		params = []interface{}{cnpj, email, cnpj, email, excludeID}
	} else {
		// Para inserts
		query = `
			SELECT 
				COUNT(CASE WHEN cnpj = ? THEN 1 END) as cnpj_count,
				COUNT(CASE WHEN email = ? THEN 1 END) as email_count
			FROM agrofood.vendors 
			WHERE cnpj = ? OR email = ?`
		params = []interface{}{cnpj, email, cnpj, email}
	}

	var cnpjCount, emailCount int
	err := db.QueryRow(query, params...).Scan(&cnpjCount, &emailCount)
	if err != nil {
		return nil, err
	}

	return &VendorValidationResult{
		CnpjExists:  cnpjCount > 0,
		EmailExists: emailCount > 0,
	}, nil
}


// @Summary Obter todos os vendors
// @Description Obtém todos os vendors
// @Tags Vendors
// @Success 200 {array} Vendor
// @Failure 500 {object} map[string]string "Erro ao buscar vendors"
// @Router /vendors [get]
func GetAllVendors(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorsQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.vendors
        `

		rows, err := db.Query(vendorsQuery)
		if err != nil {
			log.Println("Erro ao buscar vendors:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar vendors"})
		}
		defer rows.Close()

		var vendors []Vendor
		for rows.Next() {
			var vendor Vendor
			if err := rows.Scan(&vendor.ID, &vendor.Name, &vendor.Description, &vendor.Address, &vendor.Neighborhood, &vendor.City, &vendor.State, &vendor.Country, &vendor.Phone, &vendor.Email, &vendor.UsersId, &vendor.Cep, &vendor.Cnpj); err != nil {
				log.Println("Erro ao escanear vendor:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler vendor"})
			}
			vendors = append(vendors, vendor)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar vendors"})
		}

		return c.Status(200).JSON(vendors)
	}
}

// @Summary Obter vendor por ID
// @Description Obtém um vendor específico pelo ID
// @Tags Vendors
// @Param id path int true "ID do Vendor"
// @Success 200 {object} Vendor
// @Failure 404 {object} map[string]string "Vendor não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar vendor"
// @Router /vendors/{id} [get]
func GetVendorByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("id")

		vendorQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.vendors
            WHERE id = ?
        `

		var vendor Vendor
		err := db.QueryRow(vendorQuery, vendorID).Scan(&vendor.ID, &vendor.Name, &vendor.Description, &vendor.Address, &vendor.Neighborhood, &vendor.City, &vendor.State, &vendor.Country, &vendor.Phone, &vendor.Email, &vendor.UsersId, &vendor.Cep, &vendor.Cnpj)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Vendor não encontrado"})
			}
			log.Println("Erro ao buscar vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar vendor"})
		}

		return c.Status(200).JSON(vendor)
	}
}

// @Summary Criar um novo vendor
// @Description Cria um novo vendor
// @Tags Vendors
// @Accept json
// @Produce json
// @Param vendor body Vendor true "Vendor para criar"
// @Success 201 {object} Vendor
// @Failure 400 {object} map[string]string "Erro ao criar vendor"
// @Failure 500 {object} map[string]string "Erro ao criar vendor"
// @Router /vendors [post]
func CreateVendor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var vendor Vendor

		if err := c.BodyParser(&vendor); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Erro ao analisar o corpo da requisição"})
		}

		// Validação otimizada em uma única consulta
		validation, err := validateVendorData(db, vendor.Cnpj, vendor.Email, 0)
		if err != nil {
			log.Printf("Erro ao validar dados do vendor: %v", err)
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
            INSERT INTO agrofood.vendors (name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `

		result, err := db.Exec(insertQuery, vendor.Name, vendor.Description, vendor.Address, vendor.Neighborhood, vendor.City, vendor.State, vendor.Country, vendor.Phone, vendor.Email, vendor.UsersId, vendor.Cep, vendor.Cnpj)
		if err != nil {
			log.Println("Erro ao criar vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar vendor"})
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter o ID do novo vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar vendor"})
		}

		vendor.ID = int(id)
		return c.Status(200).JSON(fiber.Map{"message": "Vendor cadastrado com sucesso!"})

	}
}

// @Summary Deletar vendor por ID
// @Description Deleta um vendor específico pelo ID
// @Tags Vendors
// @Param id path int true "ID do Vendor"
// @Success 200 {object} map[string]string "Vendor deletado com sucesso"
// @Failure 404 {object} map[string]string "Vendor não encontrado"
// @Failure 500 {object} map[string]string "Erro ao deletar vendor"
// @Router /vendors/{id} [delete]
func DeleteVendor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("id")

		deleteQuery := `
            DELETE FROM agrofood.vendors
            WHERE id = ?
        `

		result, err := db.Exec(deleteQuery, vendorID)
		if err != nil {
			log.Println("Erro ao deletar vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar vendor"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Erro ao obter o número de linhas afetadas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao deletar vendor"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Vendor não encontrado"})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Vendor deletado com sucesso"})
	}
}

// @Summary Atualizar vendor por ID
// @Description Atualiza um vendor específico pelo ID (permite atualizações parciais)
// @Tags Vendors
// @Accept json
// @Produce json
// @Param id path int true "ID do Vendor"
// @Param vendor body Vendor true "Dados do vendor para atualizar"
// @Success 200 {object} Vendor
// @Failure 400 {object} map[string]string "Erro ao analisar requisição"
// @Failure 404 {object} map[string]string "Vendor não encontrado"
// @Failure 500 {object} map[string]string "Erro ao atualizar vendor"
// @Router /vendors/{id} [patch]
func UpdateVendor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("id")

		// Primeiro, verificar se o vendor existe
		checkQuery := `SELECT id FROM agrofood.vendors WHERE id = ?`
		var existingID int
		err := db.QueryRow(checkQuery, vendorID).Scan(&existingID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Vendor não encontrado"})
			}
			log.Println("Erro ao verificar vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar vendor"})
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
		args = append(args, vendorID)

		// Construir e executar a query de update
		updateQuery := `UPDATE agrofood.vendors SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`

		_, err = db.Exec(updateQuery, args...)
		if err != nil {
			log.Println("Erro ao atualizar vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar vendor"})
		}

		// Buscar e retornar o vendor atualizado
		vendorQuery := `
			SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
			FROM agrofood.vendors
			WHERE id = ?
		`

		var vendor Vendor
		err = db.QueryRow(vendorQuery, vendorID).Scan(&vendor.ID, &vendor.Name, &vendor.Description, &vendor.Address, &vendor.Neighborhood, &vendor.City, &vendor.State, &vendor.Country, &vendor.Phone, &vendor.Email, &vendor.UsersId, &vendor.Cep, &vendor.Cnpj)
		if err != nil {
			log.Println("Erro ao buscar vendor atualizado:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar vendor atualizado"})
		}

		return c.Status(200).JSON(vendor)
	}
}

// @Summary Obter vendor por User ID
// @Description Obtém um vendor específico pelo users_id
// @Tags Vendors
// @Param users_id path int true "ID do Usuário"
// @Success 200 {object} Vendor
// @Failure 404 {object} map[string]string "Vendor não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar vendor"
// @Router /vendors/user/{users_id} [get]
func GetVendorByUserID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("users_id")

		vendorQuery := `
            SELECT id, name, description, address, neighborhood, city, state, country, phone, email, users_id, cep, cnpj
            FROM agrofood.vendors
            WHERE users_id = ?
        `

		var vendor Vendor
		err := db.QueryRow(vendorQuery, userID).Scan(&vendor.ID, &vendor.Name, &vendor.Description, &vendor.Address, &vendor.Neighborhood, &vendor.City, &vendor.State, &vendor.Country, &vendor.Phone, &vendor.Email, &vendor.UsersId, &vendor.Cep, &vendor.Cnpj)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Vendor não encontrado para este usuário"})
			}
			log.Println("Erro ao buscar vendor por users_id:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar vendor"})
		}

		return c.Status(200).JSON(vendor)
	}
}