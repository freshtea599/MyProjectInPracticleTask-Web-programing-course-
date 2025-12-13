package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

// Структуры данных
type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"-"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Birthdate  string `json:"birthdate"`
	Gender     string `json:"gender"`
	ProfileTag string `json:"profile_tag"`
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Group struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	UserID int    `json:"user_id"`
}

type Task struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Done    bool   `json:"done"`
	GroupID int    `json:"group_id"`
}

type CartItem struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	ProductID int    `json:"product_id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Price     int    `json:"price"`
}

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	ImageURL    string `json:"image_url"`
	IsActive    bool   `json:"is_active"`
}

type Review struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Username    string `json:"username,omitempty"` // для фронта (join)
	Rating      int    `json:"rating"`
	Comment     string `json:"comment"`
	Status      string `json:"status"` // 'pending', 'approved', 'rejected'
	CreatedAt   string `json:"created_at"`
	ModeratedAt string `json:"moderated_at,omitempty"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type AdminReviewAction struct {
	Action string `json:"action"` // 'approve' или 'reject'
}

// Ошибка валидации
type ErrorResponse struct {
	Error string `json:"error"`
}

func init() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Проверяем наличие обязательных переменных
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtKey = []byte(jwtSecret)
}

func main() {
	var err error

	// Подключение к БД из переменной окружения
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Проверка подключения к БД
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	defer db.Close()

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Публичные маршруты
	e.POST("/api/register", register)
	e.POST("/api/login", login)
	e.GET("/health", healthCheck)

	// Защищённые маршруты
	r := e.Group("/api")
	r.Use(authMiddleware)

	// ПУБЛИЧНЫЕ (неавторизованные)
	e.GET("/api/products", getProducts)
	e.GET("/api/reviews", getReviews)

	// ЗАЩИЩЁННЫЕ (требуют авторизацию)
	r.POST("/reviews", createReview)

	// АДМИНСКИЕ
	admin := e.Group("/api/admin")
	admin.Use(authMiddleware)
	admin.Use(adminMiddleware)

	// Товары админ
	admin.GET("/products", getAdminProducts)
	admin.POST("/products", createProduct)
	admin.PUT("/products/:id", updateProduct)
	admin.DELETE("/products/:id", deleteProduct)

	// Отзывы админ
	admin.GET("/reviews", getAdminReviews)
	admin.POST("/reviews/:id/approve", approveReview)
	admin.POST("/reviews/:id/reject", rejectReview)
	admin.DELETE("/reviews/:id", deleteReview)

	// Профиль
	r.GET("/profile", getProfile)
	r.PUT("/profile", updateProfile)

	// Группы
	r.GET("/groups", getGroups)
	r.POST("/groups", createGroup)
	r.PUT("/groups/:id", updateGroup)
	r.DELETE("/groups/:id", deleteGroup)

	// Задачи
	r.GET("/groups/:id/tasks", getTasksByGroup)
	r.POST("/groups/:id/tasks", createTask)
	r.PUT("/tasks/:id", updateTask)
	r.DELETE("/tasks/:id", deleteTask)

	// Корзина
	r.GET("/cart", getCart)
	r.POST("/cart", addToCart)
	r.DELETE("/cart", clearCart)

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

		// Извлекаем token из "Bearer <token>"
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization format"})
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// ВАЖНО: проверяем метод подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
		}

		// Проверяем, что UserID установлен
		if claims.UserID == 0 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token claims"})
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		return next(c)
	}
}

// ============ Middleware для админа============

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("role").(string)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "admin access required"})
		}
		return next(c)
	}
}

// ============ Регистрация ============

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	// Валидация
	if err := validateUsername(req.Username); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}
	if err := validatePassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	// Генерируем профиль-тег (заменяем функцию atoi)
	profileTag := generateProfileTag()

	// Вставляем пользователя (ID генерируется БД как SERIAL)
	var userID int
	err = db.QueryRow(
		`INSERT INTO users (username, password, profile_tag, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Username, string(hash), profileTag, "user").Scan(&userID)

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

// ============ Авторизация ============

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	var user User
	var role string
	err := db.QueryRow(
		`SELECT id, password, role FROM users WHERE username=$1`,
		req.Username).Scan(&user.ID, &user.Password, &role)

	if err == sql.ErrNoRows {
		// Одинаковая ошибка для защиты от перебора логинов
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	}
	if err != nil {
		log.Printf("Login query error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	// Сравнение хешей паролей
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
	}

	token, err := createJWT(user.ID, role)
	if err != nil {
		log.Printf("JWT creation error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":    token,
		"UserID":   user.ID,
		"Username": req.Username,
		"role":     role,
	})
}

func createJWT(userID int, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ============ Профиль ============

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
}

func getProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var user User
	err := db.QueryRow(`
		SELECT id, username, 
			COALESCE(first_name, '') AS first_name,
			COALESCE(last_name, '') AS last_name,
			COALESCE(TO_CHAR(birthdate, 'YYYY-MM-DD'), '') AS birthdate,
			COALESCE(gender, '') AS gender,
			COALESCE(profile_tag, '') AS profile_tag
		FROM users WHERE id=$1`,
		userID).Scan(
		&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Birthdate, &user.Gender, &user.ProfileTag)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile not found"})
	}
	if err != nil {
		log.Printf("Get profile error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusOK, user)
}

func updateProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	// Валидация даты рождения если задана
	if req.Birthdate != "" {
		if _, err := time.Parse("2006-01-02", req.Birthdate); err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid birthdate format, use YYYY-MM-DD"})
		}
	}

	// Валидация пола если задан
	if req.Gender != "" && req.Gender != "M" && req.Gender != "F" && req.Gender != "O" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid gender value"})
	}

	result, err := db.Exec(`
		UPDATE users 
		SET first_name=$1, last_name=$2, birthdate=$3, gender=$4 
		WHERE id=$5`,
		req.FirstName, req.LastName, req.Birthdate, req.Gender, userID)

	if err != nil {
		log.Printf("Update profile error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("RowsAffected error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "profile updated successfully"})
}

// ============ Группы ============

type CreateGroupRequest struct {
	Title string `json:"title"`
}

func getGroups(c echo.Context) error {
	userID := c.Get("user_id").(int)

	rows, err := db.Query(
		`SELECT id, title FROM groups WHERE user_id=$1 ORDER BY id DESC`,
		userID)
	if err != nil {
		log.Printf("Get groups query error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Title); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		g.UserID = userID
		groups = append(groups, g)
	}

	// ВАЖНО: проверяем ошибки после цикла
	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	// Возвращаем пустой массив вместо null
	if groups == nil {
		groups = []Group{}
	}

	return c.JSON(http.StatusOK, groups)
}

func createGroup(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title cannot be empty"})
	}

	if len(req.Title) > 255 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title too long"})
	}

	var groupID int
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

type UpdateGroupRequest struct {
	Title string `json:"title"`
}

func updateGroup(c echo.Context) error {
	userID := c.Get("user_id").(int)
	groupID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid group id"})
	}

	var req UpdateGroupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title cannot be empty"})
	}

	result, err := db.Exec(
		`UPDATE groups SET title=$1 WHERE id=$2 AND user_id=$3`,
		req.Title, groupID, userID)

	if err != nil {
		log.Printf("Update group error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "group not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "group updated successfully"})
}

func deleteGroup(c echo.Context) error {
	userID := c.Get("user_id").(int)
	groupID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid group id"})
	}

	result, err := db.Exec(
		`DELETE FROM groups WHERE id=$1 AND user_id=$2`,
		groupID, userID)

	if err != nil {
		log.Printf("Delete group error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "group not found"})
	}

	return c.NoContent(http.StatusOK)
}

// ============ Задачи ============

type CreateTaskRequest struct {
	Title string `json:"title"`
}

func getTasksByGroup(c echo.Context) error {
	groupID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid group id"})
	}

	rows, err := db.Query(
		`SELECT id, title, done FROM tasks WHERE group_id=$1 ORDER BY id DESC`,
		groupID)
	if err != nil {
		log.Printf("Get tasks query error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		t.GroupID = groupID
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return c.JSON(http.StatusOK, tasks)
}

func createTask(c echo.Context) error {
	groupID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid group id"})
	}

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title cannot be empty"})
	}

	var taskID int
	err = db.QueryRow(
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

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func updateTask(c echo.Context) error {
	taskID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid task id"})
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	// Проверяем, что хотя бы одно поле для обновления
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

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "task updated successfully"})
}

func deleteTask(c echo.Context) error {
	taskID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid task id"})
	}

	result, err := db.Exec(`DELETE FROM tasks WHERE id=$1`, taskID)
	if err != nil {
		log.Printf("Delete task error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "task not found"})
	}

	return c.NoContent(http.StatusOK)
}

// ============ Корзина ============

type AddToCartRequest struct {
	ProductID int    `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Price     int    `json:"price"`
}

func getCart(c echo.Context) error {
	userID := c.Get("user_id").(int)

	rows, err := db.Query(
		`SELECT id, product_id, name, image, price FROM cart_items WHERE user_id=$1`,
		userID)
	if err != nil {
		log.Printf("Get cart query error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var cart []CartItem
	for rows.Next() {
		var item CartItem
		item.UserID = userID
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Image, &item.Price); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		cart = append(cart, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if cart == nil {
		cart = []CartItem{}
	}

	return c.JSON(http.StatusOK, cart)
}

func addToCart(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var item AddToCartRequest
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	// Валидация
	if item.ProductID <= 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product id"})
	}
	if item.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "product name cannot be empty"})
	}
	if item.Price < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid price"})
	}

	var cartItemID int
	err := db.QueryRow(
		`INSERT INTO cart_items (user_id, product_id, name, image, price) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, item.ProductID, item.Name, item.Image, item.Price).Scan(&cartItemID)

	if err != nil {
		log.Printf("Add to cart error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": cartItemID,
	})
}

func clearCart(c echo.Context) error {
	userID := c.Get("user_id").(int)

	result, err := db.Exec(`DELETE FROM cart_items WHERE user_id=$1`, userID)
	if err != nil {
		log.Printf("Clear cart error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, _ := result.RowsAffected()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "cart cleared",
		"deleted": rows,
	})
}

// ============ Вспомогательные функции ============

// safeAtoi безопасно конвертирует строку в int с обработкой ошибок
func safeAtoi(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value")
	}
	return n, nil
}

// validateUsername проверяет корректность имени пользователя
func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must not exceed 50 characters")
	}
	// Проверяем только буквы, цифры и подчеркивание
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return fmt.Errorf("username can only contain letters, numbers, dashes and underscores")
		}
	}
	return nil
}

// validatePassword проверяет требования к паролю
func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}
	return nil
}

// generateProfileTag генерирует уникальный профиль-тег
func generateProfileTag() string {
	return fmt.Sprintf("User%d", time.Now().UnixNano())
}

// healthCheck простая проверка здоровья сервера
func healthCheck(c echo.Context) error {
	if err := db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database connection failed"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
}

// ============ ТОВАРЫ (PUBLIC) ============

func getProducts(c echo.Context) error {
	rows, err := db.Query(
		`SELECT id, name, description, price, image_url, is_active FROM products WHERE is_active = true ORDER BY id DESC`)
	if err != nil {
		log.Printf("Get products error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if products == nil {
		products = []Product{}
	}

	return c.JSON(http.StatusOK, products)
}

// ============ ОТЗЫВЫ (PUBLIC) ============

func getReviews(c echo.Context) error {
	rows, err := db.Query(`
		SELECT r.id, r.user_id, u.username, r.rating, r.comment, r.status, r.created_at
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
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Username, &rev.Rating, &rev.Comment, &rev.Status, &rev.CreatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		reviews = append(reviews, rev)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if reviews == nil {
		reviews = []Review{}
	}

	return c.JSON(http.StatusOK, reviews)
}

func createReview(c echo.Context) error {
	userID := c.Get("user_id").(int)

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

	var reviewID int
	err := db.QueryRow(
		`INSERT INTO reviews (user_id, rating, comment, status) VALUES ($1, $2, $3, 'pending') RETURNING id`,
		userID, req.Rating, req.Comment).Scan(&reviewID)

	if err != nil {
		log.Printf("Create review error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id": reviewID,
	})
}

// ============ АДМИН: ТОВАРЫ ============

func getAdminProducts(c echo.Context) error {
	rows, err := db.Query(
		`SELECT id, name, description, price, image_url, is_active FROM products ORDER BY id DESC`)
	if err != nil {
		log.Printf("Get admin products error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if products == nil {
		products = []Product{}
	}

	return c.JSON(http.StatusOK, products)
}

func createProduct(c echo.Context) error {
	var req Product
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "name cannot be empty"})
	}

	if req.Price < 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "price must be >= 0"})
	}

	var productID int
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
	productID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product id"})
	}

	var req Product
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
	}

	result, err := db.Exec(
		`UPDATE products SET name=$1, description=$2, price=$3, image_url=$4, is_active=$5 WHERE id=$6`,
		req.Name, req.Description, req.Price, req.ImageURL, req.IsActive, productID)

	if err != nil {
		log.Printf("Update product error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "product updated"})
}

func deleteProduct(c echo.Context) error {
	productID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product id"})
	}

	result, err := db.Exec(`DELETE FROM products WHERE id=$1`, productID)

	if err != nil {
		log.Printf("Delete product error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
	}

	return c.NoContent(http.StatusOK)
}

// ============ АДМИН: ОТЗЫВЫ ============

func getAdminReviews(c echo.Context) error {
	status := c.QueryParam("status")
	if status == "" {
		status = "pending"
	}

	rows, err := db.Query(`
		SELECT r.id, r.user_id, u.username, r.rating, r.comment, r.status, r.created_at
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.status = $1
		ORDER BY r.created_at DESC
	`, status)

	if err != nil {
		log.Printf("Get admin reviews error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.UserID, &rev.Username, &rev.Rating, &rev.Comment, &rev.Status, &rev.CreatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		reviews = append(reviews, rev)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if reviews == nil {
		reviews = []Review{}
	}

	return c.JSON(http.StatusOK, reviews)
}

func approveReview(c echo.Context) error {
	adminID := c.Get("user_id").(int)
	reviewID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
	}

	result, err := db.Exec(
		`UPDATE reviews SET status='approved', moderated_by=$1, moderated_at=NOW() WHERE id=$2`,
		adminID, reviewID)

	if err != nil {
		log.Printf("Approve review error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "review approved"})
}

func rejectReview(c echo.Context) error {
	adminID := c.Get("user_id").(int)
	reviewID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
	}

	result, err := db.Exec(
		`UPDATE reviews SET status='rejected', moderated_by=$1, moderated_at=NOW() WHERE id=$2`,
		adminID, reviewID)

	if err != nil {
		log.Printf("Reject review error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}

	return c.NoContent(http.StatusOK)
}
func deleteReview(c echo.Context) error {
	reviewID, err := safeAtoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid review id"})
	}

	result, err := db.Exec(`DELETE FROM reviews WHERE id=$1`, reviewID)
	if err != nil {
		log.Printf("Delete review error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	if rows == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "review not found"})
	}

	return c.NoContent(http.StatusOK)
}
