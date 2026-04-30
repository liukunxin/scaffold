package convert

import (
	"go-infra-starter/internal/app/user/dto"
	"go-infra-starter/internal/app/user/model"
	"go-infra-starter/internal/app/user/ro"
	"go-infra-starter/internal/app/user/vo"
)

func CreateReqToDTO(req *ro.CreateUserReq) *dto.CreateUserInput {
	return &dto.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	}
}

func ModelToVO(user *model.User) *vo.UserDetail {
	return &vo.UserDetail{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

