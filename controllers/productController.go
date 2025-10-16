package controllers

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
)

// Define um struct para o produto

type ProductBySKU struct {
	ID         int     `json:"id"`
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	UsersId    int     `json:"users_id"`
	Quantity   int     `json:"quantity"`
	CategoryId int     `json:"categories_product_id"`
}

type Product struct {
	ID           int     `json:"id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	UsersId      int     `json:"users_id"`
	Quantity     int     `json:"quantity"`
	CategoryName string  `json:"category_name"`
}

type ProductHome struct {
	ID        int     `json:"id"`
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	ImagePath string  `json:"image_path"`
}

type ProductCreate struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	Price      string  `json:"price"`
	UsersId    int     `json:"users_id"`
	Quantity   string  `json:"quantity"`
	CategoryId int     `json:"categories_product_id"`
}

type ProductCreateRaw struct {
	SKU        string      `json:"sku"`
	Name       string      `json:"name"`
	Price      string 	   `json:"price"`
	UsersId    int         `json:"users_id"`
	Quantity   string       `json:"quantity"`
	CategoryId int         `json:"categories_product_id"`
}


// @Summary Obter produto por ID
// @Description Obtém um produto com base no ID
// @Tags Products
// @Param id path int true "ID do produto"
// @Success 200 {object} Product
// @Failure 404 {object} map[string]string "Produto não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar produto"
// @Router /products/id/{id} [get]
func GetProductByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		productQuery := `
			SELECT 
				p.id, 
				p.sku, 
				p.name, 
				p.price, 
				p.quantity, 
				cp.name AS category_name 
			FROM products p
			INNER JOIN categories_products cp 
				ON p.categories_products_id = cp.id
			WHERE p.id = ?
		`

		row := db.QueryRow(productQuery, id)
		var product Product
		if err := row.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.CategoryName); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"message": "Produto não encontrado"})
			}
			log.Println("Erro ao buscar produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produto"})
		}

		return c.Status(200).JSON(product)
	}
}

// @Summary Obter produtos por nome da categoria com paginação
// @Description Obtém todos os produtos de uma categoria específica pelo nome da categoria com paginação
// @Tags Products
// @Param category_name path string true "Nome da Categoria"
// @Param page query int false "Número da página" default(1)
// @Param limit query int false "Limite de itens por página" default(10)
// @Success 200 {array} Product
// @Failure 500 {object} map[string]string "Erro ao buscar produtos"
// @Router /products/category/{category_name} [get]
func GetProductsByCategoryName(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		categoryName := c.Params("category_name")
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		offset := (page - 1) * limit

		productsQuery := `
            SELECT 
                p.id, 
                p.sku, 
                p.name, 
                p.price, 
                p.quantity, 
                cp.name AS category_name 
            FROM products p
            INNER JOIN categories_products cp 
                ON p.categories_products_id = cp.id
            WHERE cp.name = ?
            LIMIT ? OFFSET ?
        `

		rows, err := db.Query(productsQuery, categoryName, limit, offset)
		if err != nil {
			log.Println("Erro ao buscar produtos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produtos"})
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.CategoryName); err != nil {
				log.Println("Erro ao escanear produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler produto"})
			}
			products = append(products, product)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
		}

		return c.Status(200).JSON(products)
	}
}

// @Summary Obter produtos por ID da categoria com paginação
// @Description Obtém todos os produtos de uma categoria específica pelo ID da categoria com paginação
// @Tags Products
// @Param category_id path int true "ID da Categoria"
// @Param page query int false "Número da página" default(1)
// @Param limit query int false "Limite de itens por página" default(10)
// @Success 200 {array} Product
// @Failure 500 {object} map[string]string "Erro ao buscar produtos"
// @Router /products/category/id/{category_id} [get]
func GetProductsByCategoryID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		categoryID := c.Params("category_id")
		page := c.QueryInt("page", 1)
		limit := c.QueryInt("limit", 10)
		offset := (page - 1) * limit

		productsQuery := `
            SELECT 
                p.id, 
                p.sku, 
                p.name, 
                p.price, 
                p.quantity, 
                cp.name AS category_name 
            FROM products p
            INNER JOIN categories_products cp 
                ON p.categories_products_id = cp.id
            WHERE cp.id = ?
            LIMIT ? OFFSET ?
        `

		rows, err := db.Query(productsQuery, categoryID, limit, offset)
		if err != nil {
			log.Println("Erro ao buscar produtos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produtos"})
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.CategoryName); err != nil {
				log.Println("Erro ao escanear produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler produto"})
			}
			products = append(products, product)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
		}

		return c.Status(200).JSON(products)
	}
}

// @Summary Obter todos os produtos em destaque
// @Description Obtém todos os produtos com imagens em destaque
// @Tags Products
// @Success 200 {array} Product
// @Router /products/home [get]
func GetAllProductsHome(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productsQuery := `
			SELECT p.id, p.sku, p.name, p.price, p.quantity, i.path AS imagePath
			FROM products p
			LEFT JOIN images i ON p.id = i.products_id
			WHERE i.type = 'featured_image'
		`

		rows, err := db.Query(productsQuery)
		if err != nil {
			log.Println("Erro ao buscar produtos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produtos"})
		}
		defer rows.Close()

		var products []ProductHome
		for rows.Next() {
			var product ProductHome
			if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.ImagePath); err != nil {
				log.Println("Erro ao escanear produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler produto"})
			}
			products = append(products, product)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
		}

		return c.Status(200).JSON(products)
	}
}

// @Summary Obter todos os produtos
// @Description Obtém todos os produtos
// @Tags Products
// @Success 200 {array} Product
// @Router /products [get]
func GetAllProducts(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		productsQuery := `
			SELECT 
				p.id, 
				p.sku, 
				p.name, 
				p.price, 
				p.quantity, 
				cp.name AS category_name 
			FROM products p
			INNER JOIN categories_products cp 
				ON p.categories_products_id = cp.id
		`

		rows, err := db.Query(productsQuery)
		if err != nil {
			log.Println("Erro ao buscar produtos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produtos"})
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.CategoryName); err != nil {
				log.Println("Erro ao escanear produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler produto"})
			}
			products = append(products, product)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
		}

		return c.Status(200).JSON(products)
	}
}

// @Summary Obter todos os produtos por ID do usuário
// @Description Obtém todos os produtos associados a um usuário específico
// @Tags Products
// @Param user_id path int true "ID do Usuário"
// @Success 200 {array} Product
// @Failure 500 {object} map[string]string "Erro ao buscar produtos"
// @Router /products/user/{user_id} [get]
func GetAllProductsByUserID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")

		productsQuery := `
			SELECT 
				p.id, 
				p.sku, 
				p.name, 
				p.price, 
				p.quantity, 
				cp.name AS category_name 
			FROM products p
			INNER JOIN categories_products cp 
				ON p.categories_products_id = cp.id
			WHERE p.users_id = ?
		`

		rows, err := db.Query(productsQuery, userID)
		if err != nil {
			log.Println("Erro ao buscar produtos:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produtos"})
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.Quantity, &product.CategoryName); err != nil {
				log.Println("Erro ao escanear produto:", err)
				return c.Status(500).JSON(fiber.Map{"error": "Erro ao ler produto"})
			}
			products = append(products, product)
		}

		if err := rows.Err(); err != nil {
			log.Println("Erro com as linhas:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar produtos"})
		}

		return c.Status(200).JSON(products)
	}
}

// @Summary Obter produto por SKU
// @Description Obtém um produto com base no SKU
// @Tags Products
// @Param sku path string true "SKU do produto"
// @Success 200 {object} Product
// @Failure 404 {object} map[string]string "Produto não encontrado"
// @Failure 500 {object} map[string]string "Erro ao buscar produto"
// @Router /products/{sku} [get]
func GetProductBySKU(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sku := c.Params("sku")

		row := db.QueryRow("SELECT * FROM products WHERE sku = ?", sku)
		var product ProductBySKU
		if err := row.Scan(&product.ID, &product.SKU, &product.Name, &product.Price, &product.UsersId, &product.Quantity, &product.CategoryId); err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"message": "Produto não encontrado"})
			}
			log.Println("Erro ao ler produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar produto"})
		}

		return c.Status(200).JSON(product)
	}
}

// @Summary Criar um novo produto
// @Description Cria um novo produto no banco de dados
// @Tags Products
// @Accept json
// @Produce json
// @Param product body ProductCreate true "Dados do produto"
// @Success 200 {object} map[string]interface{} "Produto criado com sucesso"
// @Failure 400 {object} map[string]string "Dados inválidos"
// @Failure 500 {object} map[string]string "Erro ao criar produto"
// @Router /products [post]
func CreateProduct(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var productRaw ProductCreateRaw
		
		// Parse do body da requisição
		if err := c.BodyParser(&productRaw); err != nil {
			log.Println("Erro ao fazer parse do body:", err)
			return c.Status(400).JSON(fiber.Map{"error": "Dados inválidos"})
		}

	
		// Monta o produto final
		product := ProductCreate{
			SKU:        productRaw.SKU,
			Name:       productRaw.Name,
			Price:      productRaw.Price,
			UsersId:    productRaw.UsersId,
			Quantity:   productRaw.Quantity,
			CategoryId: productRaw.CategoryId,
		}

		// Validações básicas
		if product.SKU == "" || product.Name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "SKU e Nome são obrigatórios"})
		}

		if product.Price == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Preço não pode ser vazio"})
		}

		if product.Quantity == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Quantidade não pode ser negativa"})
		}

		// Verifica se o SKU já existe
		var exists int
		checkQuery := "SELECT COUNT(*) FROM products WHERE sku = ?"
		err := db.QueryRow(checkQuery, product.SKU).Scan(&exists)
		if err != nil {
			log.Println("Erro ao verificar SKU:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar produto"})
		}

		if exists > 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Produto com este SKU já existe"})
		}

		// Insere o produto no banco de dados
		insertQuery := `
			INSERT INTO products (sku, name, price, users_id, quantity, categories_products_id)
			VALUES (?, ?, ?, ?, ?, ?)
		`

		result, err := db.Exec(insertQuery, 
			product.SKU, 
			product.Name, 
			product.Price, 
			product.UsersId, 
			product.Quantity, 
			product.CategoryId,
		)

		if err != nil {
			log.Println("Erro ao inserir produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao criar produto"})
		}

		// Obtém o ID do produto inserido
		productID, err := result.LastInsertId()
		if err != nil {
			log.Println("Erro ao obter ID do produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao obter ID do produto"})
		}

		// Retorna o produto criado com sucesso
		return c.Status(200).JSON(fiber.Map{
			"message": "Produto criado com sucesso",
			"product": fiber.Map{
				"id":                      productID,
				"sku":                     product.SKU,
				"name":                    product.Name,
				"price":                   product.Price,
				"users_id":                product.UsersId,
				"quantity":                product.Quantity,
				"categories_product_id":   product.CategoryId,
			},
		})
	}
}

// @Summary Excluir produto por ID
// @Description Exclui um produto com base no ID
// @Tags Products
// @Param id path int true "ID do produto"
// @Success 200 {object} map[string]string "Produto excluído com sucesso"
// @Failure 404 {object} map[string]string "Produto não encontrado"
// @Failure 500 {object} map[string]string "Erro ao excluir produto"
// @Router /products/id/{id} [delete]
func DeleteProductByID(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// Verifica se o produto existe
		var exists int
		checkQuery := "SELECT COUNT(*) FROM products WHERE id = ?"
		err := db.QueryRow(checkQuery, id).Scan(&exists)
		if err != nil {
			log.Println("Erro ao verificar produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao verificar produto"})
		}

		if exists == 0 {
			return c.Status(404).JSON(fiber.Map{"message": "Produto não encontrado"})
		}

		// Exclui o produto
		deleteQuery := "DELETE FROM products WHERE id = ?"
		_, err = db.Exec(deleteQuery, id)
		if err != nil {
			log.Println("Erro ao excluir produto:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao excluir produto"})
		}

		return c.Status(200).JSON(fiber.Map{
			"message": "Produto excluído com sucesso",
			"id":      id,
		})
	}
}