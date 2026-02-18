package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainauth "admin.com/admin-api/internal/domain/auth"
	"github.com/google/uuid"
)

type Config struct {
	Secret    string
	Issuer    string
	Audience  string
	AccessTTL time.Duration
}

type JWT struct {
	secret    []byte
	issuer    string
	audience  string
	accessTTL time.Duration
	now       func() time.Time
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
	TokenID  string `json:"jti"`
}

func NewJWT(cfg Config) (*JWT, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, ErrJWTSecretRequired
	}
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, ErrJWTIssuerRequired
	}
	if strings.TrimSpace(cfg.Audience) == "" {
		return nil, ErrJWTAudienceNeeded
	}
	if cfg.AccessTTL <= 0 {
		return nil, ErrJWTAccessTTL
	}

	return &JWT{
		secret:    []byte(cfg.Secret),
		issuer:    cfg.Issuer,
		audience:  cfg.Audience,
		accessTTL: cfg.AccessTTL,
		now:       time.Now,
	}, nil
}

func (j *JWT) GenerateAccessToken(userID uuid.UUID) (string, time.Time, error) {
	now := j.now().UTC()
	expiresAt := now.Add(j.accessTTL)

	claims := jwtClaims{
		Subject:  userID.String(),
		Audience: j.audience,
		Issuer:   j.issuer,
		IssuedAt: now.Unix(),
		Expires:  expiresAt.Unix(),
		TokenID:  uuid.NewString(),
	}

	header := jwtHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	}

	token, err := j.sign(header, claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (j *JWT) ParseAccessToken(token string) (*domainauth.AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return nil, ErrInvalidToken
	}

	signedValue := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !j.verifySignature(signedValue, signature) {
		return nil, ErrInvalidToken
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if strings.TrimSpace(claims.Subject) == "" || strings.TrimSpace(claims.TokenID) == "" {
		return nil, ErrInvalidToken
	}
	if claims.Issuer != j.issuer || claims.Audience != j.audience {
		return nil, ErrInvalidToken
	}

	now := j.now().UTC().Unix()
	if claims.Expires <= now {
		return nil, ErrExpiredToken
	}
	if claims.IssuedAt > now+60 {
		return nil, ErrInvalidToken
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, ErrInvalidToken
	}

	return &domainauth.AccessTokenClaims{
		Subject:   claims.Subject,
		Audience:  claims.Audience,
		Issuer:    claims.Issuer,
		IssuedAt:  time.Unix(claims.IssuedAt, 0).UTC(),
		ExpiresAt: time.Unix(claims.Expires, 0).UTC(),
		TokenID:   claims.TokenID,
	}, nil
}

func (j *JWT) sign(header jwtHeader, claims jwtClaims) (string, error) {
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signedValue := headerEncoded + "." + claimsEncoded

	signature := j.signValue(signedValue)
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	return signedValue + "." + signatureEncoded, nil
}

func (j *JWT) signValue(value string) []byte {
	mac := hmac.New(sha256.New, j.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func (j *JWT) verifySignature(value string, provided []byte) bool {
	expected := j.signValue(value)
	return hmac.Equal(expected, provided)
}
