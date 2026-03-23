package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToUserResponse(user *entity.User) *model.UserResponse {
	return &model.UserResponse{
		ID: user.UserId,
		Username: user.Username,
		Email: user.Email,
		Role: user.Role,
		CreatedAt: user.CreatedAt,
	}
}

func ToUsersResponse(users []*entity.User) []*model.UserResponse {
	usersResponse := make([]*model.UserResponse, len(users))
	for i, user := range users {
		usersResponse[i] = ToUserResponse(user)
	}
	return usersResponse
}
