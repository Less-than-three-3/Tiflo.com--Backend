package model

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	JwtPrefix = "Bearer "
)

type JwtClaims struct {
	jwt.StandardClaims
	UserId uuid.UUID `json:"userId"`
}
