package main

import (
	"log"

	"bugforge-backend/internal/config"
	"bugforge-backend/internal/database"

	pg "bugforge-backend/internal/repository/postgres"

	"bugforge-backend/internal/service"

	"bugforge-backend/internal/http/controllers"
	mw "bugforge-backend/internal/http/middlewares"
	"bugforge-backend/internal/http/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "bugforge-backend/docs"

	ws "bugforge-backend/internal/websocket"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	// -----------------------
	// Repositories
	// -----------------------
	projectRepo := pg.NewProjectRepository(db)
	userRepo := pg.NewUserRepo(db)
	issueRepo := pg.NewIssueRepository(db)
	commentRepo := pg.NewCommentRepository(db)
	activityRepo := pg.NewActivityRepository(db)
	projectMemberRepo := pg.NewProjectMemberRepo(db)
	kanbanRepo := pg.NewKanbanRepoPG(db)
	labelRepo := pg.NewLabelRepository(db)

	// -----------------------
	// Services
	// -----------------------
	projectService := service.NewProjectService(projectRepo)
	userService := service.NewUserService(userRepo, projectRepo, projectMemberRepo)
	authService := service.NewAuthService(userRepo)
	issueService := service.NewIssueService(issueRepo, projectRepo, userRepo, commentRepo, activityRepo)
	projectMemberService := service.NewProjectMemberService(projectRepo, userRepo, projectMemberRepo)
	kanbanService := service.NewKanbanService(issueRepo, projectRepo, projectMemberRepo, kanbanRepo)
	labelService := service.NewLabelService(labelRepo, projectRepo)

	// -----------------------
	// Controllers
	// -----------------------
	projectController := controllers.NewProjectController(projectService)
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService, userService)

	// Split issue controllers
	issueController := controllers.NewIssueController(issueService)
	issueCommentController := controllers.NewIssueCommentController(issueService)
	issueRelationController := controllers.NewIssueRelationController(issueService)
	issueAttachmentController := controllers.NewIssueAttachmentController(issueService)
	issueChecklistController := controllers.NewIssueChecklistController(issueService)
	issueSubtaskController := controllers.NewIssueSubtaskController(issueService)

	projectMemberController := controllers.NewProjectMemberController(projectMemberService)
	projectLabelController := controllers.NewProjectLabelController(labelService)

	// -----------------------
	// Fiber app + WS Hub
	// -----------------------
	app := fiber.New()
	hub := ws.NewHub()

	// Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// API Root (CORS applied)
	api := app.Group("/api", cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
	}))

	// Public Auth Routes
	authPublic := api.Group("/auth")
	routes.AuthRoutes(authPublic, authController)

	// Protected Auth Routes
	authProtected := api.Group("/auth")
	authProtected.Use(mw.JWTProtected())
	routes.AuthProtectedRoutes(authProtected, authController)

	// Protected API Routes
	protected := api.Use(mw.JWTProtected())

	routes.ProjectRoutes(protected, projectController, issueController, projectMemberController, projectLabelController)

	// NEW ISSUE ROUTE WIRING
	routes.IssueRoutes(
		protected,
		issueController,
		issueCommentController,
		issueRelationController,
		issueAttachmentController,
		issueChecklistController,
		issueSubtaskController,
	)

	routes.UserRoutes(protected, userController)

	// Kanban
	routes.RegisterKanbanRoutes(protected, kanbanService, hub)

	// WebSocket Routes (no CORS)
	wsGroup := app.Group("/ws")
	routes.RegisterWebSocketRoutes(wsGroup, hub)

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Println("BugForge server running on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
