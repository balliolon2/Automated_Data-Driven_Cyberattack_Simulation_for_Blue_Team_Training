package main

// @title SOC Trainer REST API
// @version 1.0
// @description Automated Data-Driven Cyberattack Simulation for Blue Team Training platform REST API.
// @termsOfService https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training

// @contact.name SOC Trainer Development Team
// @license.name MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOi..."

//go:generate swag init -g main.go --parseDependency --parseInternal

import (
	"log"
	"os"

	"cybersim/controllers"
	_ "cybersim/docs"
	"cybersim/middlewares"
	"cybersim/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/pgvector/pgvector-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Run migrations for Spec 1 (nickname, specialist role, specialist_applications table)
	if err := models.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

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
	simController := controllers.NewSimulationController(db)
	specialistController := controllers.NewSpecialistController(db)
	adminController := controllers.NewAdminController(db)
	threadController := controllers.NewThreadController(db)

	// Routes
	api := r.Group("/api")
	{
		api.POST("/register", authController.Register)
		api.POST("/login", authController.Login)
		api.GET("/exams/sample", examController.GetSampleQuestion)

		// Protected user profile routes
		userGroup := api.Group("/user", middlewares.AuthMiddleware())
		{
			userGroup.GET("/profile", authController.GetProfile)
			userGroup.PUT("/nickname", authController.UpdateNickname)
		}

		// Protected specialist application routes
		specialist := api.Group("/specialist", middlewares.AuthMiddleware())
		{
			specialist.POST("/apply", specialistController.Apply)
			specialist.GET("/application-status", specialistController.GetApplicationStatus)
		}

		// Specialist Review Console (Specialist and Admin only)
		specialistReviews := api.Group("/specialist", middlewares.AuthMiddleware(), middlewares.RequireRoles("specialist", "admin"))
		{
			specialistReviews.GET("/submissions", specialistController.GetSubmissions)
			specialistReviews.GET("/submissions/:session_id", specialistController.GetSubmissionDetail)
		}

		// Analysis Threads (Authenticated: Any role can read, Specialist/Admin can create)
		threads := api.Group("/threads", middlewares.AuthMiddleware())
		{
			threads.GET("", threadController.ListThreads)
			threads.GET("/:id", threadController.GetThread)
			threads.POST("", middlewares.RequireRoles("specialist", "admin"), threadController.CreateThread)
			threads.PUT("/:id", threadController.UpdateThread)
			threads.DELETE("/:id", threadController.DeleteThread)
		}

		// Application document streaming (accessible by Admin or applicant owner)
		api.GET("/admin/applications/:id/files/:file_type", middlewares.AuthMiddleware(), adminController.StreamDocument)
		api.GET("/applications/:id/files/:file_type", middlewares.AuthMiddleware(), adminController.StreamDocument)

		// Protected admin routes (Strictly Admin only)
		admin := api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RequireAdmin())
		{
			admin.GET("/applications", adminController.ListApplications)
			admin.POST("/applications/:id/approve", adminController.Approve)
			admin.POST("/applications/:id/reject", adminController.Reject)
		}

		// Protected exam routes
		exams := api.Group("/exams", middlewares.AuthMiddleware())
		{
			exams.POST("/pre-test", examController.StartPreTest)
			exams.POST("/post-test", examController.StartPostTest)
			exams.GET("/session", examController.GetActiveSession)
			exams.POST("/submit-answer", examController.SubmitAnswer)
		}

		// Protected simulation routes
		simulation := api.Group("/simulation", middlewares.AuthMiddleware())
		{
			simulation.GET("/status", simController.GetStatus)
			simulation.POST("/start", simController.StartScenario)
			simulation.GET("/session", simController.GetActiveSession)
			simulation.POST("/log-query", simController.LogQuery)
			simulation.POST("/submit", simController.SubmitScenario)
			simulation.GET("/result/:sessionId", simController.GetResult)
		}

		// Protected analytics / research routes
		analytics := api.Group("/analytics", middlewares.AuthMiddleware())
		{
			analytics.GET("/research-summary", simController.GetResearchSummary)
		}

		// Swagger Documentation
		if os.Getenv("ENABLE_SWAGGER") != "false" {
			api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
			log.Println("Swagger documentation enabled at /api/docs/index.html")
		}
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
