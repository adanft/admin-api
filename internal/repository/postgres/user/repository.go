package postgres

import (
	"context"
	"strings"

	"admin.com/admin-api/internal/domain"
	userdomain "admin.com/admin-api/internal/domain/user"
	pgroot "admin.com/admin-api/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	dbConn *bun.DB
}

func NewUserRepository(dbConn *bun.DB) *UserRepository {
	return &UserRepository{
		dbConn: dbConn,
	}
}

func (repo *UserRepository) GetUser(ctx context.Context, id uuid.UUID) (*userdomain.User, error) {
	model := new(DBUser)
	if err := repo.dbConn.NewSelect().Model(model).Where("id = ?", id).Limit(1).Scan(ctx); err != nil {
		return nil, pgroot.MapSelectError(err)
	}

	return ToDomainUser(model), nil
}

func (repo *UserRepository) CreateUser(ctx context.Context, user *userdomain.User) error {
	model := FromDomainUser(user)

	if _, err := repo.dbConn.NewInsert().Model(model).Exec(ctx); err != nil {
		return pgroot.MapPersistenceWriteError(err, pgroot.MapUserIdentityUniqueConstraint)
	}

	SyncDomainUserFromModel(user, model)
	return nil
}

func (repo *UserRepository) GetUsers(ctx context.Context) ([]userdomain.User, error) {
	var users []DBUser

	err := repo.dbConn.NewSelect().Model(&users).Scan(ctx)
	if err != nil {
		return []userdomain.User{}, pgroot.WrapInternal(err)
	}

	return ToDomainUsers(users), nil
}

func (repo *UserRepository) UpdateUser(ctx context.Context, user *userdomain.User) error {
	model := FromDomainUser(user)

	res, err := repo.dbConn.NewUpdate().
		Model(model).
		Column("name", "last_name", "username", "email", "avatar").
		Where("id = ?", user.ID).
		Exec(ctx)

	if err != nil {
		return pgroot.MapPersistenceWriteError(err, pgroot.MapUserIdentityUniqueConstraint)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return pgroot.WrapInternal(err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user := &DBUser{ID: id}

	res, err := repo.dbConn.NewDelete().Model(user).WherePK().Exec(ctx)

	if err != nil {
		return pgroot.WrapInternal(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return pgroot.WrapInternal(err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func GetUserByID(ctx context.Context, dbConn bun.IDB, id uuid.UUID) (*userdomain.User, error) {
	model := new(DBUser)
	if err := dbConn.NewSelect().Model(model).Where("id = ?", id).Limit(1).Scan(ctx); err != nil {
		return nil, pgroot.MapSelectError(err)
	}

	return ToDomainUser(model), nil
}

func GetUserByIdentity(ctx context.Context, dbConn bun.IDB, identity string) (*userdomain.User, error) {
	model := new(DBUser)

	identity = strings.TrimSpace(identity)
	query := dbConn.NewSelect().Model(model).Limit(1)
	if strings.Contains(identity, "@") {
		query = query.Where("lower(email) = lower(?)", identity)
	} else {
		query = query.Where("username = ?", identity)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, pgroot.MapSelectError(err)
	}

	return ToDomainUser(model), nil
}
