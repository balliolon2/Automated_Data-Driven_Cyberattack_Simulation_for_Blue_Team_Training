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
	"embed"
	"errors"
	"log"
	"os"

	"cybersim/controllers"
	_ "cybersim/docs"
	"cybersim/middlewares"
	"cybersim/models"
	"cybersim/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/pgvector/pgvector-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// scenarioSeedJSON embeds the static Scenario seed so the startup evidence
// remediation works in every deployment mode, including the runtime image that
// copies only the compiled binary.
//
//go:embed scenarios_seed.json
var scenarioSeedJSON []byte

//go:embed theory
var theoryMarkdownFS embed.FS

func main() {
	services.RegisterEmbeddedTheoryFS(theoryMarkdownFS)
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
		if errors.Is(err, models.ErrSelectionEnumUnavailable) {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Printf("Migration warning: %v", err)
	}

	// Archive legacy False Positive Scenarios that do not satisfy the evidence
	// contract. Seed-authored static Scenarios and completed Simulation Sessions
	// are never modified.
	if err := services.EnsureScenarioEvidenceQualityFromSeed(db, scenarioSeedJSON); err != nil {
		log.Fatalf("Scenario evidence quality gate failed: %v", err)
	}

	// Initialize Gin
	r := gin.Default()

	// Setup CORS
	allowedOrigins := []string{"http://localhost:5173", "http://localhost"}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	r.Use(middlewares.TelemetryMiddleware(services.GlobalNFRTelemetry()))

	// Controllers
	authController := controllers.NewAuthController(db)
	examController := controllers.NewExamController(db)
	learnController := controllers.NewLearnController(db)
	simController := controllers.NewSimulationController(db)
	specialistController := controllers.NewSpecialistController(db)
	adminController := controllers.NewAdminController(db)
	threadController := controllers.NewThreadController(db)
	commentController := controllers.NewCommentController(db)
	upvoteController := controllers.NewUpvoteController(db)
	notificationController := controllers.NewNotificationController(db)
	evaluationController := controllers.NewEvaluationController(db)
	nfrController := controllers.NewNFRController(db, r)

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
			userGroup.PUT("/language", authController.UpdateLanguage)
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

			// Thread Upvote
			threads.POST("/:id/upvote", upvoteController.ToggleThreadUpvote)

			// Thread Comments (Create requires auth)
			threads.POST("/:id/comments", commentController.CreateComment)
		}

		// Thread Comments (Public read with optional auth)
		api.GET("/threads/:id/comments", middlewares.OptionalAuthMiddleware(), commentController.ListComments)

		// Comment moderation & actions (Authenticated)
		comments := api.Group("/comments", middlewares.AuthMiddleware())
		{
			comments.PUT("/:id", commentController.UpdateComment)
			comments.DELETE("/:id", commentController.DeleteComment)
			comments.POST("/:id/pin", commentController.TogglePinComment)
			comments.POST("/:id/hide", commentController.ToggleHideComment)
			comments.POST("/:id/upvote", upvoteController.ToggleCommentUpvote)
		}

		// In-App Notifications (Authenticated)
		notifications := api.Group("/notifications", middlewares.AuthMiddleware())
		{
			notifications.GET("", notificationController.ListNotifications)
			notifications.PUT("/:id/read", notificationController.MarkAsRead)
			notifications.PUT("/read-all", notificationController.MarkAllAsRead)
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
			admin.GET("/evaluations", evaluationController.GetAdminEvaluations)
			admin.GET("/evaluations/statistics", evaluationController.GetAdminStatistics)
			admin.GET("/evaluations/export", evaluationController.ExportAdminEvaluations)
			admin.GET("/nfr/live", nfrController.GetLiveNFR)
			admin.POST("/nfr/telemetry/reset", nfrController.ResetTelemetryWindow)
			admin.POST("/nfr/runs", nfrController.CreateNFRRun)
			admin.GET("/nfr/runs", nfrController.ListNFRRuns)
			admin.GET("/nfr/runs/:runId", nfrController.GetNFRRun)
			admin.DELETE("/nfr/runs/:runId", nfrController.DeleteNFRRun)
			admin.POST("/nfr/probe/scalability", nfrController.RunScalabilityProbe)
			admin.GET("/nfr/export", nfrController.ExportNFRReport)
		}

		// Protected exam routes
		exams := api.Group("/exams", middlewares.AuthMiddleware())
		{
			exams.POST("/pre-test", examController.StartPreTest)
			exams.POST("/post-test", examController.StartPostTest)
			exams.GET("/session", examController.GetActiveSession)
			exams.POST("/submit-answer", examController.SubmitAnswer)
			exams.POST("/sessions/:sessionId/finalize", examController.FinalizeSession)
			exams.GET("/history", examController.GetExamHistory)
			exams.GET("/sessions/:sessionId/review", examController.GetExamReview)
		}

		// Protected Learning Theory Hub routes
		learn := api.Group("/learn", middlewares.AuthMiddleware())
		{
			learn.GET("/domains", learnController.GetDomains)
			learn.POST("/domains/:domainId/complete", learnController.CompleteDomain)
			learn.GET("/subtopics/:subTopicId", learnController.GetSubTopic)
			learn.POST("/subtopics/:subTopicId/sample-questions/:questionId/attempt", learnController.AttemptSampleQuestion)
			learn.POST("/subtopics/:subTopicId/complete", learnController.CompleteSubTopic)
		}

		// Protected simulation routes
		simulation := api.Group("/simulation", middlewares.AuthMiddleware())
		{
			simulation.GET("/status", simController.GetStatus)
			simulation.POST("/start", simController.StartScenario)
			simulation.GET("/session", simController.GetActiveSession)
			simulation.POST("/log-query", simController.LogQuery)
			simulation.POST("/select-log", simController.SelectLog)
			simulation.POST("/select-log/clear", simController.ClearLogSelection)
			simulation.POST("/submit", simController.SubmitScenario)
			simulation.GET("/result/:sessionId", simController.GetResult)
		}

		// Protected analytics / research routes
		analytics := api.Group("/analytics", middlewares.AuthMiddleware())
		{
			analytics.GET("/research-summary", simController.GetResearchSummary)
		}

		// Protected post-test evaluation questionnaire routes
		evaluation := api.Group("/evaluation", middlewares.AuthMiddleware())
		{
			evaluation.GET("/questionnaire/status", evaluationController.GetQuestionnaireStatus)
			evaluation.POST("/questionnaire", evaluationController.SubmitQuestionnaire)
		}

		// Health Check Endpoint
		api.GET("/health", nfrController.HealthCheck)

		// Swagger Documentation
		if os.Getenv("ENABLE_SWAGGER") == "true" {
			api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
			log.Println("Swagger documentation enabled at /api/docs/index.html")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
