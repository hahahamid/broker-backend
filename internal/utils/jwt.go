package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateTokens(userID, secret, refreshSecret string, accessExpMin int) (accessToken, refreshToken string, err error) {
	// Access token
	atClaims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(accessExpMin) * time.Minute).Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err = at.SignedString([]byte(secret))
	if err != nil {
		return
	}

	// Refresh token
	rtClaims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err = rt.SignedString([]byte(refreshSecret))
	return
}

func ValidateToken(tokenStr, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
}
