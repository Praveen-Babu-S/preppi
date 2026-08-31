package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
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
			return nil, fmt.Errorf("read public key: %w", err)
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
		if err != nil {
			return nil, fmt.Errorf("parse public key: %w", err)
		}
		m.publicKey = pubKey
	}

	if privKeyPath != "" {
		privBytes, err := os.ReadFile(privKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}
		privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
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
		return "", fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, nil
}

func (m *Manager) ValidateToken(tokenStr string) (*Claims, error) {
	if m.publicKey == nil {
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
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func ExtractToken(ctx context.Context) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMissingToken
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, ErrMissingToken
	}

	return nil, ErrMissingToken
}

func ExtractTokenFromString(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return header
}
