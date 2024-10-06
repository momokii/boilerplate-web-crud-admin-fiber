package handlers

import (
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	projectRepo  repository.ProjectRepository
	dailyLogRepo repository.DailyLogRepository
}

func NewDashboardHandler(projectRepo repository.ProjectRepository, dailyLogRepo repository.DailyLogRepository) *DashboardHandler {
	return &DashboardHandler{
		projectRepo,
		dailyLogRepo,
	}
}

func (h *DashboardHandler) ViewDashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	return c.Render("pages/dashboard", fiber.Map{
		"Title": "Dashboard",
		"User":  user,
		"Breadcrumb": models.BreadCrumb{
			BeforeName: "Dashboard",
			BeforeLink: "/",
		},
	})
}

func (h *DashboardHandler) DashboardData(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	if user.Role == 3 { // super admin
		user.Id = 0
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	project_data, err := h.projectRepo.FindProjectsStats(tx, user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	project_status_data, err := h.projectRepo.FindProjectStatusStats(tx, user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	newest_created_projects, _, err := h.projectRepo.FindWithPagination(tx, 5, 1, "", "", "", "", user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	newest_daily_logs, _, err := h.dailyLogRepo.FindWithPagination(tx, 10, 1, "", 0, "", "", user.Id, user.Role)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Dashboard Data", fiber.Map{
		"project_data":            project_data,
		"project_status_data":     project_status_data,
		"newest_created_projects": newest_created_projects,
		"newest_daily_logs":       newest_daily_logs,
	})
}
