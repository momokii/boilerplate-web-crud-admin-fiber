package main

import (
	"fiber-prjct-management-web/internal/handlers"
	"fiber-prjct-management-web/internal/middleware"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	database.ConnectDB()
	middleware.InitStore()
	// repo init
	userRepo := repository.NewUserRepository(database.DB)
	projectRepo := repository.NewProjectRepository(database.DB)
	dailyLogRepo := repository.NewDailyLogRepository(database.DB)

	// handler init
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, dailyLogRepo)
	dailyLogHandler := handlers.NewDailyLogHandler(projectRepo, dailyLogRepo)
	dashboardHandler := handlers.NewDashboardHandler(projectRepo, dailyLogRepo)

	// engine := html.New("./web", ".html")
	engine := html.New("./web", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// if req from api endpoint return json error
			url := c.Request().URI().String()
			urlSplit := strings.Split(url, "/")

			if urlSplit[3] == string(utils.APIRequest) {
				return utils.ErrorJSON(c, code, err.Error())
			}

			// if req from web endpoint return html error
			return c.Status(code).Render("pages/errorPage", fiber.Map{
				"Title": "Error",
				"Error": err.Error(),
				"Code":  code,
			})
		},
	})
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Static("/web", "./web")

	// routing

	// routing group
	api := app.Group("/api")

	// dashboard
	app.Get("/", middleware.IsAuthWeb, dashboardHandler.ViewDashboard)
	api.Get("/dashboard", middleware.IsAuthAPI, dashboardHandler.DashboardData)

	// project
	app.Get("/project", middleware.IsAuthWeb, middleware.IsSuperAdminOrAdmin(utils.WebRequest), projectHandler.ViewProject)
	api.Get("/projects", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), projectHandler.GetProjectsData)
	api.Get("/projects/stats", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), projectHandler.GetProjectsStats)
	api.Get("/projects/:id", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), projectHandler.GetProjectByID)
	api.Post("/projects", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), projectHandler.CreateProject)
	api.Patch("/projects/:id", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), projectHandler.EditProject)
	api.Delete("/projects/:id", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), projectHandler.DeleteProject)

	// project detail/ logs data
	app.Get("/project/:id", middleware.IsAuthWeb, middleware.IsSuperAdminOrAdmin(utils.WebRequest), dailyLogHandler.ViewProjectDetail)
	api.Get("projects/:project_id/stats", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), dailyLogHandler.GetProjectLogStats)
	api.Get("/projects/:project_id/logs", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), dailyLogHandler.GetDailyLogsData)
	api.Get("/projects/:project_id/logs/:id", middleware.IsAuthAPI, middleware.IsSuperAdminOrAdmin(utils.APIRequest), dailyLogHandler.GetOneLogData)
	api.Post("/projects/:project_id/logs", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), dailyLogHandler.CreateDailyLog)
	api.Patch("/projects/:project_id/logs/:id", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), dailyLogHandler.UpdateDailyLog)
	api.Delete("/projects/:project_id/logs/:id", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), dailyLogHandler.DeleteLog)
	api.Delete("/projects/:project_id/logs/:id/files", middleware.IsAuthAPI, middleware.IsAdmin(utils.APIRequest), dailyLogHandler.DeleteFileLog)

	app.Get("/user", middleware.IsAuthWeb, middleware.IsSuperAdmin(utils.WebRequest), userHandler.ViewUser)
	app.Get("/user/self", middleware.IsAuthWeb, userHandler.ViewUserSelf)
	api.Get("/users", middleware.IsAuthAPI, middleware.IsSuperAdmin(utils.APIRequest), userHandler.GetUsersData)
	api.Get("/users/:id", middleware.IsAuthAPI, middleware.IsSuperAdminOrIsSelf(utils.APIRequest), userHandler.GetUserByID)
	api.Patch("/users/:id", middleware.IsAuthAPI, middleware.IsSuperAdminOrIsSelf(utils.APIRequest), userHandler.EditUser)
	api.Patch("/users/:id/password", middleware.IsAuthAPI, middleware.IsSelf(utils.APIRequest), userHandler.EditUserPassword)
	api.Post("/users", middleware.IsAuthAPI, middleware.IsSuperAdmin(utils.APIRequest), userHandler.CreateUser)
	api.Delete("/users/:id", middleware.IsAuthAPI, middleware.IsSuperAdmin(utils.APIRequest), userHandler.DeleteUser)

	app.Get("/login", authHandler.LoginView)
	api.Post("/login", authHandler.LoginWeb)
	api.Post("/logout", middleware.IsAuthAPI, authHandler.Logout)

	app.Listen(":3000")
}
