package user

type UpdateUserProfileRequest struct {
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
}
