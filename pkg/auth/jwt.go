package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrMissingToken = errors.New("missing token")
	ErrNoPublicKey  = errors.New("public key not configured")
	ErrNoPrivateKey = errors.New("private key not configured")
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(pubKeyPath, privKeyPath string, accessTTL, refreshTTL time.Duration) (*Manager, error) {
	m := &Manager{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}

	if pubKeyPath != "" {
		pubBytes, err := os.ReadFile(pubKeyPath)
		if err != nil {
			log.Info().Err(err).Msg("unable to read public key")
			return nil, err
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
		if err != nil {
			log.Info().Err(err).Msg("unable to parse public key")
			return nil, err
		}
		m.publicKey = pubKey
	}

	if privKeyPath != "" {
		privBytes, err := os.ReadFile(privKeyPath)
		if err != nil {
			log.Info().Err(err).Msg("unable to read private key")
			return nil, err
		}
		privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
		if err != nil {
			log.Info().Err(err).Msg("unable to parse private key")
			return nil, err
		}
		m.privateKey = privKey
	}

	return m, nil
}

func (m *Manager) GenerateAccessToken(userID, role string) (string, error) {
	if m.privateKey == nil {
		return "", ErrNoPrivateKey
	}

	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("unable to sign access token")
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (m *Manager) GenerateRefreshToken(userID string) (string, error) {
	if m.privateKey == nil {
		return "", ErrNoPrivateKey
	}

	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("unable to sign refresh token")
		return "", fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, nil
}

func (m *Manager) ValidateToken(tokenStr string) (*Claims, error) {
	if m.publicKey == nil {
		log.Warn().Msg("public key not configured, cannot validate token")
		return nil, ErrNoPublicKey
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			log.Info().Str("subject", claims.Subject).Msg("token expired")
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenUsedBeforeIssued), errors.Is(err, jwt.ErrTokenNotValidYet):
			log.Warn().Err(err).Str("subject", claims.Subject).Msg("token not yet valid")
			return nil, ErrInvalidToken
		default:
			log.Warn().Err(err).Msg("invalid token")
			return nil, ErrInvalidToken
		}
	}

	if !token.Valid {
		log.Warn().Msg("invalid token")
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ErrMissingToken is kept for callers that extract tokens before validation.
func ExtractTokenFromString(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return header
}
