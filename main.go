package main

import (
	"log"
	"os"

	"cybersim/controllers"
	"cybersim/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/pgvector/pgvector-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=admin password=password dbname=cyber_sim port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	log.Println("Successfully connected to the database!")

	// Initialize Gin
	r := gin.Default()

	// Setup CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost"}, // Vite dev server and docker frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Controllers
	authController := controllers.NewAuthController(db)
	examController := controllers.NewExamController(db)

	// Routes
	api := r.Group("/api")
	{
		api.POST("/register", authController.Register)
		api.POST("/login", authController.Login)

		// Protected exam routes
		exams := api.Group("/exams", middlewares.AuthMiddleware())
		{
			exams.POST("/pre-test", examController.StartPreTest)
			exams.POST("/post-test", examController.StartPostTest)
			exams.GET("/session", examController.GetActiveSession)
			exams.POST("/submit-answer", examController.SubmitAnswer)
		}
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
