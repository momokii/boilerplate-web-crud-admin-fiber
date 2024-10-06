package utils

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func CommitOrRollback(tx *sql.Tx, c *fiber.Ctx) {
	if r := recover(); r != nil {
		_ = tx.Rollback()
		ErrorJSON(c, fiber.StatusInternalServerError, "Internal Server Error")
	} else {
		_ = tx.Commit()
	}
}
