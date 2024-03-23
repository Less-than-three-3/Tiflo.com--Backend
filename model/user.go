package model

import "github.com/google/uuid"

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
