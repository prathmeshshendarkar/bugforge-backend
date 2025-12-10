package main

import (
	"log"

	"bugforge-backend/internal/config"
	"bugforge-backend/internal/database"
	"bugforge-backend/internal/events/handlers"

	ws "bugforge-backend/internal/websocket"
	commentws "bugforge-backend/internal/websocket/comments"
	"bugforge-backend/internal/websocket/notifications"

	pg "bugforge-backend/internal/repository/postgres"
	kanbanrepo "bugforge-backend/internal/repository/postgres/KanbanPostgres"

	"bugforge-backend/internal/service"

	"bugforge-backend/internal/http/controllers"
	mw "bugforge-backend/internal/http/middlewares"
	"bugforge-backend/internal/http/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	// WebSocket hubs
	commentHub := commentws.NewCommentHub() // issue comments WS
	hub := ws.NewHub()               // kanban WS

	// -----------------------
	// Repositories
	// -----------------------
	projectRepo := pg.NewProjectRepository(db)
	userRepo := pg.NewUserRepo(db)
	issueRepo := pg.NewIssueRepository(db)
	commentRepo := pg.NewCommentRepository(db)
	activityRepo := pg.NewActivityRepository(db)
	projectMemberRepo := pg.NewProjectMemberRepo(db)
	kanbanRepo := kanbanrepo.NewKanbanRepo(db)
	labelRepo := pg.NewLabelRepository(db)
	notificationRepo := pg.NewNotificationRepoPG(db)

	// -----------------------
	// Services
	// -----------------------
	projectService := service.NewProjectService(projectRepo, activityRepo)
	userService := service.NewUserService(userRepo, projectRepo, projectMemberRepo)
	authService := service.NewAuthService(userRepo)
	activityService :=  service.NewActivityService(activityRepo);
	
	notifHub := notifications.NewNotificationHub()
	go notifHub.Run()
	notificationService := service.NewNotificationService(notificationRepo, userRepo, notifHub)

	issueService := service.NewIssueService(
		issueRepo, projectRepo, userRepo, commentRepo, activityRepo, activityService, commentHub, notificationService,
	)

	projectMemberService := service.NewProjectMemberService(projectRepo, userRepo, projectMemberRepo)
	kanbanService := service.NewKanbanService(issueRepo, projectRepo, projectMemberRepo, kanbanRepo)
	labelService := service.NewLabelService(labelRepo, projectRepo)
	

	handlers.RegisterNotificationHandlers(notificationService)

	// -----------------------
	// Controllers
	// -----------------------
	projectController := controllers.NewProjectController(projectService)
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService, userService)

	issueController := controllers.NewIssueController(issueService)
	issueCommentController := controllers.NewIssueCommentController(issueService)
	issueRelationController := controllers.NewIssueRelationController(issueService)
	issueAttachmentController := controllers.NewIssueAttachmentController(issueService)
	issueChecklistController := controllers.NewIssueChecklistController(issueService)
	issueSubtaskController := controllers.NewIssueSubtaskController(issueService)

	projectMemberController := controllers.NewProjectMemberController(projectMemberService)
	projectLabelController := controllers.NewProjectLabelController(labelService)

	notificationController := controllers.NewNotificationController(notificationService)


	// -----------------------
	// Fiber App
	// -----------------------
	app := fiber.New()

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api", cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
	}))

	// Public
	authPublic := api.Group("/auth")
	routes.AuthRoutes(authPublic, authController)

	// Auth Protected
	authProtected := api.Group("/auth")
	authProtected.Use(mw.JWTProtected())
	routes.AuthProtectedRoutes(authProtected, authController)

	// Protected
	protected := api.Use(mw.JWTProtected())

	routes.ProjectRoutes(protected, projectController, issueController, projectMemberController, projectLabelController)

	// Issue routes (with WS)
	routes.IssueRoutes(
		protected,
		issueController,
		issueCommentController,
		issueRelationController,
		issueAttachmentController,
		issueChecklistController,
		issueSubtaskController,
		commentHub,
	)

	routes.UserRoutes(protected, userController)

	// Kanban WS
	routes.RegisterKanbanRoutes(protected, kanbanService, hub)

	routes.NotificationRoutes(protected, notificationController)

	// Global WS routes
	wsGroup := app.Group("/ws")
	routes.RegisterWebSocketRoutes(wsGroup, hub)
	routes.RegisterIssueCommentWS(wsGroup, commentHub)
	routes.RegisterNotificationWSRoutes(wsGroup, notifHub)

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Println("BugForge server running on :8080")
	log.Fatal(app.Listen(":8080"))
}
