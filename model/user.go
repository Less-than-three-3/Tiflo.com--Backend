package model

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserCtx = "UserId"
)

type User struct {
	UserId   uuid.UUID `json:"userId"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
}

type UserLogin struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func GetUserId(context *gin.Context) (uuid.UUID, error) {
	userIdStr, exists := context.Get(UserCtx)
	if !exists || userIdStr == "" {
		return uuid.Nil, errors.New("no user id in context")
	}

	return uuid.Parse(userIdStr.(string))
}
