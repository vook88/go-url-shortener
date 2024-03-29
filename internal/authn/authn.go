package authn

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

var ErrUserIDNotFound = errors.New("userID not found in token")
var ErrTokenIsNotValid = errors.New("token is not valid")

const TokenExp = time.Hour * 24
const SecretKey = "supersecretkey"

func BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, ErrTokenIsNotValid
	}

	if claims.UserID == 0 {
		return -1, ErrUserIDNotFound
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}
