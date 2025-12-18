package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtKey []byte

// ============ Структуры данных ============

type User struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"-"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Birthdate  string `json:"birthdate"`
	IsMale     bool   `json:"is_male"` // Указатель, чтобы корректно обрабатывать NULL
	ProfileTag string `json:"profile_tag"`
	IsAdmin    bool   `json:"is_admin"`
}

type Claims struct {
	UserID  string `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

type Group struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	UserID string `json:"user_id"`
}

type Task struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Done    bool   `json:"done"`
	GroupID string `json:"group_id"`
}

type CartItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Price     int    `json:"price"`
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	ImageURL    string `json:"image_url"`
	IsActive    bool   `json:"is_active"`
}

type Review struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Username    string  `json:"username,omitempty"`
	ProductID   *string `json:"product_id"`
	Rating      int     `json:"rating"`
	Comment     string  `json:"comment"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	ModeratedAt string  `json:"moderated_at,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateGroupRequest struct {
	Title string `json:"title"`
}

type UpdateGroupRequest struct {
	Title string `json:"title"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

type AddToCartRequest struct {
	ProductID string `json:"product_id"`
}

type CreateReviewRequest struct {
	ProductID string `json:"product_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
}

type ProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	ImageURL    string `json:"image_url"`
	IsActive    bool   `json:"is_active"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ============ Инициализация ============

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtKey = []byte(jwtSecret)
}

func main() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	defer db.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	e.POST("/api/register", register)
	e.POST("/api/login", login)
	e.GET("/health", healthCheck)
	e.GET("/api/products", getProducts)
	e.GET("/api/reviews", getReviews)

	r := e.Group("/api")
	r.Use(authMiddleware)

	r.POST("/reviews", createReview)
	r.GET("/profile", getProfile)
	r.PUT("/profile", updateProfile)

	r.GET("/groups", getGroups)
	r.POST("/groups", createGroup)
	r.PUT("/groups/:id", updateGroup)
	r.DELETE("/groups/:id", deleteGroup)

	r.GET("/groups/:id/tasks", getTasksByGroup)
	r.POST("/groups/:id/tasks", createTask)
	r.PUT("/tasks/:id", updateTask)
	r.DELETE("/tasks/:id", deleteTask)

	r.GET("/cart", getCart)
	r.POST("/cart", addToCart)
	r.DELETE("/cart", clearCart)

	admin := e.Group("/api/admin")
	admin.Use(authMiddleware)
	admin.Use(adminMiddleware)

	admin.GET("/products", getAdminProducts)
	admin.POST("/products", createProduct)
	admin.PUT("/products/:id", updateProduct)
	admin.DELETE("/products/:id", deleteProduct)

	admin.GET("/reviews", getAdminReviews)
	admin.POST("/reviews/:id/approve", approveReview)
	admin.POST("/reviews/:id/reject", rejectReview)
	admin.DELETE("/reviews/:id", deleteReview)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// ============ Middleware ============

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing authorization header"})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization format"})
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
		}

		if claims.UserID == "" {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token claims"})
		}

		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin)
		return next(c)
	}
}

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin, ok := c.Get("is_admin").(bool)
		if !ok || !isAdmin {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "admin access required"})
		}
		return next(c)
	}
}

// ============ Авторизация ============

func register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if err := validateUsername(req.Username); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}
	if err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	profileTag := generateProfileTag()
	var userID string

	err = db.QueryRow(
		`INSERT INTO users (username, password, profile_tag, is_admin) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Username, string(hash), profileTag, false).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, ErrorResponse{Error: "username already exists"})
		}
		log.Printf("Registration error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":          userID,
		"username":    req.Username,
		"profile_tag": profileTag,
	})
}

func login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	var user User
	err := db.QueryRow(
		`SELECT id, password, is_admin FROM users WHERE username=$1`,
		req.Username).Scan(&user.ID, &user.Password, &user.IsAdmin)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	}
	if err != nil {
		log.Printf("Login error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	}

	token, err := createJWT(user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("JWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":    token,
		"user_id":  user.ID,
		"username": req.Username,
		"is_admin": user.IsAdmin,
	})
}

func createJWT(userID string, isAdmin bool) (string, error) {
	claims := &Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ============ Профили ============

func getProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var user User
	err := db.QueryRow(`
		SELECT id, username, 
			COALESCE(first_name, '') AS first_name,
			COALESCE(last_name, '') AS last_name,
			COALESCE(TO_CHAR(birthdate, 'YYYY-MM-DD'), '') AS birthdate,
			is_male,
			COALESCE(profile_tag, '') AS profile_tag,
			is_admin
		FROM users WHERE id=$1`,
		userID).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Birthdate, &user.IsMale, &user.ProfileTag, &user.IsAdmin)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile not found"})
	}
	if err != nil {
		log.Printf("Profile error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, user)
}

func updateProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Birthdate != "" {
		if _, err := time.Parse("2006-01-02", req.Birthdate); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid birthdate format, use YYYY-MM-DD"})
		}
	}

	var isMale *bool

	switch req.Gender {
	case "M":
		t := true
		isMale = &t
	case "F":
		f := false
		isMale = &f
	case "":
		isMale = nil
	default:
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "gender must be 'M' (Male) or 'F' (Female)"})
	}
	result, err := db.Exec(`
        UPDATE users 
        SET first_name=$1, 
            last_name=$2, 
            birthdate=$3, 
            is_male = COALESCE($4, is_male) 
        WHERE id=$5`,
		req.FirstName, req.LastName, req.Birthdate, isMale, userID)

	if err != nil {
		log.Printf("Update profile error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "profile updated successfully"})
}

// ============ Группы ============

func getGroups(c echo.Context) error {
	userID := c.Get("user_id").(string)

	rows, err := db.Query(
		`SELECT id, title FROM groups WHERE user_id=$1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		log.Printf("Get groups error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Title); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		g.UserID = userID
		groups = append(groups, g)
	}

	if groups == nil {
		groups = []Group{}
	}
	return c.JSON(http.StatusOK, groups)
}

func createGroup(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title cannot be empty"})
	}

	var groupID string
	err := db.QueryRow(
		`INSERT INTO groups (title, user_id) VALUES ($1, $2) RETURNING id`,
		req.Title, userID).Scan(&groupID)

	if err != nil {
		log.Printf("Create group error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    groupID,
		"title": req.Title,
	})
}

func updateGroup(c echo.Context) error {
	userID := c.Get("user_id").(string)
	groupID := c.Param("id")

	var req UpdateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	result, err := db.Exec(
		`UPDATE groups SET title=$1 WHERE id=$2 AND user_id=$3`,
		req.Title, groupID, userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "group not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "group updated"})
}

func deleteGroup(c echo.Context) error {
	userID := c.Get("user_id").(string)
	groupID := c.Param("id")

	result, err := db.Exec(`DELETE FROM groups WHERE id=$1 AND user_id=$2`, groupID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "group not found"})
	}

	return c.NoContent(http.StatusOK)
}

// ============ Таски ============

func getTasksByGroup(c echo.Context) error {
	groupID := c.Param("id")

	rows, err := db.Query(
		`SELECT id, title, done FROM tasks WHERE group_id=$1 ORDER BY created_at DESC`,
		groupID)
	if err != nil {
		log.Printf("Get tasks error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			continue
		}
		t.GroupID = groupID
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = []Task{}
	}
	return c.JSON(http.StatusOK, tasks)
}

func createTask(c echo.Context) error {
	groupID := c.Param("id")

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title cannot be empty"})
	}

	var taskID string
	err := db.QueryRow(
		`INSERT INTO tasks (title, group_id, done) VALUES ($1, $2, false) RETURNING id`,
		req.Title, groupID).Scan(&taskID)

	if err != nil {
		log.Printf("Create task error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":       taskID,
		"title":    req.Title,
		"done":     false,
		"group_id": groupID,
	})
}

func updateTask(c echo.Context) error {
	taskID := c.Param("id")

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Title == nil && req.Done == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "nothing to update"})
	}

	var query string
	var args []interface{}

	if req.Title != nil && req.Done != nil {
		query = `UPDATE tasks SET title=$1, done=$2 WHERE id=$3`
		args = []interface{}{*req.Title, *req.Done, taskID}
	} else if req.Title != nil {
		query = `UPDATE tasks SET title=$1 WHERE id=$2`
		args = []interface{}{*req.Title, taskID}
	} else {
		query = `UPDATE tasks SET done=$1 WHERE id=$2`
		args = []interface{}{*req.Done, taskID}
	}

	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Update task error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "task updated"})
}

func deleteTask(c echo.Context) error {
	taskID := c.Param("id")
	result, err := db.Exec(`DELETE FROM tasks WHERE id=$1`, taskID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
	}
	return c.NoContent(http.StatusOK)
}

// ============ Карты товаров ============

func getCart(c echo.Context) error {
	userID := c.Get("user_id").(string)

	rows, err := db.Query(`
		SELECT c.id, c.product_id, c.quantity, p.name, p.image_url, p.price 
		FROM cart_items c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id=$1`,
		userID)
	if err != nil {
		log.Printf("Get cart error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var cart []CartItem
	for rows.Next() {
		var item CartItem
		item.UserID = userID

		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Name, &item.Image, &item.Price); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		cart = append(cart, item)
	}

	if cart == nil {
		cart = []CartItem{}
	}
	return c.JSON(http.StatusOK, cart)
}

func addToCart(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}
	if req.ProductID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product id"})
	}

	var cartItemID string
	err := db.QueryRow(
		`INSERT INTO cart_items (user_id, product_id, quantity) 
		 VALUES ($1, $2, 1) RETURNING id`,
		userID, req.ProductID).Scan(&cartItemID)

	if err != nil {
		log.Printf("Add to cart error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": cartItemID,
	})
}

func clearCart(c echo.Context) error {
	userID := c.Get("user_id").(string)
	result, err := db.Exec(`DELETE FROM cart_items WHERE user_id=$1`, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "cart cleared",
		"deleted": rows,
	})
}

// ============ Продукты ============

func getProducts(c echo.Context) error {
	rows, err := db.Query(
		`SELECT id, name, description, price, image_url, is_active FROM products WHERE is_active = true ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("Get products error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive); err != nil {
			continue
		}
		products = append(products, p)
	}
	if products == nil {
		products = []Product{}
	}
	return c.JSON(http.StatusOK, products)
}

func getAdminProducts(c echo.Context) error {
	rows, err := db.Query(
		`SELECT id, name, description, price, image_url, is_active FROM products ORDER BY created_at DESC`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive); err != nil {
			continue
		}
		products = append(products, p)
	}
	if products == nil {
		products = []Product{}
	}
	return c.JSON(http.StatusOK, products)
}

func createProduct(c echo.Context) error {
	var req ProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "name cannot be empty"})
	}
	if req.Price < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "price must be >= 0"})
	}

	var productID string
	err := db.QueryRow(
		`INSERT INTO products (name, description, price, image_url, is_active) VALUES ($1, $2, $3, $4, true) RETURNING id`,
		req.Name, req.Description, req.Price, req.ImageURL).Scan(&productID)

	if err != nil {
		log.Printf("Create product error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": productID,
	})
}

func updateProduct(c echo.Context) error {
	productID := c.Param("id")
	var req ProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	result, err := db.Exec(
		`UPDATE products SET name=$1, description=$2, price=$3, image_url=$4, is_active=$5 WHERE id=$6`,
		req.Name, req.Description, req.Price, req.ImageURL, req.IsActive, productID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "product updated"})
}

func deleteProduct(c echo.Context) error {
	productID := c.Param("id")
	result, err := db.Exec(`DELETE FROM products WHERE id=$1`, productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
	}
	return c.NoContent(http.StatusOK)
}

// ============ Отзывы ============

func getReviews(c echo.Context) error {
	rows, err := db.Query(`
        SELECT r.id, r.user_id, u.username, r.product_id, r.rating, r.comment, r.status, r.created_at
        FROM reviews r
        JOIN users u ON r.user_id = u.id
        WHERE r.status = 'approved'
        ORDER BY r.created_at DESC
        LIMIT 50
    `)
	if err != nil {
		log.Printf("Get reviews error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Username, &rev.ProductID, &rev.Rating, &rev.Comment, &rev.Status, &rev.CreatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		reviews = append(reviews, rev)
	}
	if reviews == nil {
		reviews = []Review{}
	}
	return c.JSON(http.StatusOK, reviews)
}

func createReview(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Rating < 1 || req.Rating > 5 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "rating must be between 1 and 5"})
	}
	if req.Comment == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "comment cannot be empty"})
	}

	var productID *string
	if req.ProductID != "" {
		productID = &req.ProductID
	} else {
		productID = nil
	}

	var reviewID string
	err := db.QueryRow(
		`INSERT INTO reviews (user_id, product_id, rating, comment, status) 
         VALUES ($1, $2, $3, $4, 'pending') RETURNING id`,
		userID, productID, req.Rating, req.Comment).Scan(&reviewID)

	if err != nil {
		log.Printf("Create review error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": reviewID,
	})
}

func getAdminReviews(c echo.Context) error {
	status := c.QueryParam("status")
	if status == "" {
		status = "pending"
	}

	rows, err := db.Query(`
		SELECT r.id, r.user_id, u.username, r.product_id, r.rating, r.comment, r.status, r.created_at
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.status = $1
		ORDER BY r.created_at DESC
	`, status)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Username, &rev.ProductID, &rev.Rating, &rev.Comment, &rev.Status, &rev.CreatedAt); err != nil {
			continue
		}
		reviews = append(reviews, rev)
	}
	if reviews == nil {
		reviews = []Review{}
	}
	return c.JSON(http.StatusOK, reviews)
}

func approveReview(c echo.Context) error {
	adminID := c.Get("user_id").(string)
	reviewID := c.Param("id")

	result, err := db.Exec(
		`UPDATE reviews SET status='approved', moderated_by=$1, moderated_at=NOW() WHERE id=$2`,
		adminID, reviewID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "review approved"})
}

func rejectReview(c echo.Context) error {
	adminID := c.Get("user_id").(string)
	reviewID := c.Param("id")

	result, err := db.Exec(
		`UPDATE reviews SET status='rejected', moderated_by=$1, moderated_at=NOW() WHERE id=$2`,
		adminID, reviewID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}
	return c.NoContent(http.StatusOK)
}

func deleteReview(c echo.Context) error {
	reviewID := c.Param("id")
	result, err := db.Exec(`DELETE FROM reviews WHERE id=$1`, reviewID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}
	return c.NoContent(http.StatusOK)
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return fmt.Errorf("invalid characters in username")
		}
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return fmt.Errorf("password length must be between 8 and 128")
	}
	return nil
}

func generateProfileTag() string {
	return fmt.Sprintf("User%d", time.Now().UnixNano())
}

func healthCheck(c echo.Context) error {
	if err := db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database connection failed"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}
