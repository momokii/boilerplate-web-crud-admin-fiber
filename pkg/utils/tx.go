package utils

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func CommitOrRollback(tx *sql.Tx, c *fiber.Ctx, err error) {
	if p := recover(); p != nil {
		_ = tx.Rollback()
		panic(p)
	} else if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			err = fmt.Errorf("error rolling back transaction: %v (original error : %w)", rbErr, err)
		}
	} else {
		cerr := tx.Commit()
		if cerr != nil {
			err = fmt.Errorf("error committing transaction: %v (original error : %w)", cerr, err)
		}
	}
}
