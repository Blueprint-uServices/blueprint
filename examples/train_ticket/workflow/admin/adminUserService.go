package admin

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/user"
)

type AdminUserService interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
	AddUser(ctx context.Context, user user.User) error
	UpdateUser(ctx context.Context, user user.User) (bool, error)
	DeleteUser(ctx context.Context, userID string) error
}

type AdminUserServiceImpl struct {
	userService user.UserService
}

func NewAdminUserServiceImpl(ctx context.Context, userService user.UserService) (*AdminUserServiceImpl, error) {
	return &AdminUserServiceImpl{userService}, nil
}

func (a *AdminUserServiceImpl) GetAllUsers(ctx context.Context) ([]user.User, error) {
	return a.userService.GetAllUsers(ctx)
}

func (a *AdminUserServiceImpl) AddUser(ctx context.Context, u user.User) error {
	return a.userService.SaveUser(ctx, u)
}

func (a *AdminUserServiceImpl) UpdateUser(ctx context.Context, u user.User) (bool, error) {
	return a.userService.UpdateUser(ctx, u)
}

func (a *AdminUserServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	return a.userService.DeleteUser(ctx, userID)
}
