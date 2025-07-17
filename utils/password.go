package utils

import (
	"chat-server/global"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

func GenerateFromPassword(password string) (string, error) {
	logger := global.CHAT_LOG
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("utils-->GenerateFromPassword-->加密密码出错", "err", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func CompareHashAndPassword(oldPassword, newPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(newPassword))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
