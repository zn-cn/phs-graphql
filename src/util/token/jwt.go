package token

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
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

func ValidateJWT(token, secret string) (jwt.MapClaims, bool) {
	t, _ := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, t.Valid
	}
	return claims, t.Valid
}

// old
// GetJWTInfo 获取 payload
func GetJWTInfo(c echo.Context) jwt.MapClaims {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims
}
