// cmd/myapp/main.go
package main

import (
	"io"
	"log"
	"myapp/internal/handlers"
	customMiddleware "myapp/internal/middleware"
	"myapp/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"html/template"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// Serve static files
	e.Static("/static", "web/static")

	// Database connection
	// dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "db_user:TMyXiZHZLazri0[a@tcp(127.0.0.1:3306)/fgo_myapp?charset=utf8mb4&parseTime=True&loc=Local"
	log.Println("Connecting to database...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	log.Println("Connected to database successfully")

	// Migrate the schema
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("failed to migrate database:", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize user handler
	userHandler := handlers.UserHandler{DB: db}

	// Routes
	e.GET("/", func(c echo.Context) error {
		// return c.String(http.StatusOK, "Hello, World!")
		return c.Render(http.StatusOK, "index.html", nil)
	})
	e.POST("/users", userHandler.CreateUser)
	e.POST("/login", userHandler.LoginUser)

	// Protected routes
	r := e.Group("/restricted")
	r.Use(customMiddleware.JWTMiddleware())
	r.GET("/users/:id", userHandler.GetUser)
	r.PUT("/users/:id", userHandler.UpdateUser)
	r.GET("/users", userHandler.ListUsers)

	// Template renderer
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseFiles("web/templates/index.html")),
	}

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// TemplateRenderer is a custom HTML template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
