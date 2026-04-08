package user

import "time"

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    *string   `json:"avatar,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Avatar   *string `json:"avatar"`
	Role     string  `json:"role"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password *string `json:"password,omitempty"`
	Avatar   *string `json:"avatar"`
	Role     string  `json:"role"`
}

func ToUserResponse(user *User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserResponseList(users []*User) []*UserResponse {
	if users == nil {
		return nil
	}

	res := make([]*UserResponse, len(users))
	for i, user := range users {
		resp := ToUserResponse(user)
		if resp != nil {
			res[i] = resp
		}
	}
	return res
}
