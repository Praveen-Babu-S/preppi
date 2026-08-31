package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// tempKeyFiles writes an RSA keypair to temp files and returns their paths.
func tempKeyFiles(t *testing.T) (pubPath, privPath string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	dir := t.TempDir()
	privPath = filepath.Join(dir, "test_rsa")
	pubPath = filepath.Join(dir, "test_rsa.pub")

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubASN1})
	if err := os.WriteFile(pubPath, pubPEM, 0600); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return pubPath, privPath
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	pub, priv := tempKeyFiles(t)
	m, err := NewManager(pub, priv, 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	m := newTestManager(t)

	userID := "42"
	role := "student"

	token, err := m.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := m.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %q, got %q", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("expected role %q, got %q", role, claims.Role)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	m := newTestManager(t)

	token, err := m.GenerateRefreshToken("99")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	claims, err := m.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "99" {
		t.Errorf("expected userID 99, got %q", claims.UserID)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	m := newTestManager(t)

	_, err := m.ValidateToken("not-a-real-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestMissingPrivateKeyRejected(t *testing.T) {
	pub, _ := tempKeyFiles(t)
	// Only public key path provided, no private key — GenerateAccessToken must fail.
	m, err := NewManager(pub, "", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if _, err := m.GenerateAccessToken("1", "student"); err == nil {
		t.Fatal("expected error when private key missing, got nil")
	}
}

func TestExtractTokenFromString(t *testing.T) {
	cases := map[string]string{
		"Bearer abc123": "abc123",
		"abc123":        "abc123",
		"":              "",
	}
	for in, want := range cases {
		if got := ExtractTokenFromString(in); got != want {
			t.Errorf("ExtractTokenFromString(%q) = %q, want %q", in, got, want)
		}
	}
}
