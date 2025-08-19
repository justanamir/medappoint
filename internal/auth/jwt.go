package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func SignJWT(secret, issuer string, ttlMinutes int, uid int64, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: uid,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlMinutes) * time.Minute)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}
