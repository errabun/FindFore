package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	PlayerID     int64
	TokenVersion int32
}

func GenerateToken(playerID int64, tokenVersion int32, secret string) (string, error) {
	claims := jwt.MapClaims{
		"player_id":     playerID,
		"token_version": tokenVersion,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
		"iat":           time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return Claims{}, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	playerID, ok := claims["player_id"].(float64)
	if !ok {
		return Claims{}, fmt.Errorf("invalid player_id in token")
	}

	var tokenVersion int32
	if v, ok := claims["token_version"].(float64); ok {
		tokenVersion = int32(v)
	}

	return Claims{
		PlayerID:     int64(playerID),
		TokenVersion: tokenVersion,
	}, nil
}
