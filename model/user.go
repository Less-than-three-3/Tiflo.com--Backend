package model

import (
	"errors"
	"fmt"
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
	userId := context.GetString(UserCtx)
	fmt.Println(userId)

	fmt.Println(context)

	if userId == "" {
		return uuid.Nil, errors.New("no user id in context")
	}

	userIdStr := fmt.Sprintf("%v", userId)
	return uuid.Parse(userIdStr)
}
