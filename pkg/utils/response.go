package utils

import (
	"github.com/gofiber/fiber/v2"
)

func RespondMessage(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   false,
		"message": message,
	})
}

func RespondWithData(c *fiber.Ctx, code int, message string, data interface{}) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   false,
		"message": message,
		"data":    data,
	})
}

func RespondWithPagination(c *fiber.Ctx, code int, message string, total int, page int, perPage int, dataName string, data interface{}) error {
	return c.Status(code).JSON(fiber.Map{
		"error":   false,
		"message": message,
		"data": fiber.Map{
			dataName:  data,
			"total":   total,
			"page":    page,
			"perPage": perPage,
		},
		"recordsFiltered": total,
		"recordsTotal":    total,
	})
}
