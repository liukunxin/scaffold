package convert

import (
	"single-starter/internal/app/user/dto"
	"single-starter/internal/app/user/model"
	"single-starter/internal/app/user/ro"
	"single-starter/internal/app/user/vo"
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
