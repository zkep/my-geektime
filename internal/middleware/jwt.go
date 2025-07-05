package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/zkep/my-geektime/internal/global"
	"github.com/zkep/my-geektime/libs/rest"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := ParseToken(global.JWT, c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
			return
		}
		if err = doWithToken(c, token); err != nil {
			return
		}
		c.Next()
	}
}

func doWithToken(c *gin.Context, token *jwt.Token) error {
	claims := make(map[string]any)
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
		return ErrExpiredToken
	}
	for key, value := range mapClaims {
		claims[key] = value
	}
	if claims["exp"] == nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
		return ErrExpiredToken
	}
	if _, ok = claims["exp"].(float64); !ok {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
		return ErrExpiredToken
	}
	exp, exists := claims["exp"].(float64)
	if !exists {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
		return ErrExpiredToken
	}
	if int64(exp) < time.Now().UTC().Unix() {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"status": http.StatusBadRequest, "msg": "token is expired"})
		return ErrExpiredToken
	}
	c.Set(global.Identity, claims[global.Identity])
	c.Set(global.Role, claims[global.Role])
	return nil
}

func ParseToken(j *rest.JWTConfig, c *gin.Context) (*jwt.Token, error) {
	var token string
	var err error

	methods := strings.Split(j.TokenLookup, ",")
	for _, method := range methods {
		if len(token) > 0 {
			break
		}
		parts := strings.Split(strings.TrimSpace(method), ":")
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch k {
		case "header":
			token, err = jwtFromHeader(c, v)
		case "query":
			token, err = jwtFromQuery(c, v)
		case "cookie":
			token, err = jwtFromCookie(c, v)
		case "param":
			token, err = jwtFromParam(c, v)
		}
	}

	if err != nil {
		return nil, err
	}
	return j.ParseToken(token)
}

var (
	ErrExpiredToken      = errors.New("token is expired")
	ErrEmptyAuthHeader   = errors.New("auth header is empty")
	ErrInvalidAuthHeader = errors.New("auth header is invalid")
	ErrEmptyQueryToken   = errors.New("query token is empty")
	ErrEmptyCookieToken  = errors.New("cookie token is empty")
	ErrEmptyParamToken   = errors.New("parameter token is empty")
)

func jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)

	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", ErrInvalidAuthHeader
	}
	return parts[1], nil
}

func jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)

	if token == "" {
		return "", ErrEmptyQueryToken
	}

	return token, nil
}

func jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)

	if cookie == "" {
		return "", ErrEmptyCookieToken
	}

	return cookie, nil
}

func jwtFromParam(c *gin.Context, key string) (string, error) {
	token := c.Param(key)

	if token == "" {
		return "", ErrEmptyParamToken
	}

	return token, nil
}
