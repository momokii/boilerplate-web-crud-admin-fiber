package handlers

import (
	"database/sql"
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type DailyLogHandler struct {
	projectRepo  repository.ProjectRepository
	dailyLogRepo repository.DailyLogRepository
}

func NewDailyLogHandler(projectRepo repository.ProjectRepository, dailyLogRepo repository.DailyLogRepository) *DailyLogHandler {
	return &DailyLogHandler{
		projectRepo,
		dailyLogRepo,
	}
}

func (h *DailyLogHandler) ViewProjectDetail(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	return c.Render("pages/projectDetail", fiber.Map{
		"Title": "Project Detail",
		"User":  user,
		"Breadcrumb": models.BreadCrumb{
			BeforeName: "Project",
			BeforeLink: "/project",
		},
	})
}

func (h *DailyLogHandler) GetDailyLogsData(c *fiber.Ctx) error {
	// user := c.Locals("user").(models.UserSession)

	projectId, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	per_page, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	search := c.Query("search", "")
	fromDate := c.Query("from_date", "")
	toDate := c.Query("to_date", "")

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// first check if project owner
	_, err = h.projectRepo.FindByID(tx, projectId)
	if err != nil {
		// error handling this can happen if project not found or user is not project owner
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusNotFound, "Project not found")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// get all data
	logs, total, err := h.dailyLogRepo.FindWithPagination(tx, per_page, page, search, projectId, fromDate, toDate, 0, 0)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondWithPagination(c, fiber.StatusOK, "Get Daily Logs Data", total, page, per_page, "logs", logs)
}

func (h *DailyLogHandler) GetProjectLogStats(c *fiber.Ctx) error {
	// user := c.Locals("user").(models.UserSession)

	projectID, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check if projectowner
	if _, err = h.projectRepo.FindByID(tx, projectID); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project not found/ User is not project owner")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// project stats
	projectStats, err := h.dailyLogRepo.FindStats(tx, projectID)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// project stats cumulative
	projectStatsCum, err := h.dailyLogRepo.FindStatsCumulative(tx, projectID)
	if err != nil {
		fmt.Println("sini error", err)
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Get Project Log Stats", fiber.Map{
		"projectStats":    projectStats,
		"projectStatsCum": projectStatsCum,
	})
}

func (h *DailyLogHandler) GetOneLogData(c *fiber.Ctx) error {
	// user := c.Locals("user").(models.UserSession)
	projectID, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	logId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid log ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check if project owner
	_, err = h.projectRepo.FindByID(tx, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project not found/ User is not project owner")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// get log data
	log, err := h.dailyLogRepo.FindByID(tx, logId)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// check if log is in same project
	if log.ProjectId != projectID {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log id not found on this project")
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Get One Log Data", log)
}

func (h *DailyLogHandler) CreateDailyLog(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)
	projectID, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	income, err := strconv.Atoi(c.FormValue("income"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Income must be a number")
	}

	expense, err := strconv.Atoi(c.FormValue("expense"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Expense must be a number")
	}

	logInput := models.DailyLogInput{
		ProjectId:   projectID,
		LogDate:     c.FormValue("log_date"),
		Description: c.FormValue("description"),
		Issues:      c.FormValue("issues"),
		Income:      income,
		Expense:     expense,
		File:        "",
	}

	err = utils.ValidateStruct(logInput)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "ProjectId":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project ID is required")
			case "LogDate":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log Date is required")
			case "Income":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Income is required")
			case "Expense":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Expense is required")
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check if today log already exist
	checkLogToday, err := h.dailyLogRepo.FindByDate(tx, logInput.LogDate, projectID)
	if (err != nil) && (err != sql.ErrNoRows) {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	if checkLogToday.Id != 0 {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Daily log with selected date already exist")
	}

	// check if user is project owner
	_, err = h.projectRepo.FindIfProjectOwner(tx, projectID, user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project not found")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	logInput.ProjectId = projectID

	// upload file if uploaded
	file, err := c.FormFile("file")
	if (err != nil) && (file != nil) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if file != nil {
		filename, err := utils.GenerateNameLogsFiles(logInput.LogDate, logInput.ProjectId)
		if err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}

		filepath, err := utils.FileUpload(c, "logs/", filename)
		if err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}

		logInput.File = filepath
	}

	if err = h.dailyLogRepo.Create(tx, &logInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Create Daily Log")
}

func (h *DailyLogHandler) UpdateDailyLog(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	projectID, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	logId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid log ID")
	}

	income, err := strconv.Atoi(c.FormValue("income"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Income must be a number")
	}

	expense, err := strconv.Atoi(c.FormValue("expense"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Expense must be a number")
	}

	logUpdateInput := models.DailyLogInput{
		LogDate:     c.FormValue("log_date"),
		Description: c.FormValue("description"),
		Issues:      c.FormValue("issues"),
		ProjectId:   projectID,
		Income:      income,
		Expense:     expense,
	}

	err = utils.ValidateStruct(logUpdateInput)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "ProjectId":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Project ID is required")
			case "LogDate":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log Date is required")
			case "Income":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Income is required")
			case "Expense":
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check if project owner
	if _, err := h.dailyLogRepo.FindIfProjectAndLogOwner(tx, projectID, logId, user.Id); err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log data on project not found/ User is not log owner")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// check if log date already exist on another log data
	checkLogToday, err := h.dailyLogRepo.FindByDate(tx, logUpdateInput.LogDate, projectID)
	if (err != nil) && (err != sql.ErrNoRows) {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	if (checkLogToday.Id != logId) && (checkLogToday.Id != 0) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Daily log with selected date already exist")
	}

	// process file upload if user upload new file
	file, err := c.FormFile("file")
	if (err != nil) && (file != nil) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if file != nil {
		// delete old file if exist
		if checkLogToday.File.String != "" {
			if err := utils.DeleteFile(checkLogToday.File.String); err != nil {
				return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		// upload file
		filename, err := utils.GenerateNameLogsFiles(logUpdateInput.LogDate, logUpdateInput.ProjectId)
		if err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}

		filepath, err := utils.FileUpload(c, "logs/", filename)
		if err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}

		logUpdateInput.File = filepath
	} else {
		logUpdateInput.File = checkLogToday.File.String
	}

	// update log data
	if err = h.dailyLogRepo.Update(tx, &logUpdateInput, logId); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Update Daily Log")
}

func (h *DailyLogHandler) DeleteLog(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)
	projectID, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	logId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid log ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check if user is project owner
	log, err := h.dailyLogRepo.FindIfProjectAndLogOwner(tx, projectID, logId, user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log data on project not found/ User is not log owner")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// delete file first if the file is exist
	if log.File.String != "" {
		if err = utils.DeleteFile(log.File.String); err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	// delete log
	if err = h.dailyLogRepo.Delete(tx, logId); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Delete Daily Log")
}

func (h *DailyLogHandler) DeleteFileLog(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserSession)

	project_id, err := strconv.Atoi(c.Params("project_id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid project ID")
	}

	log_id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid log ID")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// find if log exist and the log is the owner of the project
	log, err := h.dailyLogRepo.FindIfProjectAndLogOwner(tx, project_id, log_id, user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "Log data on project not found/ User is not log owner")
		}

		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// delete file first if there is a file
	if log.File.String != "" {
		err = utils.DeleteFile(log.File.String)
		if err != nil {
			return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	// update file path to empty the filename
	log_update := models.DailyLogInput{
		File:        "",
		ProjectId:   project_id,
		LogDate:     log.LogDate,
		Description: log.Description,
		Issues:      log.Issues,
		Income:      log.Income,
		Expense:     log.Expense,
	}
	if err = h.dailyLogRepo.Update(tx, &log_update, log_id); err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Delete File Log")
}
