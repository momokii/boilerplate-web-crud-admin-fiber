package middleware

import (
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/internal/repository"
	"fiber-prjct-management-web/pkg/database"
	"fiber-prjct-management-web/pkg/utils"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt/v5"
)

var Store *session.Store

func InitStore() {
	Store = session.New(session.Config{
		Expiration:     7 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
	})
}

func CreateSession(c *fiber.Ctx, key string, value interface{}) error {
	session, err := Store.Get(c)
	if err != nil {
		return err
	}
	defer session.Save()

	session.Set(key, value)

	return nil
}

func DeleteSession(c *fiber.Ctx) error {
	session, err := Store.Get(c)
	if err != nil {
		return err
	}

	session.Destroy()
	c.ClearCookie()

	return nil
}

func CheckSession(c *fiber.Ctx, key string) (interface{}, error) {
	session, err := Store.Get(c)
	if err != nil {
		return nil, err
	}

	return session.Get(key), nil
}

func AuthSessCheckerView(c *fiber.Ctx) error {
	checkUser, err := CheckSession(c, "token")
	if err != nil {
		return err
	}

	if checkUser != nil {
		return c.Redirect("/")
	}

	return nil
}

func ValidateAndGetUserData(c *fiber.Ctx, token string) (models.UserSession, error) {
	userSess := models.UserSession{}

	decode_token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return userSess, err
	}
	userId := decode_token.Claims.(jwt.MapClaims)["id"].(float64)

	// user data for local communication
	tx, err := database.DB.Begin()
	if err != nil {
		return userSess, err
	}
	defer utils.CommitOrRollback(tx, c)

	userRepo := repository.NewUserRepository(database.DB)
	userData, err := userRepo.FindByID(tx, int(userId))
	if err != nil {
		return userSess, err
	}

	if userData.IsDeleted {
		return userSess, fmt.Errorf("user not found")
	}

	userSession := models.UserSession{
		Id:       userData.Id,
		Username: userData.Username,
		Role:     userData.Role,
	}

	return userSession, nil
}

func IsAuthAPI(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if header == "" {
		return utils.ErrorJSON(c, fiber.StatusUnauthorized, "Need token header")
	}

	headerSplit := strings.Split(header, "Bearer ")
	if len(headerSplit) != 2 {
		return utils.ErrorJSON(c, fiber.StatusUnauthorized, "Invalid token")
	}

	token := headerSplit[1]

	// token validation
	userSession, err := ValidateAndGetUserData(c, token)
	if err != nil {
		return utils.ErrorJSON(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	c.Locals("user", userSession)
	c.Locals("typereq", utils.APIRequest)

	return c.Next()
}

func IsAuthWeb(c *fiber.Ctx) error {
	token, err := CheckSession(c, "token")
	if err != nil {
		DeleteSession(c)
		return c.Redirect("/login")
	}

	if token == nil {
		DeleteSession(c)
		return c.Redirect("/login")
	}

	// token validation
	userSession, err := ValidateAndGetUserData(c, token.(string))
	if err != nil {
		DeleteSession(c)
		return c.Redirect("/login")
	}

	// store information for next data
	c.Locals("user", userSession)
	c.Locals("typereq", utils.WebRequest)

	return c.Next()
}
