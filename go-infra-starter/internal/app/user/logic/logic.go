package logic

import (
	"context"

	"go-infra-starter/internal/app/user/convert"
	"go-infra-starter/internal/app/user/dto"
	"go-infra-starter/internal/app/user/service"
	"go-infra-starter/internal/app/user/vo"
)

type UserLogic interface {
	CreateUser(ctx context.Context, in *dto.CreateUserInput) (*vo.UserDetail, error)
	GetUser(ctx context.Context, id string) (*vo.UserDetail, error)
}

type userLogic struct {
	userSvc service.UserService
}

func NewUserLogic(userSvc service.UserService) UserLogic {
	return &userLogic{userSvc: userSvc}
}

func (l *userLogic) CreateUser(ctx context.Context, in *dto.CreateUserInput) (*vo.UserDetail, error) {
	user, err := l.userSvc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return convert.ModelToVO(user), nil
}

func (l *userLogic) GetUser(ctx context.Context, id string) (*vo.UserDetail, error) {
	user, err := l.userSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convert.ModelToVO(user), nil
}

