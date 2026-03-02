package utils

import "github.com/google/uuid"

func GenerateUUid() string {
	return uuid.New().String()
}
