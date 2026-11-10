package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtkey = []byte("your-secret-key-change-in-production")

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func JwtKey() []byte {
	return jwtkey
}

func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate((time.Now().Add(24 * time.Hour))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtkey)
}
