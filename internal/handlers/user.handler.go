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
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo}
}

func (h *UserHandler) ViewUser(c *fiber.Ctx) error {
	// get local user session data
	user := c.Locals("user").(models.UserSession)

	return c.Render("pages/user", fiber.Map{
		"Title": "User Data",
		"User":  user,
		"Breadcrumb": models.BreadCrumb{
			BeforeName: "Dashboard",
			BeforeLink: "/",
		},
	})
}

func (h *UserHandler) ViewUserSelf(c *fiber.Ctx) error {
	// get local user session data
	user := c.Locals("user").(models.UserSession)

	return c.Render("pages/userDetail", fiber.Map{
		"Title": "User Self Data",
		"User":  user,
		"Breadcrumb": models.BreadCrumb{
			BeforeName: "Dashboard",
			BeforeLink: "/",
		},
	})
}

func (h *UserHandler) GetUsersData(c *fiber.Ctx) error {
	perPage, err := strconv.Atoi(c.Query("per_page", "10"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid page value")
	}

	searchQuery := c.Query("search", "")
	role := c.Query("role", "")
	fromDate := c.Query("from_date", "")
	toDate := c.Query("to_date", "")

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	users, total, err := h.userRepo.FindWithPagination(tx, perPage, page, searchQuery, role, toDate, fromDate)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.RespondWithPagination(c, fiber.StatusOK, "Get Users Data", total, page, perPage, "users", users)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid id")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	check_user, err := h.userRepo.FindByID(tx, id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if check_user.Id == 0 {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "User not found")
	}

	return utils.RespondWithData(c, fiber.StatusOK, "Get User By ID", check_user)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	userInput := new(models.CreateUserInput)
	err := c.BodyParser(userInput)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	err = utils.ValidateStruct(userInput)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Username":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username minimal 5 karakter, merupakan alphanumerik")
			case "Password":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Password minimal 6 karakter, mengandung angka dan huruf besar")
			default:
				return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	checkUser, err := h.userRepo.FindByUsername(tx, userInput.Username)
	if (err != nil) && (err != sql.ErrNoRows) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if checkUser.Username != "" {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username already exists")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), 16)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	userInput.Password = string(hashedPass)

	err = h.userRepo.Create(tx, userInput)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "User created successfully")
}

func (h *UserHandler) EditUserPassword(c *fiber.Ctx) error {
	userInput := new(models.UpdatePasswordInput)

	if err := c.BodyParser(userInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if err := utils.ValidateStruct(userInput); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Id":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Id is required")
			case "OldPassword":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Old Password is required")
			case "Password":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Password minimal 6 karakter, mengandung angka dan huruf besar")
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// not check if user exist or not here bcs it's already checked in middleware using IsSelf checker
	// but still check user for get old pass data
	userData, err := h.userRepo.FindByID(tx, userInput.Id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	// check old password
	if err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(userInput.OldPassword)); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Old password is incorrect")
	}

	// hash password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), 16)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	userInput.Password = string(hashedPass)

	if err = h.userRepo.UpdatePassword(tx, userInput); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Edit User Password")
}

func (h *UserHandler) EditUser(c *fiber.Ctx) error {
	userInput := new(models.UpdateUserInput)

	userId, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid id")
	}

	err = c.BodyParser(userInput)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	err = utils.ValidateStruct(userInput)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Username":
				return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username minimal 5 karakter, merupakan alphanumerik")
			default:
				return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// check data user update
	userUpdate, err := h.userRepo.FindByID(tx, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrorJSON(c, fiber.StatusBadRequest, "User data not found")
		}

		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	// check username if exist
	checkUsername, err := h.userRepo.FindByUsername(tx, userInput.Username)
	if (err != nil) && (err != sql.ErrNoRows) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if (checkUsername.Username != "") && (checkUsername.Id != userId) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username already exists")
	}

	userUpdate.Username = userInput.Username
	userUpdate.Role = userInput.Role

	err = h.userRepo.Update(tx, &userUpdate)
	if err != nil {
		fmt.Println("error sini")
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Edit User")
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Invalid id")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	check_user, err := h.userRepo.FindByID(tx, id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if check_user.Id == 0 {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "User not found")
	}

	err = h.userRepo.SoftDelete(tx, id)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.RespondMessage(c, fiber.StatusOK, "Delete User")
}
