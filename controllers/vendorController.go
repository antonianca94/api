package controllers

import (
	"database/sql"
	"fmt" // ⭐ Adicione se não tiver
	"log"
	"strconv"
	"strings"
	"sync" // ⭐ ADICIONAR
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	orderCounter     int64
	lastOrderTime    int64
	orderCounterLock sync.Mutex
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

// Struct para resposta do checkout multi-vendor
type CheckoutResponse struct {
	Success     bool          `json:"success"`
	Message     string        `json:"message"`
	TotalOrders int           `json:"total_orders"`
	Orders      []OrderDetail `json:"orders"` // ✅ Com info do vendor
}

type CheckoutRequest struct {
	PaymentMethod   string `json:"payment_method"`
	ShippingAddress string `json:"shipping_address"`
	ShippingCity    string `json:"shipping_city"`
	ShippingState   string `json:"shipping_state"`
	ShippingCEP     string `json:"shipping_cep"`
	BuyersID        *int   `json:"buyers_id,omitempty"`
}

// Struct para pedido com vendor
type OrderWithVendor struct {
	ID              int     `json:"id"`
	OrderNumber     string  `json:"order_number"`
	Status          string  `json:"status"`
	Total           float64 `json:"total"`
	PaymentMethod   string  `json:"payment_method"`
	ShippingAddress string  `json:"shipping_address"`
	ShippingCity    string  `json:"shipping_city"`
	ShippingState   string  `json:"shipping_state"`
	ShippingCEP     string  `json:"shipping_cep"`
	CreatedAt       string  `json:"created_at"`
	UsersID         int     `json:"users_id"`
	VendorsID       int     `json:"vendors_id"`
	BuyersID        *int    `json:"buyers_id,omitempty"`
}

// Struct para item do pedido com informações do vendor
type OrderItemWithVendor struct {
	ID          int     `json:"id"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	OrdersID    int     `json:"orders_id"`
	ProductsID  int     `json:"products_id"`
	ProductName string  `json:"product_name"`
	VendorID    int     `json:"vendor_id"`
	VendorName  string  `json:"vendor_name"`
}

// Define struct para pedido
type Order struct {
	ID              int     `json:"id"`
	OrderNumber     string  `json:"order_number"`
	Status          string  `json:"status"`
	Total           float64 `json:"total"`
	PaymentMethod   string  `json:"payment_method"`
	ShippingAddress string  `json:"shipping_address"`
	ShippingCity    string  `json:"shipping_city"`
	ShippingState   string  `json:"shipping_state"`
	ShippingCEP     string  `json:"shipping_cep"`
	CreatedAt       string  `json:"created_at"`
	UsersID         int     `json:"users_id"`
	BuyersID        *int    `json:"buyers_id,omitempty"`
	VendorsID       int     `json:"vendors_id"`
}

// Define struct para itens do pedido
type OrderItem struct {
	ID         int     `json:"id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	OrdersID   int     `json:"orders_id"`
	ProductsID int     `json:"products_id"`
}

type VendorInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// OrderDetail (para responses detalhadas com informações do vendor)
type OrderDetail struct {
	ID              int        `json:"id"`
	OrderNumber     string     `json:"order_number"`
	Status          string     `json:"status"`
	Total           float64    `json:"total"`
	PaymentMethod   string     `json:"payment_method"`
	ShippingAddress string     `json:"shipping_address"`
	ShippingCity    string     `json:"shipping_city"`
	ShippingState   string     `json:"shipping_state"`
	ShippingCEP     string     `json:"shipping_cep"`
	CreatedAt       string     `json:"created_at"`
	UsersID         int        `json:"users_id"`
	BuyersID        *int       `json:"buyers_id,omitempty"`
	Vendor          VendorInfo `json:"vendor"`
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

func GetVendorOrders(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("vendor_id")

		if vendorID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do vendor inválido"})
		}

		// Query corrigida - removido u.email que não existe
		query := `
			SELECT 
				o.id, o.order_number, o.status, o.total, o.payment_method,
				o.shipping_address, o.shipping_city, o.shipping_state, o.shipping_cep,
				o.created_at, o.users_id, o.vendors_id, o.buyers_id,
				u.name as buyer_name
			FROM orders o
			INNER JOIN users u ON o.users_id = u.id
			WHERE o.vendors_id = ?
			ORDER BY o.created_at DESC
		`

		rows, err := db.Query(query, vendorID)
		if err != nil {
			log.Println("Erro ao buscar pedidos do vendor:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedidos"})
		}
		defer rows.Close()

		type OrderWithBuyer struct {
			Order
			BuyerName string `json:"buyer_name"`
		}

		var orders []OrderWithBuyer
		for rows.Next() {
			var order OrderWithBuyer
			// Removido &order.BuyerEmail do Scan
			err := rows.Scan(
				&order.ID, &order.OrderNumber, &order.Status, &order.Total,
				&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
				&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
				&order.UsersID, &order.VendorsID, &order.BuyersID,
				&order.BuyerName,
			)
			if err != nil {
				log.Println("Erro ao ler pedido:", err)
				continue
			}
			orders = append(orders, order)
		}

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"total":   len(orders),
			"orders":  orders,
		})
	}
}

// GetVendorOrderDetails - VERSÃO CORRIGIDA
// Substitua a função existente no seu arquivo de controllers

func GetVendorOrderDetails(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("vendor_id")
		orderID := c.Params("order_id")

		var order struct {
			ID              int     `json:"id"`
			OrderNumber     string  `json:"order_number"`
			Status          string  `json:"status"`
			Total           float64 `json:"total"`
			PaymentMethod   string  `json:"payment_method"`
			ShippingAddress string  `json:"shipping_address"`
			ShippingCity    string  `json:"shipping_city"`
			ShippingState   string  `json:"shipping_state"`
			ShippingCEP     string  `json:"shipping_cep"`
			CreatedAt       string  `json:"created_at"`
			UsersID         int     `json:"users_id"`
			VendorsID       int     `json:"vendors_id"`
			BuyerName       string  `json:"buyer_name"`
			BuyerPhone      string  `json:"buyer_phone"`
		}

		// Query corrigida - busca email e phone da tabela buyers se existir
		orderQuery := `
			SELECT 
				o.id, o.order_number, o.status, o.total, o.payment_method,
				o.shipping_address, o.shipping_city, o.shipping_state, o.shipping_cep,
				o.created_at, o.users_id, o.vendors_id,
				u.name as buyer_name,
				COALESCE(b.phone, '') as buyer_phone
			FROM orders o
			INNER JOIN users u ON o.users_id = u.id
			LEFT JOIN buyers b ON o.buyers_id = b.id
			WHERE o.id = ? AND o.vendors_id = ?
		`

		err := db.QueryRow(orderQuery, orderID, vendorID).Scan(
			&order.ID, &order.OrderNumber, &order.Status, &order.Total,
			&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
			&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
			&order.UsersID, &order.VendorsID,
			&order.BuyerName, &order.BuyerPhone,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{
					"error": "Pedido não encontrado ou não pertence a este vendor",
				})
			}
			log.Println("Erro ao buscar pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedido"})
		}

		// Buscar itens do pedido
		itemsQuery := `
			SELECT 
				oi.id, oi.quantity, oi.price, oi.products_id,
				p.name as product_name, p.sku as product_sku
			FROM order_items oi
			INNER JOIN products p ON oi.products_id = p.id
			WHERE oi.orders_id = ?
		`

		rows, err := db.Query(itemsQuery, orderID)
		if err != nil {
			log.Println("Erro ao buscar itens:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar itens"})
		}
		defer rows.Close()

		type OrderItemDetail struct {
			ID          int     `json:"id"`
			Quantity    int     `json:"quantity"`
			Price       float64 `json:"price"`
			ProductID   int     `json:"product_id"`
			ProductName string  `json:"product_name"`
			ProductSKU  string  `json:"product_sku"`
			Subtotal    float64 `json:"subtotal"`
		}

		var items []OrderItemDetail
		for rows.Next() {
			var item OrderItemDetail
			err := rows.Scan(&item.ID, &item.Quantity, &item.Price,
				&item.ProductID, &item.ProductName, &item.ProductSKU)
			if err != nil {
				log.Println("Erro ao ler item:", err)
				continue
			}
			item.Subtotal = item.Price * float64(item.Quantity)
			items = append(items, item)
		}

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"order":   order,
			"items":   items,
		})
	}
}

// UpdateOrderStatus atualiza o status de um pedido
func UpdateOrderStatus(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("vendor_id")
		orderID := c.Params("order_id")

		var statusUpdate struct {
			Status string `json:"status"`
		}

		if err := c.BodyParser(&statusUpdate); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
		}

		// Validar status
		validStatuses := map[string]bool{
			"pending":    true,
			"processing": true,
			"shipped":    true,
			"delivered":  true,
			"cancelled":  true,
		}

		if !validStatuses[statusUpdate.Status] {
			return c.Status(400).JSON(fiber.Map{
				"error": "Status inválido. Use: pending, processing, shipped, delivered ou cancelled",
			})
		}

		// Verificar se o pedido pertence ao vendor
		var currentVendorID int
		err := db.QueryRow("SELECT vendors_id FROM orders WHERE id = ?", orderID).Scan(&currentVendorID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pedido não encontrado"})
			}
			log.Println("Erro ao verificar pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar pedido"})
		}

		// Converter vendorID para int
		vendorIDInt, err := strconv.Atoi(vendorID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID do vendor inválido"})
		}

		// Verificar se o vendor tem permissão
		if currentVendorID != vendorIDInt {
			return c.Status(403).JSON(fiber.Map{
				"error": "Você não tem permissão para atualizar este pedido",
			})
		}

		// Atualizar status
		_, err = db.Exec("UPDATE orders SET status = ? WHERE id = ?",
			statusUpdate.Status, orderID)
		if err != nil {
			log.Println("Erro ao atualizar status:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar status"})
		}

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"message": "Status atualizado com sucesso",
			"status":  statusUpdate.Status,
		})
	}
}

// GetVendorOrdersStatistics retorna estatísticas dos pedidos do vendor
func GetVendorOrdersStatistics(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vendorID := c.Params("vendor_id")

		query := `
			SELECT 
				COUNT(*) as total_orders,
				COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END), 0) as processing,
				COALESCE(SUM(CASE WHEN status = 'shipped' THEN 1 ELSE 0 END), 0) as shipped,
				COALESCE(SUM(CASE WHEN status = 'delivered' THEN 1 ELSE 0 END), 0) as delivered,
				COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0) as cancelled,
				COALESCE(SUM(total), 0) as total_revenue
			FROM orders
			WHERE vendors_id = ?
		`

		var stats struct {
			TotalOrders  int     `json:"total_orders"`
			Pending      int     `json:"pending"`
			Processing   int     `json:"processing"`
			Shipped      int     `json:"shipped"`
			Delivered    int     `json:"delivered"`
			Cancelled    int     `json:"cancelled"`
			TotalRevenue float64 `json:"total_revenue"`
		}

		err := db.QueryRow(query, vendorID).Scan(
			&stats.TotalOrders,
			&stats.Pending,
			&stats.Processing,
			&stats.Shipped,
			&stats.Delivered,
			&stats.Cancelled,
			&stats.TotalRevenue,
		)

		if err != nil {
			log.Println("Erro ao buscar estatísticas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar estatísticas"})
		}

		return c.Status(200).JSON(fiber.Map{
			"success":    true,
			"statistics": stats,
		})
	}
}

// Gera número de pedido único
func generateOrderNumber() string {
	orderCounterLock.Lock()
	defer orderCounterLock.Unlock()

	now := time.Now()
	currentSecond := now.Unix()

	// Se mudou de segundo, resetar contador
	if currentSecond != lastOrderTime {
		orderCounter = 0
		lastOrderTime = currentSecond
	} else {
		orderCounter++
	}

	// Formato: ORD-YYYYMMDDHHMMSS-XXXX
	timestamp := now.Format("20060102150405")
	return fmt.Sprintf("ORD-%s-%04d", timestamp, orderCounter)
}

// FinalizeCheckout processa a finalização da compra
// @Summary Finaliza a compra do carrinho
// @Tags Checkout
// @Accept  json
// @Produce  json
// @Param user_id path int true "ID do usuário"
// @Param checkout body CheckoutRequest true "Dados do checkout"
// @Success 201 {object} Order "Pedido criado com sucesso"
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Erro ao processar pedido"
// @Router /checkout/{user_id} [post]
func FinalizeCheckout(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")

		// Validar ID do usuário
		if userID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
		}

		// Parse dos dados do checkout
		var checkoutData CheckoutRequest
		if err := c.BodyParser(&checkoutData); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Validar campos obrigatórios
		if checkoutData.PaymentMethod == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Método de pagamento é obrigatório"})
		}
		if checkoutData.ShippingAddress == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Endereço de entrega é obrigatório"})
		}
		if checkoutData.ShippingCity == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Cidade é obrigatória"})
		}
		if checkoutData.ShippingState == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Estado é obrigatório"})
		}
		if checkoutData.ShippingCEP == "" {
			return c.Status(400).JSON(fiber.Map{"error": "CEP é obrigatório"})
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			log.Println("Erro ao iniciar transação:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar pedido"})
		}
		defer tx.Rollback()

		// Buscar carrinho do usuário
		var cart Cart
		err = tx.QueryRow("SELECT id, code, created_at, users_id FROM cart WHERE users_id = ?", userID).
			Scan(&cart.ID, &cart.Code, &cart.CreatedAt, &cart.UsersID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
			}
			log.Println("Erro ao buscar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar carrinho"})
		}

		// Buscar itens do carrinho
		rows, err := tx.Query("SELECT id, quantity, cart_id, products_id FROM cart_items WHERE cart_id = ?", cart.ID)
		if err != nil {
			log.Println("Erro ao buscar itens do carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar itens do carrinho"})
		}
		defer rows.Close()

		var cartItems []CartItem
		for rows.Next() {
			var item CartItem
			if err := rows.Scan(&item.ID, &item.Quantity, &item.CartID, &item.ProductsID); err != nil {
				log.Println("Erro ao ler item do carrinho:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar itens"})
			}
			cartItems = append(cartItems, item)
		}

		// Verificar se há itens no carrinho
		if len(cartItems) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Carrinho vazio"})
		}

		// Calcular total e verificar estoque
		var orderTotal float64
		for _, item := range cartItems {
			var price float64
			var availableStock int
			err := tx.QueryRow("SELECT price, quantity FROM products WHERE id = ?", item.ProductsID).
				Scan(&price, &availableStock)
			if err != nil {
				log.Println("Erro ao buscar produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
			}

			// Verificar estoque
			if availableStock < item.Quantity {
				return c.Status(400).JSON(fiber.Map{
					"error": "Estoque insuficiente para um ou mais produtos",
				})
			}

			orderTotal += price * float64(item.Quantity)
		}

		// Criar pedido
		orderNumber := generateOrderNumber()
		createdAt := time.Now().Format("2006-01-02 15:04:05")

		orderQuery := `INSERT INTO orders 
			(order_number, status, total, payment_method, shipping_address, shipping_city, 
			shipping_state, shipping_cep, created_at, users_id, buyers_id) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		result, err := tx.Exec(orderQuery,
			orderNumber,
			"pending",
			orderTotal,
			checkoutData.PaymentMethod,
			checkoutData.ShippingAddress,
			checkoutData.ShippingCity,
			checkoutData.ShippingState,
			checkoutData.ShippingCEP,
			createdAt,
			userID,
			checkoutData.BuyersID,
		)
		if err != nil {
			log.Println("Erro ao criar pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar pedido"})
		}

		orderID, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID do pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar pedido"})
		}

		// Criar itens do pedido e atualizar estoque
		for _, item := range cartItems {
			var price float64
			err := tx.QueryRow("SELECT price FROM products WHERE id = ?", item.ProductsID).Scan(&price)
			if err != nil {
				log.Println("Erro ao buscar preço do produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
			}

			// Inserir item do pedido
			_, err = tx.Exec(`INSERT INTO order_items (quantity, price, orders_id, products_id) 
				VALUES (?, ?, ?, ?)`,
				item.Quantity, price, orderID, item.ProductsID)
			if err != nil {
				log.Println("Erro ao criar item do pedido:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar itens do pedido"})
			}

			// Atualizar estoque
			_, err = tx.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?",
				item.Quantity, item.ProductsID)
			if err != nil {
				log.Println("Erro ao atualizar estoque:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar estoque"})
			}
		}

		// Limpar carrinho
		_, err = tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cart.ID)
		if err != nil {
			log.Println("Erro ao limpar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao limpar carrinho"})
		}

		_, err = tx.Exec("DELETE FROM cart WHERE id = ?", cart.ID)
		if err != nil {
			log.Println("Erro ao deletar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao limpar carrinho"})
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			log.Println("Erro ao finalizar transação:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao finalizar pedido"})
		}

		// Retornar pedido criado
		order := Order{
			ID:              int(orderID),
			OrderNumber:     orderNumber,
			Status:          "pending",
			Total:           orderTotal,
			PaymentMethod:   checkoutData.PaymentMethod,
			ShippingAddress: checkoutData.ShippingAddress,
			ShippingCity:    checkoutData.ShippingCity,
			ShippingState:   checkoutData.ShippingState,
			ShippingCEP:     checkoutData.ShippingCEP,
			CreatedAt:       createdAt,
			UsersID:         *cart.UsersID,
			BuyersID:        checkoutData.BuyersID,
		}

		return c.Status(200).JSON(order)
	}
}

// GetOrderByID retorna um pedido pelo ID
// @Summary Busca um pedido pelo ID
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param id path int true "ID do pedido"
// @Success 200 {object} Order
// @Failure 404 {object} map[string]string "Pedido não encontrado"
// @Router /orders/{id} [get]
func GetOrderByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var order Order
		// ⭐ ADICIONADO vendors_id na query
		query := `SELECT id, order_number, status, total, payment_method, 
			shipping_address, shipping_city, shipping_state, shipping_cep, 
			created_at, users_id, vendors_id, buyers_id 
			FROM orders WHERE id = ?`

		// ⭐ ADICIONADO &order.VendorsID no Scan
		err := db.QueryRow(query, id).Scan(
			&order.ID, &order.OrderNumber, &order.Status, &order.Total,
			&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
			&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
			&order.UsersID, &order.VendorsID, &order.BuyersID,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pedido não encontrado"})
			}
			log.Println("Erro ao buscar pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedido"})
		}

		return c.Status(200).JSON(order)
	}
}

// GetUserOrders retorna todos os pedidos de um usuário
// @Summary Lista pedidos de um usuário
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param user_id path int true "ID do usuário"
// @Success 200 {array} Order
// @Router /orders/user/{user_id} [get]
func GetUserOrders(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")

		// ⭐ ADICIONADO vendors_id na query
		query := `SELECT id, order_number, status, total, payment_method, 
			shipping_address, shipping_city, shipping_state, shipping_cep, 
			created_at, users_id, vendors_id, buyers_id 
			FROM orders WHERE users_id = ? ORDER BY created_at DESC`

		rows, err := db.Query(query, userID)
		if err != nil {
			log.Println("Erro ao buscar pedidos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedidos"})
		}
		defer rows.Close()

		var orders []Order
		for rows.Next() {
			var order Order
			// ⭐ ADICIONADO &order.VendorsID no Scan
			err := rows.Scan(
				&order.ID, &order.OrderNumber, &order.Status, &order.Total,
				&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
				&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
				&order.UsersID, &order.VendorsID, &order.BuyersID,
			)
			if err != nil {
				log.Println("Erro ao ler pedido:", err)
				continue
			}
			orders = append(orders, order)
		}

		return c.Status(200).JSON(orders)
	}
}

// FinalizeCheckoutMultiVendor processa checkout separando por vendors
// @Summary Finaliza a compra criando pedidos separados por vendor
// @Tags Checkout
// @Accept  json
// @Produce  json
// @Param user_id path int true "ID do usuário"
// @Param checkout body CheckoutRequest true "Dados do checkout"
// @Success 200 {object} CheckoutResponse "Pedidos criados com sucesso"
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Erro ao processar pedido"
// @Router /checkout-multi-vendor/{user_id} [post]
func FinalizeCheckoutMultiVendor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")
		log.Printf("🔍 [CHECKOUT] UserID: %s", userID)

		if userID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
		}

		var checkoutData CheckoutRequest
		if err := c.BodyParser(&checkoutData); err != nil {
			log.Printf("❌ [CHECKOUT] Erro no parse: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Validações
		if checkoutData.PaymentMethod == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Método de pagamento é obrigatório"})
		}
		if checkoutData.ShippingAddress == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Endereço de entrega é obrigatório"})
		}
		if checkoutData.ShippingCity == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Cidade é obrigatória"})
		}
		if checkoutData.ShippingState == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Estado é obrigatório"})
		}
		if checkoutData.ShippingCEP == "" {
			return c.Status(400).JSON(fiber.Map{"error": "CEP é obrigatório"})
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			log.Println("❌ [CHECKOUT] Erro ao iniciar transação:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar pedido"})
		}
		defer tx.Rollback()

		// Buscar carrinho
		var cart Cart
		err = tx.QueryRow("SELECT id, code, created_at, users_id FROM cart WHERE users_id = ?", userID).
			Scan(&cart.ID, &cart.Code, &cart.CreatedAt, &cart.UsersID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
			}
			log.Println("❌ [CHECKOUT] Erro ao buscar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar carrinho"})
		}

		log.Printf("✅ [CHECKOUT] Carrinho encontrado - ID: %d", cart.ID)

		// ⭐ Query otimizada com JOIN para pegar vendor info
		query := `
			SELECT 
				ci.id, ci.quantity, ci.cart_id, ci.products_id,
				p.name, p.price, p.quantity as stock,
				v.id as vendor_id,
				v.name as vendor_name,
				v.email as vendor_email,
				v.phone as vendor_phone
			FROM cart_items ci
			INNER JOIN products p ON ci.products_id = p.id
			INNER JOIN vendors v ON p.users_id = v.users_id
			WHERE ci.cart_id = ?
		`

		rows, err := tx.Query(query, cart.ID)
		if err != nil {
			log.Println("❌ [CHECKOUT] Erro ao buscar itens:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar itens do carrinho"})
		}
		defer rows.Close()

		// ⭐ Estrutura para agrupar por vendor com info completa
		type VendorGroup struct {
			VendorID    int
			VendorName  string
			VendorEmail string
			VendorPhone string
			Items       []struct {
				CartItemID  int
				Quantity    int
				ProductID   int
				ProductName string
				Price       float64
				Stock       int
			}
		}

		vendorGroups := make(map[int]*VendorGroup)

		itemCount := 0
		for rows.Next() {
			var (
				cartItemID                           int
				quantity, cartID, productID          int
				productName                          string
				price                                float64
				stock, vendorID                      int
				vendorName, vendorEmail, vendorPhone string
			)

			if err := rows.Scan(&cartItemID, &quantity, &cartID,
				&productID, &productName, &price, &stock,
				&vendorID, &vendorName, &vendorEmail, &vendorPhone); err != nil {
				log.Println("❌ [CHECKOUT] Erro ao ler item:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar itens"})
			}

			itemCount++
			log.Printf("✅ [CHECKOUT] Item %d: %s (Vendor: %s)", itemCount, productName, vendorName)

			// Verificar estoque
			if stock < quantity {
				log.Printf("❌ [CHECKOUT] Estoque insuficiente: %s", productName)
				return c.Status(400).JSON(fiber.Map{
					"error": "Estoque insuficiente para: " + productName,
				})
			}

			// Criar ou atualizar grupo do vendor
			if _, exists := vendorGroups[vendorID]; !exists {
				vendorGroups[vendorID] = &VendorGroup{
					VendorID:    vendorID,
					VendorName:  vendorName,
					VendorEmail: vendorEmail,
					VendorPhone: vendorPhone,
					Items: []struct {
						CartItemID  int
						Quantity    int
						ProductID   int
						ProductName string
						Price       float64
						Stock       int
					}{},
				}
			}

			// Adicionar item ao grupo
			vendorGroups[vendorID].Items = append(vendorGroups[vendorID].Items, struct {
				CartItemID  int
				Quantity    int
				ProductID   int
				ProductName string
				Price       float64
				Stock       int
			}{
				CartItemID:  cartItemID,
				Quantity:    quantity,
				ProductID:   productID,
				ProductName: productName,
				Price:       price,
				Stock:       stock,
			})
		}

		log.Printf("📊 [CHECKOUT] Total itens: %d | Vendors: %d", itemCount, len(vendorGroups))

		if len(vendorGroups) == 0 {
			log.Println("❌ [CHECKOUT] Carrinho vazio")
			return c.Status(400).JSON(fiber.Map{"error": "Carrinho vazio"})
		}

		createdAt := time.Now().Format("2006-01-02 15:04:05")
		var createdOrders []OrderDetail // ⭐ Usar OrderDetail

		// Criar pedido para cada vendor
		for vendorID, group := range vendorGroups {
			var orderTotal float64

			log.Printf("🛒 [CHECKOUT] Processando vendor: %s (ID: %d)", group.VendorName, vendorID)

			// Calcular total
			for _, item := range group.Items {
				orderTotal += item.Price * float64(item.Quantity)
			}

			orderNumber := generateOrderNumber()

			// Inserir pedido
			orderQuery := `INSERT INTO orders 
				(order_number, status, total, payment_method, shipping_address, 
				shipping_city, shipping_state, shipping_cep, created_at, users_id, 
				vendors_id, buyers_id) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			result, err := tx.Exec(orderQuery,
				orderNumber, "pending", orderTotal,
				checkoutData.PaymentMethod,
				checkoutData.ShippingAddress,
				checkoutData.ShippingCity,
				checkoutData.ShippingState,
				checkoutData.ShippingCEP,
				createdAt, userID, vendorID,
				checkoutData.BuyersID,
			)
			if err != nil {
				log.Printf("❌ [CHECKOUT] Erro ao criar pedido vendor %d: %v", vendorID, err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar pedido"})
			}

			orderID, err := result.LastInsertId()
			if err != nil {
				log.Printf("❌ [CHECKOUT] Erro ao obter ID do pedido: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar pedido"})
			}

			log.Printf("✅ [CHECKOUT] Pedido criado - #%s (ID: %d)", orderNumber, orderID)

			// Criar itens e atualizar estoque
			for _, item := range group.Items {
				_, err = tx.Exec(`INSERT INTO order_items (quantity, price, orders_id, products_id) 
					VALUES (?, ?, ?, ?)`,
					item.Quantity, item.Price, orderID, item.ProductID)
				if err != nil {
					log.Printf("❌ [CHECKOUT] Erro ao criar item: %v", err)
					return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar itens do pedido"})
				}

				_, err = tx.Exec("UPDATE products SET quantity = quantity - ? WHERE id = ?",
					item.Quantity, item.ProductID)
				if err != nil {
					log.Printf("❌ [CHECKOUT] Erro ao atualizar estoque: %v", err)
					return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar estoque"})
				}
			}

			// ⭐ Adicionar OrderDetail com VendorInfo completo
			createdOrders = append(createdOrders, OrderDetail{
				ID:              int(orderID),
				OrderNumber:     orderNumber,
				Status:          "pending",
				Total:           orderTotal,
				PaymentMethod:   checkoutData.PaymentMethod,
				ShippingAddress: checkoutData.ShippingAddress,
				ShippingCity:    checkoutData.ShippingCity,
				ShippingState:   checkoutData.ShippingState,
				ShippingCEP:     checkoutData.ShippingCEP,
				CreatedAt:       createdAt,
				UsersID:         *cart.UsersID,
				BuyersID:        checkoutData.BuyersID,
				Vendor: VendorInfo{
					ID:    vendorID,
					Name:  group.VendorName,
					Email: group.VendorEmail,
					Phone: group.VendorPhone,
				},
			})
		}

		// Limpar carrinho
		_, err = tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cart.ID)
		if err != nil {
			log.Println("❌ [CHECKOUT] Erro ao limpar cart_items:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao limpar carrinho"})
		}

		_, err = tx.Exec("DELETE FROM cart WHERE id = ?", cart.ID)
		if err != nil {
			log.Println("❌ [CHECKOUT] Erro ao deletar cart:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao limpar carrinho"})
		}

		// Commit
		if err := tx.Commit(); err != nil {
			log.Println("❌ [CHECKOUT] Erro no commit:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao finalizar pedido"})
		}

		log.Printf("🎉 [CHECKOUT] Sucesso! %d pedidos criados", len(createdOrders))

		// ⭐ Response otimizado para multi-vendor
		response := CheckoutResponse{
			Success:     true,
			Message:     "Pedidos criados com sucesso",
			TotalOrders: len(createdOrders),
			Orders:      createdOrders,
		}

		return c.Status(200).JSON(response)
	}
}

// GetOrdersByVendor retorna pedidos agrupados por vendor
// @Summary Lista pedidos de um usuário agrupados por vendor
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param user_id path int true "ID do usuário"
// @Success 200 {object} map[string]interface{}
// @Router /orders/user/{user_id}/by-vendor [get]
func GetOrdersByVendor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")

		query := `
			SELECT 
				o.id, o.order_number, o.status, o.total, o.payment_method,
				o.shipping_address, o.shipping_city, o.shipping_state, o.shipping_cep,
				o.created_at, o.users_id, o.vendors_id, o.buyers_id,
				v.name as vendor_name
			FROM orders o
			INNER JOIN vendors v ON o.vendors_id = v.id
			WHERE o.users_id = ?
			ORDER BY o.created_at DESC
		`

		rows, err := db.Query(query, userID)
		if err != nil {
			log.Println("Erro ao buscar pedidos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedidos"})
		}
		defer rows.Close()

		ordersByVendor := make(map[string][]OrderWithVendor)

		for rows.Next() {
			var order OrderWithVendor
			var vendorName string

			err := rows.Scan(
				&order.ID, &order.OrderNumber, &order.Status, &order.Total,
				&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
				&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
				&order.UsersID, &order.VendorsID, &order.BuyersID, &vendorName,
			)
			if err != nil {
				log.Println("Erro ao ler pedido:", err)
				continue
			}

			ordersByVendor[vendorName] = append(ordersByVendor[vendorName], order)
		}

		return c.Status(200).JSON(fiber.Map{
			"success":          true,
			"orders_by_vendor": ordersByVendor,
		})
	}
}

// GetOrderWithVendorInfo retorna detalhes de um pedido com informações do vendor
// @Summary Busca pedido com informações do vendor e itens
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param id path int true "ID do pedido"
// @Success 200 {object} map[string]interface{}
// @Router /orders/{id}/details [get]
func GetOrderWithVendorInfo(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orderID := c.Params("id")

		// Buscar pedido com informações do vendor
		orderQuery := `
			SELECT 
				o.id, o.order_number, o.status, o.total, o.payment_method,
				o.shipping_address, o.shipping_city, o.shipping_state, o.shipping_cep,
				o.created_at, o.users_id, o.vendors_id, o.buyers_id,
				v.name as vendor_name, v.email as vendor_email, v.phone as vendor_phone
			FROM orders o
			INNER JOIN vendors v ON o.vendors_id = v.id
			WHERE o.id = ?
		`

		var order struct {
			OrderWithVendor
			VendorName  string
			VendorEmail string
			VendorPhone string
		}

		err := db.QueryRow(orderQuery, orderID).Scan(
			&order.ID, &order.OrderNumber, &order.Status, &order.Total,
			&order.PaymentMethod, &order.ShippingAddress, &order.ShippingCity,
			&order.ShippingState, &order.ShippingCEP, &order.CreatedAt,
			&order.UsersID, &order.VendorsID, &order.BuyersID,
			&order.VendorName, &order.VendorEmail, &order.VendorPhone,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Pedido não encontrado"})
			}
			log.Println("Erro ao buscar pedido:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar pedido"})
		}

		// Buscar itens do pedido
		itemsQuery := `
			SELECT 
				oi.id, oi.quantity, oi.price, oi.orders_id, oi.products_id,
				p.name as product_name
			FROM order_items oi
			INNER JOIN products p ON oi.products_id = p.id
			WHERE oi.orders_id = ?
		`

		rows, err := db.Query(itemsQuery, orderID)
		if err != nil {
			log.Println("Erro ao buscar itens:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar itens"})
		}
		defer rows.Close()

		var items []OrderItemWithVendor
		for rows.Next() {
			var item OrderItemWithVendor
			err := rows.Scan(&item.ID, &item.Quantity, &item.Price,
				&item.OrdersID, &item.ProductsID, &item.ProductName)
			if err != nil {
				log.Println("Erro ao ler item:", err)
				continue
			}
			items = append(items, item)
		}

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"order":   order,
			"items":   items,
		})
	}
}
