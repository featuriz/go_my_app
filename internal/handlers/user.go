// internal/handlers/user.go
package handlers

import (
	"log"
	"myapp/internal/models"
	"net/http"

	"myapp/internal/utils"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

// Create a new user
func (h *UserHandler) CreateUser(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}
	user.Password = string(hashedPassword)

	if err := h.DB.Create(user).Error; err != nil {
		log.Println("Error creating user:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, user)
}

// User login
func (h *UserHandler) LoginUser(c echo.Context) error {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&loginData); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	log.Println("Login attempt with email:", loginData.Email)

	var user models.User
	if err := h.DB.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		log.Println("Email not found:", loginData.Email)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	log.Println("Stored hashed password:", user.Password)
	log.Println("Input password:", loginData.Password)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		log.Println("Password mismatch for email:", loginData.Email)
		log.Println("Error comparing passwords:", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.Email)
	if err != nil {
		log.Println("Error generating token:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate token"})
	}

	log.Println("Login successful for email:", loginData.Email)

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

// Get user by ID
func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	user := new(models.User)

	if err := h.DB.First(user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch user"})
	}

	return c.JSON(http.StatusOK, user)
}

// Update user
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to hash password"})
		}
		user.Password = string(hashedPassword)
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update user"})
	}

	return c.JSON(http.StatusOK, user)
}

// List all users
func (h *UserHandler) ListUsers(c echo.Context) error {
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve users"})
	}
	return c.JSON(http.StatusOK, users)
}
