package logic

import (
	"context"

	"github.com/liukunxin/go-infra/pkg/base/log"
	"single-starter/internal/app/user/convert"
	"single-starter/internal/app/user/dto"
	"single-starter/internal/app/user/service"
	"single-starter/internal/app/user/vo"
	"single-starter/internal/infra/notifier"
)

type UserLogic interface {
	// CreateUser 编排创建用户流程并返回视图对象。
	CreateUser(ctx context.Context, in *dto.CreateUserInput) (*vo.UserDetail, error)
	// GetUser 编排按ID查询用户流程并返回视图对象。
	GetUser(ctx context.Context, id string) (*vo.UserDetail, error)
}

type userLogic struct {
	userSvc  service.UserService
	notifier notifier.Notifier
}

func NewUserLogic(userSvc service.UserService, userNotifier notifier.Notifier) UserLogic {
	return &userLogic{userSvc: userSvc, notifier: userNotifier}
}

func (l *userLogic) CreateUser(ctx context.Context, in *dto.CreateUserInput) (*vo.UserDetail, error) {
	user, err := l.userSvc.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	// 编排点：落库成功后再发通知。通知失败不回滚创建结果，只记录日志。
	if err = l.notifier.NotifyUserCreated(ctx, user.ID); err != nil {
		log.WithContext(ctx).Errorf("notify user created failed: uid=%s err=%v", user.ID, err)
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
