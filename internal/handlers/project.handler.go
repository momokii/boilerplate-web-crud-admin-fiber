package handlers

import (
	"database/sql"
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct {
	projectRepo  repository.ProjectRepository
	dailyLogRepo repository.DailyLogRepository
}

func NewProjectHandler(projectRepo repository.ProjectRepository, dailyLogRepo repository.DailyLogRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo,
		dailyLogRepo,
	}
}

func (h *ProjectHandler) ViewProject(c *fiber.Ctx) error {
	userData := c.Locals("user").(models.UserSession)

	return c.Render("pages/project", fiber.Map{
		"Title": "Project",
		"User":  userData,
		"Breadcrumb": models.BreadCrumb{
			BeforeName: "Dashboard",
			BeforeLink: "/",
		},
	})
}

func (h *ProjectHandler) GetProjectsData(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	perPage, err := strconv.Atoi(c.Query("per_page", "100"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	searchQuery := c.Query("search", "")
	status := c.Query("status", "")
	fromDate := c.Query("from_date", "")
	toDate := c.Query("to_date", "")

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// if superadmin can get all project from all user
	if user.Role == 3 {
		user.Id = 0
	}

	projects, total, err := h.projectRepo.FindWithPagination(tx, perPage, page, searchQuery, status, toDate, fromDate, user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondWithPagination(c, fiber.StatusOK, "Get Projects Data", total, page, perPage, "projects", projects)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	checkProjectOwner, err := h.projectRepo.FindByID(tx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusNotFound, "Project not found")
		}

		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if user.Role != 3 {
		if checkProjectOwner.CreatedBy != user.Id {
			return utils.ErrorJSON(c, fiber.StatusUnauthorized, "Unauthorized")
		}
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Get Project By ID", checkProjectOwner)
}

func (h *ProjectHandler) GetProjectsStats(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	// if user is superadmin can get all project from all user
	if user.Role == 3 {
		user.Id = 0
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// project stats by status
	projectStatusStats, err := h.projectRepo.FindProjectStatusStats(tx, user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// project stats general
	projectStats, err := h.projectRepo.FindProjectsStats(tx, user.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Get Projects Stats", fiber.Map{
		"projectStats":       projectStats,
		"projectStatusStats": projectStatusStats,
	})
}

func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {

	userData := c.Locals("user").(models.UserSession)

	projectInput := new(models.ProjectInput)
	if err := c.BodyParser(projectInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}
	if err := utils.ValidateStruct(projectInput); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Name is required minimal 5 characters and max 50 characters")
			case "Description":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Description is required minimal 5 characters and max 255 characters")
			case "Status":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Status is required")
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// input id admin created
	projectInput.CreatedBy = userData.Id

	if err := h.projectRepo.Create(tx, projectInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Create Project")
}

func (h *ProjectHandler) EditProject(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	projectId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid ID")
	}

	projectInput := new(models.ProjectInput)
	if err := c.BodyParser(projectInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if err := utils.ValidateStruct(projectInput); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Name is required minimal 5 characters and max 50 characters")
			case "Description":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Description is required minimal 5 characters and max 255 characters")
			case "Status":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Status is required")
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	checkProjectOwner, err := h.projectRepo.FindIfProjectOwner(tx, projectId, user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project not found")
		}

		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	projectInput.CreatedBy = checkProjectOwner.CreatedBy

	if err := h.projectRepo.Update(tx, projectInput, checkProjectOwner.Id); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Edit Project")
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	var log_files []string
	user := c.Locals("user").(models.UserSession)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	_, err = h.projectRepo.FindIfProjectOwner(tx, id, user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project not found")
		}

		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	// delete all daily logs related to project (is provided by database cascade), but files not, so delete all files related to project on daily logs
	// find all daily logs related to project to get file name if exist on the logs
	logs, _, err := h.dailyLogRepo.FindWithPagination(tx, 9999, 1, "", id, "", "", 0, 0)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	if len(logs) > 0 {
		// add all files name to array first bcs need confirm to delete project and logs first before delete files
		for _, log := range logs {
			if log.File.String != "" {
				log_files = append(log_files, log.File.String)
			}
		}
	}

	// delete project with all logs
	if err = h.projectRepo.Delete(tx, id); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// delete all files
	if len(log_files) > 0 {
		for _, file := range log_files {
			utils.DeleteFile(file)
		}
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Delete Project")
}
