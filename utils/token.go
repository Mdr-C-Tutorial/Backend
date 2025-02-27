package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
	"mdr/config"
)

func GenerateToken(userID string, remember bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(GetTokenDuration(remember)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["user_id"].(string), nil
	}

	return "", jwt.ErrSignatureInvalid
}

func GetTokenDuration(remember bool) time.Duration {
	if remember {
		return 30 * 24 * time.Hour // 30天
	}
	return 6 * time.Hour
}

func GetCookieMaxAge(remember bool) int {
	if remember {
		return 30 * 24 * 60 * 60 // 30天（秒）
	}
	return 6 * 60 * 60 // 6小时（秒）
}

func GenerateEmailVerificationToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"purpose": "email_verification",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func VerifyEmailToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if purpose, ok := claims["purpose"].(string); !ok || purpose != "email_verification" {
			return "", jwt.ErrSignatureInvalid
		}
		return claims["user_id"].(string), nil
	}

	return "", jwt.ErrSignatureInvalid
}