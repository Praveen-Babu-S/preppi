package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"preppi.com/pkg/auth"
	"preppi.com/services/auth/repository"
)

type fakeRepo struct {
	mu     sync.Mutex
	users  map[string]*repository.User // keyed by email
	nextID uint
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string]*repository.User{}}
}

func (f *fakeRepo) Create(ctx context.Context, u *repository.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	u.ID = f.nextID
	f.users[u.Email] = u
	return nil
}

func (f *fakeRepo) FindByEmail(ctx context.Context, email string) (*repository.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.users[email]; ok {
		return u, nil
	}
	return nil, errors.New("record not found")
}

func (f *fakeRepo) FindByID(ctx context.Context, id uint) (*repository.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("record not found")
}

func (f *fakeRepo) Update(ctx context.Context, u *repository.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[u.Email] = u
	return nil
}

func (f *fakeRepo) MarkEmailVerified(ctx context.Context, id uint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.ID == id {
			u.EmailVerified = true
		}
	}
	return nil
}

// newTestService returns an AuthService with a fake repo and an RSA-backed token manager.
func newTestService(t *testing.T) *AuthService {
	t.Helper()
	repo := newFakeRepo()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	dir := t.TempDir()
	privPath := filepath.Join(dir, "test_rsa")
	pubPath := filepath.Join(dir, "test_rsa.pub")

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	_ = os.WriteFile(privPath, privPEM, 0600)
	pubASN1, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubASN1})
	_ = os.WriteFile(pubPath, pubPEM, 0600)

	tokenMgr, err := auth.NewManager(pubPath, privPath, 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	return New(repo, tokenMgr)
}

func TestRegisterStoresHashedPassword(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	id, email, err := svc.Register(ctx, "Alice", "alice@example.com", "secret123", "student", "")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero user id")
	}
	if email != "alice@example.com" {
		t.Errorf("expected email, got %q", email)
	}

	u, err := svc.repo.FindByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if u.PasswordHash == "secret123" {
		t.Error("password stored in plaintext! must be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("secret123")); err != nil {
		t.Errorf("stored hash does not verify: %v", err)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "A", "dup@example.com", "pass123", "student", ""); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if _, _, err := svc.Register(ctx, "B", "dup@example.com", "pass456", "student", ""); err == nil {
		t.Fatal("expected ErrUserExists for duplicate email, got nil")
	} else if !errors.Is(err, ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "Bob", "bob@example.com", "correct-horse", "mentor", "math"); err != nil {
		t.Fatalf("register: %v", err)
	}

	access, refresh, id, role, err := svc.Login(ctx, "bob@example.com", "correct-horse")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if access == "" || refresh == "" {
		t.Error("expected non-empty tokens")
	}
	if id == 0 {
		t.Error("expected non-zero id")
	}
	if role != "mentor" {
		t.Errorf("expected role mentor, got %q", role)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, _, err := svc.Register(ctx, "Bob", "bob@example.com", "correct", "student", ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, _, _, _, err := svc.Login(ctx, "bob@example.com", "wrong-password"); err == nil {
		t.Fatal("expected ErrInvalidCreds for wrong password, got nil")
	}
}
