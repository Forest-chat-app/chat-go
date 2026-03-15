package user

type UpdateUserAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}
