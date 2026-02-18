package postgres

import userdomain "admin.com/admin-api/internal/domain/user"

func ToDomainUser(model *DBUser) *userdomain.User {
	return &userdomain.User{
		ID:           model.ID,
		Name:         model.Name,
		LastName:     model.LastName,
		Username:     model.Username,
		PasswordHash: model.PasswordHash,
		Email:        model.Email,
		Avatar:       model.Avatar,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func ToDomainUsers(models []DBUser) []userdomain.User {
	users := make([]userdomain.User, len(models))
	for i := range models {
		users[i] = *ToDomainUser(&models[i])
	}

	return users
}

func FromDomainUser(user *userdomain.User) *DBUser {
	return &DBUser{
		ID:           user.ID,
		Name:         user.Name,
		LastName:     user.LastName,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		Avatar:       user.Avatar,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func SyncDomainUserFromModel(dst *userdomain.User, src *DBUser) {
	dst.ID = src.ID
	dst.Name = src.Name
	dst.LastName = src.LastName
	dst.Username = src.Username
	dst.PasswordHash = src.PasswordHash
	dst.Email = src.Email
	dst.Avatar = src.Avatar
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}
