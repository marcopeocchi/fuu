package domain

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

const (
	Standard int = iota
	Admin
)

type UserEnt struct {
	ID       uint
	Username string
	Password string
	Role     int
}

type User struct {
	gorm.Model
	ID       uint
	Username string `gorm:"unique;not null"`
	Password string
	Salt     string
	Role     int
}

type UserRepository interface {
	FindById(ctx context.Context, id uint) (User, error)

	FindByUsername(ctx context.Context, username string) (User, error)

	Create(ctx context.Context, username, password string, role int) (User, error)

	Update(ctx context.Context, id uint, username, password string, role int) (User, error)

	Delete(ctx context.Context, id uint) (User, error)
}

type UserService interface {
	Login(ctx context.Context, username, password string) (*string, error)

	Create(ctx context.Context, username, password string, role int) (User, error)
}

type UserHandler interface {
	Login() http.HandlerFunc

	Logout() http.HandlerFunc
}
