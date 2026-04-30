package service

import (
	"context"
	"strconv"
	"time"

	"github.com/liukunxin/go-infra/pkg/base/uuid"
	"go-infra-starter/internal/app/user/dao"
	"go-infra-starter/internal/app/user/dto"
	"go-infra-starter/internal/app/user/model"
)

type UserService interface {
	Create(ctx context.Context, in *dto.CreateUserInput) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type userService struct {
	repo dao.UserRepo
}

func NewUserService(repo dao.UserRepo) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, in *dto.CreateUserInput) (*model.User, error) {
	now := time.Now()
	user := &model.User{
		ID:        strconv.FormatInt(uuid.GetIDService().GenerateUserID(), 10),
		Name:      in.Name,
		Email:     in.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

