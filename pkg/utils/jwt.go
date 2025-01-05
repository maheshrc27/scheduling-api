package utils

import (
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maheshrc27/postflow/internal/transfer"
)

func GenerateToken(secretKey, userID string, tokenDuration time.Duration) (string, error) {
	log.Println("generate user id", userID)
	claims := transfer.CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "postflow",
		},
	}

	log.Printf("%v", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))

	if err != nil {
		slog.Info(err.Error())
		return "", err
	}

	return signedToken, nil
}

func ValidateToken(secretKey, tokenString string) (*transfer.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &transfer.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		slog.Info(err.Error())
		return nil, err
	}

	if claims, ok := token.Claims.(*transfer.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
