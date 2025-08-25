package controllers

import (
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para carrinho de compras
type Cart struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at"`
	UsersID   *int   `json:"users_id,omitempty"`
}

// Define um struct para itens do carrinho
type CartItem struct {
	ID         int  `json:"id"`
	Quantity   int  `json:"quantity"`
	CartID     *int `json:"cart_id,omitempty"`
	ProductsID *int `json:"products_id,omitempty"`
}

// Define um struct para carrinho com itens (para busca completa)
type CartWithItems struct {
	ID        int        `json:"id"`
	Code      string     `json:"code"`
	CreatedAt string     `json:"created_at"`
	UsersID   *int       `json:"users_id,omitempty"`
	Items     []CartItem `json:"items,omitempty"`
}

// Função para gerar código aleatório
func generateRandomCode() string {
	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 9)
	for i := range code {
		code[i] = characters[rand.Intn(len(characters))]
	}
	return string(code)
}

// ==================== SHOPPING CART ENDPOINTS ====================

// CreateCart cria um novo Carrinho
// @Summary Cria um novo Carrinho de Compras
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param cart body Cart true "Dados do novo Carrinho"
// @Success 201 {object} Cart "Carrinho criado com sucesso"
// @Failure 400 {object} map[string]string "Dados de entrada inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar carrinho"
// @Router /cart [post]
func CreateCart(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var newCart Cart

		// Lê os dados do novo Carrinho do corpo da requisição
		if err := c.BodyParser(&newCart); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Gera código se não fornecido
		if newCart.Code == "" {
			newCart.Code = generateRandomCode()
		}

		// Se não foi fornecido created_at, usa o timestamp atual
		if newCart.CreatedAt == "" {
			newCart.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
		}

		// Insere o novo Carrinho no banco de dados
		query := "INSERT INTO cart (code, created_at, users_id) VALUES (?, ?, ?)"
		result, err := db.Exec(query, newCart.Code, newCart.CreatedAt, newCart.UsersID)
		if err != nil {
			log.Println("Erro ao criar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar carrinho"})
		}

		// Obtém o ID do novo Carrinho inserido
		cartID, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID do novo carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID do novo carrinho"})
		}

		// Atribui o ID ao novo Carrinho
		newCart.ID = int(cartID)

		return c.Status(201).JSON(newCart)
	}
}

// GetCarts retorna a lista de Carrinhos
// @Summary Lista todos os Carrinhos de Compras
// @Tags Cart
// @Accept  json
// @Produce  json
// @Success 200 {array} Cart
// @Router /cart [get]
func GetCarts(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, code, created_at, users_id FROM cart")
		if err != nil {
			log.Println("Erro ao buscar carrinhos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar carrinhos"})
		}
		defer rows.Close()

		var carts []Cart
		for rows.Next() {
			var cart Cart
			if err := rows.Scan(&cart.ID, &cart.Code, &cart.CreatedAt, &cart.UsersID); err != nil {
				log.Println("Erro ao ler dados do carrinho:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Falha ao ler dados do carrinho"})
			}
			carts = append(carts, cart)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro ao processar carrinhos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar carrinhos"})
		}

		return c.Status(200).JSON(carts)
	}
}

// GetCartByID retorna um carrinho baseado no ID
// @Summary Busca um carrinho pelo ID
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param id path int true "ID do carrinho"
// @Success 200 {object} Cart
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Falha ao buscar carrinho"
// @Router /cart/{id} [get]
func GetCartByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do carrinho inválido"})
		}

		var cart Cart
		query := "SELECT id, code, created_at, users_id FROM cart WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&cart.ID, &cart.Code, &cart.CreatedAt, &cart.UsersID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
			}
			log.Println("Erro ao buscar carrinho pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar carrinho"})
		}

		// Retorna o carrinho encontrado
		return c.Status(200).JSON(cart)
	}
}

// GetCartByUserID retorna um carrinho baseado no ID do usuário
// @Summary Busca um carrinho pelo ID do usuário
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param user_id path int true "ID do usuário"
// @Success 200 {object} Cart
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Falha ao buscar carrinho"
// @Router /cart/user/{user_id} [get]
func GetCartByUserID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")

		// Verifica se o ID é válido
		if userID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do usuário inválido"})
		}

		var cart Cart
		query := "SELECT id, code, created_at, users_id FROM cart WHERE users_id = ?"
		err := db.QueryRow(query, userID).Scan(&cart.ID, &cart.Code, &cart.CreatedAt, &cart.UsersID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado para este usuário"})
			}
			log.Println("Erro ao buscar carrinho pelo ID do usuário:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar carrinho"})
		}

		return c.Status(200).JSON(cart)
	}
}

// GetCartWithItems retorna um carrinho com seus itens
// @Summary Busca um carrinho com todos os seus itens
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param id path int true "ID do carrinho"
// @Success 200 {object} CartWithItems
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Falha ao buscar carrinho"
// @Router /cart/{id}/items [get]
func GetCartWithItems(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do carrinho inválido"})
		}

		var cartWithItems CartWithItems

		// Busca o carrinho
		query := "SELECT id, code, created_at, users_id FROM cart WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&cartWithItems.ID, &cartWithItems.Code, &cartWithItems.CreatedAt, &cartWithItems.UsersID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
			}
			log.Println("Erro ao buscar carrinho pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar carrinho"})
		}

		// Busca os itens do carrinho
		itemsQuery := "SELECT id, quantity, cart_id, products_id FROM cart_items WHERE cart_id = ?"
		rows, err := db.Query(itemsQuery, id)
		if err != nil {
			log.Println("Erro ao buscar itens do carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar itens do carrinho"})
		}
		defer rows.Close()

		var items []CartItem
		for rows.Next() {
			var item CartItem
			if err := rows.Scan(&item.ID, &item.Quantity, &item.CartID, &item.ProductsID); err != nil {
				log.Println("Erro ao ler dados do item:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Falha ao ler dados do item"})
			}
			items = append(items, item)
		}

		cartWithItems.Items = items

		return c.Status(200).JSON(cartWithItems)
	}
}

// UpdateCart atualiza parcialmente um Carrinho com base no ID
// @Summary Atualiza parcialmente um Carrinho pelo ID
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param id path int true "ID do Carrinho"
// @Param cart body Cart true "Dados do Carrinho para atualização parcial"
// @Success 200 {object} Cart "Carrinho atualizado com sucesso"
// @Failure 400 {object} map[string]string "ID inválido ou dados de entrada inválidos"
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Erro ao atualizar carrinho"
// @Router /cart/{id} [patch]
func UpdateCart(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		cartID, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
		}

		// Struct temporária para pegar os dados de entrada
		var cartUpdates Cart
		if err := c.BodyParser(&cartUpdates); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Verifica se o carrinho existe
		var existingCart Cart
		err = db.QueryRow("SELECT id, code, created_at, users_id FROM cart WHERE id = ?", cartID).Scan(
			&existingCart.ID, &existingCart.Code, &existingCart.CreatedAt, &existingCart.UsersID)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
		} else if err != nil {
			log.Println("Erro ao buscar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar carrinho"})
		}

		// Atualiza somente os campos que foram enviados no body
		if cartUpdates.Code != "" {
			existingCart.Code = cartUpdates.Code
		}
		if cartUpdates.UsersID != nil {
			existingCart.UsersID = cartUpdates.UsersID
		}

		// Atualiza os dados no banco de dados
		_, err = db.Exec("UPDATE cart SET code = ?, users_id = ? WHERE id = ?",
			existingCart.Code, existingCart.UsersID, cartID)
		if err != nil {
			log.Println("Erro ao atualizar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar carrinho"})
		}

		return c.Status(200).JSON(existingCart)
	}
}

// DeleteCartByID deleta um carrinho baseado no ID
// @Summary Deleta um carrinho pelo ID
// @Tags Cart
// @Accept  json
// @Produce  json
// @Param id path int true "ID do carrinho"
// @Success 200 {object} map[string]string "Carrinho deletado com sucesso"
// @Failure 404 {object} map[string]string "Carrinho não encontrado"
// @Failure 500 {object} map[string]string "Falha ao deletar carrinho"
// @Router /cart/{id} [delete]
func DeleteCartByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é um número válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do carrinho inválido"})
		}

		query := "DELETE FROM cart WHERE id = ?"
		result, err := db.Exec(query, id)
		if err != nil {
			log.Println("Erro ao deletar carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao deletar carrinho"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Erro ao verificar linhas afetadas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao verificar linhas afetadas"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Carrinho não encontrado"})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Carrinho deletado com sucesso"})
	}
}

// ==================== CART ITEMS ENDPOINTS ====================

// CreateCartItem cria um novo Item do Carrinho
// @Summary Cria um novo Item do Carrinho
// @Tags CartItems
// @Accept  json
// @Produce  json
// @Param item body CartItem true "Dados do novo Item"
// @Success 201 {object} CartItem "Item criado com sucesso"
// @Failure 400 {object} map[string]string "Dados de entrada inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar item"
// @Router /cart/cart-items [post]
func CreateCartItem(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var newItem CartItem

		// Lê os dados do novo Item do corpo da requisição
		if err := c.BodyParser(&newItem); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Verificar estoque antes de adicionar
		var availableStock int
		stockQuery := "SELECT quantity FROM products WHERE id = ?"
		err := db.QueryRow(stockQuery, newItem.ProductsID).Scan(&availableStock)
		if err != nil {
			log.Println("Erro ao verificar estoque:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar estoque do produto"})
		}

		if availableStock < newItem.Quantity {
			return c.Status(400).JSON(fiber.Map{"error": "Quantidade solicitada excede o estoque disponível"})
		}

		// Verificar se o item já existe no carrinho
		var existingQuantity int
		existingQuery := "SELECT quantity FROM cart_items WHERE cart_id = ? AND products_id = ?"
		err = db.QueryRow(existingQuery, newItem.CartID, newItem.ProductsID).Scan(&existingQuantity)

		if err == nil {
			// Item existe, atualizar quantidade
			totalQuantity := existingQuantity + newItem.Quantity
			if totalQuantity > availableStock {
				return c.Status(400).JSON(fiber.Map{"error": "Quantidade total excede o estoque disponível"})
			}

			updateQuery := "UPDATE cart_items SET quantity = ? WHERE cart_id = ? AND products_id = ?"
			_, err = db.Exec(updateQuery, totalQuantity, newItem.CartID, newItem.ProductsID)
			if err != nil {
				log.Println("Erro ao atualizar item no carrinho:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar item no carrinho"})
			}

			// Buscar o item atualizado para retornar
			query := "SELECT id, quantity, cart_id, products_id FROM cart_items WHERE cart_id = ? AND products_id = ?"
			err = db.QueryRow(query, newItem.CartID, newItem.ProductsID).Scan(
				&newItem.ID, &newItem.Quantity, &newItem.CartID, &newItem.ProductsID)
			if err != nil {
				log.Println("Erro ao buscar item atualizado:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar item atualizado"})
			}

			return c.Status(200).JSON(newItem)
		} else if err == sql.ErrNoRows {
			// Item não existe, inserir novo
			query := "INSERT INTO cart_items (quantity, cart_id, products_id) VALUES (?, ?, ?)"
			result, err := db.Exec(query, newItem.Quantity, newItem.CartID, newItem.ProductsID)
			if err != nil {
				log.Println("Erro ao criar item do carrinho:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar item do carrinho"})
			}

			itemID, err := result.LastInsertId()
			if err != nil {
				log.Println("Erro ao obter ID do novo item:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID do novo item"})
			}

			newItem.ID = int(itemID)
			return c.Status(201).JSON(newItem)
		} else {
			log.Println("Erro ao verificar item existente:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar item existente"})
		}
	}
}

// GetCartItems retorna a lista de Itens do Carrinho
// @Summary Lista todos os Itens do Carrinho
// @Tags CartItems
// @Accept  json
// @Produce  json
// @Success 200 {array} CartItem
// @Router /cart/cart-items [get]
func GetCartItems(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, quantity, cart_id, products_id FROM cart_items")
		if err != nil {
			log.Println("Erro ao buscar itens do carrinho:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar itens do carrinho"})
		}
		defer rows.Close()

		var items []CartItem
		for rows.Next() {
			var item CartItem
			if err := rows.Scan(&item.ID, &item.Quantity, &item.CartID, &item.ProductsID); err != nil {
				log.Println("Erro ao ler dados do item:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Falha ao ler dados do item"})
			}
			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro ao processar itens:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar itens"})
		}

		return c.Status(200).JSON(items)
	}
}

// GetCartItemByID retorna um item baseado no ID
// @Summary Busca um item do carrinho pelo ID
// @Tags CartItems
// @Accept  json
// @Produce  json
// @Param id path int true "ID do item"
// @Success 200 {object} CartItem
// @Failure 404 {object} map[string]string "Item não encontrado"
// @Failure 500 {object} map[string]string "Falha ao buscar item"
// @Router /cart/cart-items/{id} [get]
func GetCartItemByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do item inválido"})
		}

		var item CartItem
		query := "SELECT id, quantity, cart_id, products_id FROM cart_items WHERE id = ?"
		err := db.QueryRow(query, id).Scan(&item.ID, &item.Quantity, &item.CartID, &item.ProductsID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "Item não encontrado"})
			}
			log.Println("Erro ao buscar item pelo ID:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao buscar item"})
		}

		return c.Status(200).JSON(item)
	}
}

// UpdateCartItem atualiza parcialmente um Item com base no ID
// @Summary Atualiza parcialmente um Item pelo ID
// @Tags CartItems
// @Accept  json
// @Produce  json
// @Param id path int true "ID do Item"
// @Param item body CartItem true "Dados do Item para atualização parcial"
// @Success 200 {object} CartItem "Item atualizado com sucesso"
// @Failure 400 {object} map[string]string "ID inválido ou dados de entrada inválidos"
// @Failure 404 {object} map[string]string "Item não encontrado"
// @Failure 500 {object} map[string]string "Erro ao atualizar item"
// @Router /cart/cart-items/{id} [patch]
func UpdateCartItem(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		itemID, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID inválido"})
		}

		// Struct temporária para pegar os dados de entrada
		var itemUpdates CartItem
		if err := c.BodyParser(&itemUpdates); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dados de entrada inválidos"})
		}

		// Verifica se o item existe
		var existingItem CartItem
		err = db.QueryRow("SELECT id, quantity, cart_id, products_id FROM cart_items WHERE id = ?", itemID).Scan(
			&existingItem.ID, &existingItem.Quantity, &existingItem.CartID, &existingItem.ProductsID)
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Item não encontrado"})
		} else if err != nil {
			log.Println("Erro ao buscar item:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar item"})
		}

		// Verificar estoque se a quantidade for alterada
		if itemUpdates.Quantity != 0 && itemUpdates.Quantity != existingItem.Quantity {
			var availableStock int
			stockQuery := "SELECT quantity FROM products WHERE id = ?"
			err := db.QueryRow(stockQuery, existingItem.ProductsID).Scan(&availableStock)
			if err != nil {
				log.Println("Erro ao verificar estoque:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar estoque do produto"})
			}

			if itemUpdates.Quantity > availableStock {
				return c.Status(400).JSON(fiber.Map{"error": "Quantidade solicitada excede o estoque disponível"})
			}
		}

		// Atualiza somente os campos que foram enviados no body
		if itemUpdates.Quantity != 0 {
			existingItem.Quantity = itemUpdates.Quantity
		}
		if itemUpdates.CartID != nil {
			existingItem.CartID = itemUpdates.CartID
		}
		if itemUpdates.ProductsID != nil {
			existingItem.ProductsID = itemUpdates.ProductsID
		}

		// Atualiza os dados no banco de dados
		_, err = db.Exec("UPDATE cart_items SET quantity = ?, cart_id = ?, products_id = ? WHERE id = ?",
			existingItem.Quantity, existingItem.CartID, existingItem.ProductsID, itemID)
		if err != nil {
			log.Println("Erro ao atualizar item:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao atualizar item"})
		}

		return c.Status(200).JSON(existingItem)
	}
}

// DeleteCartItemByID deleta um item baseado no ID
// @Summary Deleta um item do carrinho pelo ID
// @Tags CartItems
// @Accept  json
// @Produce  json
// @Param id path int true "ID do item"
// @Success 200 {object} map[string]string "Item deletado com sucesso"
// @Failure 404 {object} map[string]string "Item não encontrado"
// @Failure 500 {object} map[string]string "Falha ao deletar item"
// @Router /cart/cart-items/{id} [delete]
func DeleteCartItemByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o ID é um número válido
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ID do item inválido"})
		}

		query := "DELETE FROM cart_items WHERE id = ?"
		result, err := db.Exec(query, id)
		if err != nil {
			log.Println("Erro ao deletar item:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao deletar item"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Erro ao verificar linhas afetadas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Falha ao verificar linhas afetadas"})
		}

		if rowsAffected == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Item não encontrado"})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Item deletado com sucesso"})
	}
}
