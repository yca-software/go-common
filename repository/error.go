package repository

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	error_helpers "github.com/yca-software/go-common/error"
)

func WrapSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return error_helpers.NewNotFoundError(err, "", nil)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return error_helpers.NewConflictError(err, "", map[string]any{
				"constraint_name": pgErr.ConstraintName,
			})
		case "23503":
			return error_helpers.NewUnprocessableEntityError(err, "", nil)
		}
	}

	return err
}

func ErrNotFoundNoRowsAffected() error {
	return error_helpers.NewNotFoundError(errors.New("no rows affected"), "", nil)
}
