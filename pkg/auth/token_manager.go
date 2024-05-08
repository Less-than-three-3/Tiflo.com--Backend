package auth

import (
	"errors"
	"time"

	"tiflo/model"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// TokenManager provides logic for JWT tokens generation and parsing.
type TokenManager interface {
	NewJWT(userId uuid.UUID) (string, error)
	Parse(accessToken string) (uuid.UUID, error)
}

type Manager struct {
	signingKey string
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewJWT(userId uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &model.JwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId: userId,
	})

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) Parse(accessToken string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(accessToken, &model.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.signingKey), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	myClaims := token.Claims.(*model.JwtClaims)
	return myClaims.UserId, nil
}
