package utils

import (
	"chat-server/global"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// AdminTokenClaims 管理端JWT Claims（独立定义，避免循环依赖）
type AdminTokenClaims struct {
	UserID      string `json:"user_id"`
	UserAccount string `json:"user_account"`
	RoleId      int16  `json:"role_id"`
	jwt.RegisteredClaims
}

// GenerateAdminToken 生成管理端 token（7天有效期）
func GenerateAdminToken(userID, userAccount string, roleId int16) (string, error) {
	claims := &AdminTokenClaims{
		UserID:      userID,
		UserAccount: userAccount,
		RoleId:      roleId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    global.CHAT_CONFIG.JWT.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(global.CHAT_CONFIG.JWT.Secret))
}
