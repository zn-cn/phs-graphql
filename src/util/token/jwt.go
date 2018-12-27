package token

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func GetJWTToken(data map[string]interface{}, secret string, expire time.Duration) (token string) {
	t := jwt.New(jwt.SigningMethodHS256)
	claims := t.Claims.(jwt.MapClaims)
	for key, value := range data {
		claims[key] = value
	}
	claims["exp"] = time.Now().Add(expire).Unix()
	token, _ = t.SignedString([]byte(secret))
	return
}

func ValidateJWT(authScheme, token, secret string) (jwt.MapClaims, bool) {
	if authScheme != token[:len(authScheme)] {
		return nil, false
	}
	t, _ := jwt.Parse(token[len(authScheme)+1:], func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, t.Valid
	}
	return claims, t.Valid
}
