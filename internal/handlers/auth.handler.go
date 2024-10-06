package handlers

import (
	"database/sql"
	"fiber-prjct-management-web/internal/middleware"
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo repository.UserRepository
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo}
}

func (h *AuthHandler) LoginView(c *fiber.Ctx) error {
	// if stil have session redirect to dashboard
	if err := middleware.AuthSessCheckerView(c); err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Render("pages/login", fiber.Map{
		"Title": "Login",
	})
}

func (h *AuthHandler) LoginWeb(c *fiber.Ctx) error {

	loginInput := new(models.Login)
	err := c.BodyParser(loginInput)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	err = utils.ValidateStruct(loginInput)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Masukan username dan password")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	defer utils.CommitOrRollback(tx, c)

	// username checking
	userLogin, err := h.userRepo.FindByUsername(tx, loginInput.Username)
	if (err != nil) && (err != sql.ErrNoRows) {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, err.Error())
	}

	if userLogin.Id == 0 || userLogin.IsDeleted {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username/Password salah")
	}

	// password checking
	err = bcrypt.CompareHashAndPassword([]byte(userLogin.Password), []byte(loginInput.Password))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusBadRequest, "Username/Password salah")
	}

	// create cookie jwt
	sign := jwt.New(jwt.SigningMethodHS256)
	claims := sign.Claims.(jwt.MapClaims)
	claims["id"] = userLogin.Id
	claims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	token, err := sign.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	// create session
	middleware.CreateSession(c, "token", token)
	// also save token to cookie
	c.Cookie(&fiber.Cookie{
		Name:  "token",
		Value: token,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login",
		"error":   false,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	middleware.DeleteSession(c)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logout",
		"error":   false,
	})
}
