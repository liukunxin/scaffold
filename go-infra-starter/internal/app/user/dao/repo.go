package dao

import (
	"context"
	"fmt"
	"sync"

	"go-infra-starter/internal/app/user/model"
)

type UserRepo interface {
	// Create 持久化用户实体。
	Create(ctx context.Context, user *model.User) error
	// GetByID 按ID查询用户实体。
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type memoryUserRepo struct {
	mu   sync.RWMutex
	data map[string]*model.User
}

func NewUserRepo() UserRepo {
	return &memoryUserRepo{
		data: make(map[string]*model.User),
	}
}

func (r *memoryUserRepo) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[user.ID] = user
	return nil
}

func (r *memoryUserRepo) GetByID(_ context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return user, nil
}
