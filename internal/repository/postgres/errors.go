package postgres

import (
	"database/sql"
	"errors"

	"admin.com/admin-api/internal/domain"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	pgErrCodeUniqueViolation           = "23505"
	pgErrCodeForeignKeyViolation       = "23503"
	pgErrCodeNotNullViolation          = "23502"
	pgErrCodeCheckViolation            = "23514"
	pgErrCodeStringDataRightTruncation = "22001"
)

type UniqueConstraintMapper func(constraintName string) error

func MapPersistenceWriteError(err error, mapUniqueConstraint UniqueConstraintMapper) error {
	var pgErr pgdriver.Error
	if !errors.As(err, &pgErr) {
		return WrapInternal(err)
	}

	switch pgErr.Field('C') {
	case pgErrCodeUniqueViolation:
		if mapUniqueConstraint == nil {
			return domain.ErrConflict
		}

		return mapUniqueConstraint(pgErr.Field('n'))
	case pgErrCodeForeignKeyViolation, pgErrCodeNotNullViolation, pgErrCodeCheckViolation, pgErrCodeStringDataRightTruncation:
		return domain.ErrBadRequest
	default:
		return WrapInternal(err)
	}
}

func WrapInternal(err error) error {
	if err == nil {
		return domain.ErrInternalServerError
	}

	return errors.Join(domain.ErrInternalServerError, err)
}

func MapSelectError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	return WrapInternal(err)
}

func MapUserIdentityUniqueConstraint(constraintName string) error {
	switch constraintName {
	case "users_username_key":
		return errors.Join(domain.ErrConflict, domain.ErrUsernameExists)
	case "users_email_key", "users_email_lower_uidx":
		return errors.Join(domain.ErrConflict, domain.ErrEmailExists)
	default:
		return domain.ErrConflict
	}
}
