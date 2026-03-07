package user

type UpdateUserInfoRequest struct {
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
}
