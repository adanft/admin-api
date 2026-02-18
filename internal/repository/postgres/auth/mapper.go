package postgres

import domainauth "admin.com/admin-api/internal/domain/auth"

func toDomainRefreshToken(model *DBRefreshToken) *domainauth.RefreshToken {
	return &domainauth.RefreshToken{
		ID:         model.ID,
		UserID:     model.UserID,
		FamilyID:   model.FamilyID,
		TokenHash:  model.TokenHash,
		ExpiresAt:  model.ExpiresAt,
		RevokedAt:  model.RevokedAt,
		LastUsedAt: model.LastUsedAt,
		CreatedAt:  model.CreatedAt,
	}
}

func fromDomainRefreshToken(model *domainauth.RefreshToken) *DBRefreshToken {
	return &DBRefreshToken{
		ID:         model.ID,
		UserID:     model.UserID,
		FamilyID:   model.FamilyID,
		TokenHash:  model.TokenHash,
		ExpiresAt:  model.ExpiresAt,
		RevokedAt:  model.RevokedAt,
		LastUsedAt: model.LastUsedAt,
		CreatedAt:  model.CreatedAt,
	}
}

func syncDomainRefreshTokenFromModel(dst *domainauth.RefreshToken, src *DBRefreshToken) {
	dst.ID = src.ID
	dst.UserID = src.UserID
	dst.FamilyID = src.FamilyID
	dst.TokenHash = src.TokenHash
	dst.ExpiresAt = src.ExpiresAt
	dst.RevokedAt = src.RevokedAt
	dst.LastUsedAt = src.LastUsedAt
	dst.CreatedAt = src.CreatedAt
}
