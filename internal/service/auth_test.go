package service

import (
	"context"
	"testing"

	"github.com/Wild-sergunys/flowmodel/internal/model"
)

type mockUserRepo struct {
	users map[string]*model.User
}

func (m *mockUserRepo) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	user, ok := m.users[login]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func TestAuthService_Login(t *testing.T) {
	// пароль admin123 для admin
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"admin": {
				ID:           1,
				Login:        "admin",
				PasswordHash: "$2a$10$dbGDmlZf4wi74l3FfTJyU./jGCVXliu59pyYmXbmCrTXXQuPz9meu",
				Role:         "admin",
			},
		},
	}

	svc := NewAuthService(repo, "test-secret-key")

	tests := []struct {
		name     string
		login    string
		password string
		wantOk   bool
		wantRole string
	}{
		{
			name:     "успешный вход",
			login:    "admin",
			password: "admin123",
			wantOk:   true,
			wantRole: "admin",
		},
		{
			name:     "неверный пароль",
			login:    "admin",
			password: "wrong",
			wantOk:   false,
		},
		{
			name:     "пользователь не найден",
			login:    "notfuser",
			password: "anything",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, token, err := svc.Login(context.Background(), tt.login, tt.password)

			if tt.wantOk {
				if err != nil {
					t.Fatalf("неожиданная ошибка: %v", err)
				}
				if user == nil {
					t.Fatal("ожидался пользователь, получен nil")
				}
				if user.Role != tt.wantRole {
					t.Errorf("роль = %v, want %v", user.Role, tt.wantRole)
				}
				if token == "" {
					t.Error("токен не должен быть пустым")
				}
			} else {
				if err == nil {
					t.Fatal("ожидалась ошибка, получен nil")
				}
				if err != ErrInvalidCredentials {
					t.Errorf("ошибка = %v, want %v", err, ErrInvalidCredentials)
				}
			}
		})
	}
}
