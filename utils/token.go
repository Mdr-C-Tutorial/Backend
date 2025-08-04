package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"mdr/config"
)

func GenerateToken(userID string, remember bool) (string, error) {
	expirationTime := time.Now().Add(GetTokenDuration(remember))
	fmt.Printf("Generating token for user %s with expiration time: %v (remember: %t)\n", userID, expirationTime, remember)
	
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
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
	var duration time.Duration
	if remember {
		duration = time.Duration(config.AppConfig.JWT.RememberMeTTL) * time.Second // 30天
		fmt.Printf("GetTokenDuration (remember=true): %v\n", duration)
	} else {
		duration = time.Duration(config.AppConfig.JWT.TTL) * time.Second // 6小时
		fmt.Printf("GetTokenDuration (remember=false): %v\n", duration)
	}
	return duration
}

func GetCookieMaxAge(remember bool) int {
	if remember {
		return int(config.AppConfig.JWT.RememberMeTTL) // 30天（秒）
	}
	return int(config.AppConfig.JWT.TTL) // 6小时（秒）
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