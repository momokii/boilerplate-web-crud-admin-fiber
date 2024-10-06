package middleware

import (
	"fiber-prjct-management-web/internal/models"
	"fiber-prjct-management-web/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func IsSuperAdmin(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		if user.Role != 3 {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}

func IsAdmin(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		if user.Role != 1 {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}

func IsSelf(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		// get id context from param
		userId, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusBadRequest, "Invalid user id")
		}

		if user.Id != userId {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}

func IsSuperAdminOrAdmin(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		if (user.Role != 1) && (user.Role != 3) {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}

func IsSuperAdminOrIsSelf(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		// get id context from param
		userId, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusBadRequest, "Invalid user id")
		}

		if (user.Id != userId) && (user.Role != 3) {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}

func IsAdminOrIsSelf(reqType utils.RequestType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.UserSession)

		// get id context from param
		userId, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusBadRequest, "Invalid user id")
		}

		if (user.Id != userId) && (user.Role != 1) {
			return utils.HandlerUnauthorizedResponse(c, reqType, fiber.StatusUnauthorized, "Unauthorized")
		}

		return c.Next()
	}
}
