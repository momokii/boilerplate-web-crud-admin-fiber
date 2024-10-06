package utils

import (
	"github.com/gofiber/fiber/v2"
)

type RequestType string // ResponseType is a type for response type

const (
	APIRequest RequestType = "api"
	WebRequest RequestType = "web"
)

func HandlerUnauthorizedResponse(c *fiber.Ctx, responseType RequestType, code int, message string) error {
	if responseType == APIRequest {
		return ErrorJSON(c, code, message)
	}

	return c.Redirect("/")
}

func ErrorJSON(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}
