package postgres

import (
	"context"
	"time"

	"admin.com/admin-api/internal/domain"
	domainauth "admin.com/admin-api/internal/domain/auth"
	userdomain "admin.com/admin-api/internal/domain/user"
	pgroot "admin.com/admin-api/internal/repository/postgres"
	userpostgres "admin.com/admin-api/internal/repository/postgres/user"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuthRepository struct {
	dbConn *bun.DB
}

func NewAuthRepository(dbConn *bun.DB) *AuthRepository {
	return &AuthRepository{dbConn: dbConn}
}

func (repo *AuthRepository) CreateUser(ctx context.Context, user *userdomain.User) error {
	model := userpostgres.FromDomainUser(user)
	if _, err := repo.dbConn.NewInsert().Model(model).Exec(ctx); err != nil {
		return pgroot.MapPersistenceWriteError(err, mapAuthUniqueConstraint)
	}

	userpostgres.SyncDomainUserFromModel(user, model)
	return nil
}

func (repo *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error) {
	return userpostgres.GetUserByID(ctx, repo.dbConn, id)
}

func (repo *AuthRepository) GetUserByIdentity(ctx context.Context, identity string) (*userdomain.User, error) {
	return userpostgres.GetUserByIdentity(ctx, repo.dbConn, identity)
}

func (repo *AuthRepository) CreateRefreshToken(ctx context.Context, token *domainauth.RefreshToken) error {
	model := fromDomainRefreshToken(token)
	if _, err := repo.dbConn.NewInsert().Model(model).Exec(ctx); err != nil {
		return pgroot.MapPersistenceWriteError(err, mapAuthUniqueConstraint)
	}

	syncDomainRefreshTokenFromModel(token, model)
	return nil
}

func (repo *AuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domainauth.RefreshToken, error) {
	model := new(DBRefreshToken)
	if err := repo.dbConn.NewSelect().Model(model).Where("token_hash = ?", tokenHash).Limit(1).Scan(ctx); err != nil {
		return nil, pgroot.MapSelectError(err)
	}

	return toDomainRefreshToken(model), nil
}

func (repo *AuthRepository) RotateRefreshToken(ctx context.Context, currentTokenID uuid.UUID, nextToken *domainauth.RefreshToken, usedAt time.Time) error {
	nextTokenModel := fromDomainRefreshToken(nextToken)

	err := repo.dbConn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewUpdate().
			Model((*DBRefreshToken)(nil)).
			Set("revoked_at = ?", usedAt).
			Set("last_used_at = ?", usedAt).
			Where("id = ?", currentTokenID).
			Where("revoked_at IS NULL").
			Exec(ctx)
		if err != nil {
			return pgroot.WrapInternal(err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return pgroot.WrapInternal(err)
		}

		if rows == 0 {
			return domain.ErrConflict
		}

		if _, err := tx.NewInsert().Model(nextTokenModel).Exec(ctx); err != nil {
			return pgroot.MapPersistenceWriteError(err, mapAuthUniqueConstraint)
		}

		return nil
	})
	if err != nil {
		return err
	}

	syncDomainRefreshTokenFromModel(nextToken, nextTokenModel)
	return nil
}

func (repo *AuthRepository) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	_, err := repo.dbConn.NewUpdate().
		Model((*DBRefreshToken)(nil)).
		Set("revoked_at = ?", revokedAt).
		Set("last_used_at = ?", revokedAt).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL").
		Exec(ctx)
	if err != nil {
		return pgroot.WrapInternal(err)
	}

	return nil
}

func mapAuthUniqueConstraint(constraintName string) error {
	switch constraintName {
	case "auth_refresh_tokens_token_hash_uidx":
		return domain.ErrConflict
	default:
		return pgroot.MapUserIdentityUniqueConstraint(constraintName)
	}
}
